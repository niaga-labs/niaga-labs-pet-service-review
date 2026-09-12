package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/kafka"
	reviewDomain "github.com/niaga-labs/niaga-labs-pet-service-review/internal/domain/review"
	"go.uber.org/zap"
)

// CreateReviewRequest holds data to create a review.
type CreateReviewRequest struct {
	BookingID    uuid.UUID `json:"booking_id" binding:"required"`
	RevieweeID   uuid.UUID `json:"reviewee_id" binding:"required"`
	RevieweeType string    `json:"reviewee_type" binding:"required"`
	Rating       int       `json:"rating" binding:"required,min=1,max=5"`
	Comment      string    `json:"comment"`
	PhotoURLs    []string  `json:"photo_urls"`
}

// ReviewDTO is the API response representation of a review.
type ReviewDTO struct {
	ID           uuid.UUID `json:"id"`
	BookingID    uuid.UUID `json:"booking_id"`
	ReviewerID   uuid.UUID `json:"reviewer_id"`
	RevieweeID   uuid.UUID `json:"reviewee_id"`
	RevieweeType string    `json:"reviewee_type"`
	Rating       int       `json:"rating"`
	Comment      string    `json:"comment"`
	PhotoURLs    []string  `json:"photo_urls"`
	CreatedAt    time.Time `json:"created_at"`
}

// RatingSummaryDTO is the rating summary for a reviewee.
type RatingSummaryDTO struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int64   `json:"total_reviews"`
}

// ReviewService handles review use cases.
type ReviewService struct {
	repo     reviewDomain.ReviewRepository
	producer *kafka.Producer
	logger   *zap.Logger
}

// NewReviewService creates a new ReviewService.
func NewReviewService(repo reviewDomain.ReviewRepository, producer *kafka.Producer, logger *zap.Logger) *ReviewService {
	return &ReviewService{repo: repo, producer: producer, logger: logger}
}

// CreateReview creates a new review for a booking.
func (s *ReviewService) CreateReview(ctx context.Context, reviewerID uuid.UUID, req CreateReviewRequest) (*ReviewDTO, error) {
	// Check if already reviewed
	exists, err := s.repo.ExistsForBooking(ctx, req.BookingID, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("you have already reviewed this booking")
	}

	rev, err := reviewDomain.NewReview(
		req.BookingID,
		reviewerID,
		req.RevieweeID,
		reviewDomain.RevieweeType(req.RevieweeType),
		req.Rating,
		req.Comment,
		req.PhotoURLs,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, rev); err != nil {
		return nil, fmt.Errorf("failed to save review: %w", err)
	}

	// Publish review.submitted event
	s.publishReviewSubmitted(ctx, rev)

	s.logger.Info("review created",
		zap.String("id", rev.ID().String()),
		zap.String("booking_id", req.BookingID.String()),
		zap.Int("rating", req.Rating),
	)

	return toReviewDTO(rev), nil
}

// GetReviewsByReviewee returns paginated reviews for a runner/shop/owner.
func (s *ReviewService) GetReviewsByReviewee(ctx context.Context, revieweeID uuid.UUID, revieweeType string, page, limit int) ([]*ReviewDTO, int64, error) {
	offset := (page - 1) * limit
	reviews, total, err := s.repo.FindByRevieweeID(ctx, revieweeID, reviewDomain.RevieweeType(revieweeType), limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*ReviewDTO, len(reviews))
	for i, r := range reviews {
		dtos[i] = toReviewDTO(r)
	}
	return dtos, total, nil
}

// GetMyReviews returns paginated reviews written by the current user.
func (s *ReviewService) GetMyReviews(ctx context.Context, reviewerID uuid.UUID, page, limit int) ([]*ReviewDTO, int64, error) {
	offset := (page - 1) * limit
	reviews, total, err := s.repo.FindByReviewerID(ctx, reviewerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*ReviewDTO, len(reviews))
	for i, r := range reviews {
		dtos[i] = toReviewDTO(r)
	}
	return dtos, total, nil
}

// GetRatingSummary returns the average rating and total review count.
func (s *ReviewService) GetRatingSummary(ctx context.Context, revieweeID uuid.UUID, revieweeType string) (*RatingSummaryDTO, error) {
	avg, count, err := s.repo.AverageRating(ctx, revieweeID, reviewDomain.RevieweeType(revieweeType))
	if err != nil {
		return nil, err
	}
	return &RatingSummaryDTO{
		AverageRating: avg,
		TotalReviews:  count,
	}, nil
}

func (s *ReviewService) publishReviewSubmitted(ctx context.Context, rev *reviewDomain.Review) {
	evt := map[string]interface{}{
		"review_id":     rev.ID().String(),
		"booking_id":    rev.BookingID().String(),
		"reviewer_id":   rev.ReviewerID().String(),
		"reviewee_id":   rev.RevieweeID().String(),
		"reviewee_type": string(rev.RevieweeType()),
		"rating":        rev.Rating(),
		"occurred_at":   time.Now().UTC(),
	}

	cloudEvent, err := kafka.NewCloudEvent("service-review", "review.submitted", evt)
	if err != nil {
		s.logger.Error("failed to create cloud event", zap.Error(err))
		return
	}

	if err := s.producer.PublishEvent(ctx, "review-events", cloudEvent); err != nil {
		s.logger.Error("failed to publish review.submitted event", zap.Error(err))
	}
}

func toReviewDTO(r *reviewDomain.Review) *ReviewDTO {
	return &ReviewDTO{
		ID:           r.ID(),
		BookingID:    r.BookingID(),
		ReviewerID:   r.ReviewerID(),
		RevieweeID:   r.RevieweeID(),
		RevieweeType: string(r.RevieweeType()),
		Rating:       r.Rating(),
		Comment:      r.Comment(),
		PhotoURLs:    r.PhotoURLs(),
		CreatedAt:    r.CreatedAt(),
	}
}
