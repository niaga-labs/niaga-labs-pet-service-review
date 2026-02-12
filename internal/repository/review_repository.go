package repository

import (
	"context"
	"encoding/json"
	"time"

	reviewDomain "github.com/Kilat-Pet-Delivery/service-review/internal/domain/review"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReviewModel is the GORM model for the reviews table.
type ReviewModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	BookingID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ReviewerID   uuid.UUID `gorm:"type:uuid;not null;index"`
	RevieweeID   uuid.UUID `gorm:"type:uuid;not null;index:idx_reviewee"`
	RevieweeType string    `gorm:"type:varchar(20);not null;index:idx_reviewee"`
	Rating       int       `gorm:"not null"`
	Comment      string    `gorm:"type:text"`
	PhotoURLs    string    `gorm:"type:jsonb;default:'[]'"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

// TableName sets the table name.
func (ReviewModel) TableName() string { return "reviews" }

// GormReviewRepository implements ReviewRepository using GORM.
type GormReviewRepository struct {
	db *gorm.DB
}

// NewGormReviewRepository creates a new GormReviewRepository.
func NewGormReviewRepository(db *gorm.DB) *GormReviewRepository {
	return &GormReviewRepository{db: db}
}

// Save persists a new review.
func (r *GormReviewRepository) Save(ctx context.Context, rev *reviewDomain.Review) error {
	model := toReviewModel(rev)
	return r.db.WithContext(ctx).Create(&model).Error
}

// FindByID returns a review by ID.
func (r *GormReviewRepository) FindByID(ctx context.Context, id uuid.UUID) (*reviewDomain.Review, error) {
	var model ReviewModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return toReviewDomain(&model), nil
}

// FindByBookingID returns all reviews for a booking.
func (r *GormReviewRepository) FindByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*reviewDomain.Review, error) {
	var models []ReviewModel
	if err := r.db.WithContext(ctx).Where("booking_id = ?", bookingID).Find(&models).Error; err != nil {
		return nil, err
	}
	return toReviewDomainSlice(models), nil
}

// FindByRevieweeID returns paginated reviews for a specific reviewee.
func (r *GormReviewRepository) FindByRevieweeID(ctx context.Context, revieweeID uuid.UUID, revieweeType reviewDomain.RevieweeType, limit, offset int) ([]*reviewDomain.Review, int64, error) {
	var models []ReviewModel
	var total int64

	query := r.db.WithContext(ctx).Where("reviewee_id = ? AND reviewee_type = ?", revieweeID, string(revieweeType))
	query.Model(&ReviewModel{}).Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return toReviewDomainSlice(models), total, nil
}

// FindByReviewerID returns paginated reviews written by a user.
func (r *GormReviewRepository) FindByReviewerID(ctx context.Context, reviewerID uuid.UUID, limit, offset int) ([]*reviewDomain.Review, int64, error) {
	var models []ReviewModel
	var total int64

	query := r.db.WithContext(ctx).Where("reviewer_id = ?", reviewerID)
	query.Model(&ReviewModel{}).Count(&total)

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return toReviewDomainSlice(models), total, nil
}

// AverageRating returns the average rating and count for a reviewee.
func (r *GormReviewRepository) AverageRating(ctx context.Context, revieweeID uuid.UUID, revieweeType reviewDomain.RevieweeType) (float64, int64, error) {
	var result struct {
		Avg   float64
		Count int64
	}
	err := r.db.WithContext(ctx).
		Model(&ReviewModel{}).
		Select("COALESCE(AVG(rating), 0) as avg, COUNT(*) as count").
		Where("reviewee_id = ? AND reviewee_type = ?", revieweeID, string(revieweeType)).
		Scan(&result).Error
	return result.Avg, result.Count, err
}

// ExistsForBooking checks if a review already exists for a booking from a reviewer.
func (r *GormReviewRepository) ExistsForBooking(ctx context.Context, bookingID, reviewerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReviewModel{}).
		Where("booking_id = ? AND reviewer_id = ?", bookingID, reviewerID).
		Count(&count).Error
	return count > 0, err
}

func toReviewModel(r *reviewDomain.Review) ReviewModel {
	photoJSON, _ := json.Marshal(r.PhotoURLs())
	return ReviewModel{
		ID:           r.ID(),
		BookingID:    r.BookingID(),
		ReviewerID:   r.ReviewerID(),
		RevieweeID:   r.RevieweeID(),
		RevieweeType: string(r.RevieweeType()),
		Rating:       r.Rating(),
		Comment:      r.Comment(),
		PhotoURLs:    string(photoJSON),
		CreatedAt:    r.CreatedAt(),
		UpdatedAt:    r.UpdatedAt(),
	}
}

func toReviewDomain(m *ReviewModel) *reviewDomain.Review {
	var photoURLs []string
	_ = json.Unmarshal([]byte(m.PhotoURLs), &photoURLs)

	return reviewDomain.Reconstruct(
		m.ID,
		m.BookingID,
		m.ReviewerID,
		m.RevieweeID,
		reviewDomain.RevieweeType(m.RevieweeType),
		m.Rating,
		m.Comment,
		photoURLs,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func toReviewDomainSlice(models []ReviewModel) []*reviewDomain.Review {
	reviews := make([]*reviewDomain.Review, len(models))
	for i, m := range models {
		reviews[i] = toReviewDomain(&m)
	}
	return reviews
}
