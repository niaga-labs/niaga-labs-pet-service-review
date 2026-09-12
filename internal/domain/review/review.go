package review

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RevieweeType represents who is being reviewed.
type RevieweeType string

const (
	RevieweeRunner RevieweeType = "runner"
	RevieweeShop   RevieweeType = "shop"
	RevieweeOwner  RevieweeType = "owner"
)

// IsValid returns true if the reviewee type is recognized.
func (r RevieweeType) IsValid() bool {
	switch r {
	case RevieweeRunner, RevieweeShop, RevieweeOwner:
		return true
	}
	return false
}

// Review is the aggregate root for the review domain.
type Review struct {
	id           uuid.UUID
	bookingID    uuid.UUID
	reviewerID   uuid.UUID
	revieweeID   uuid.UUID
	revieweeType RevieweeType
	rating       int
	comment      string
	photoURLs    []string
	createdAt    time.Time
	updatedAt    time.Time
}

// NewReview creates a new review.
func NewReview(bookingID, reviewerID, revieweeID uuid.UUID, revieweeType RevieweeType, rating int, comment string, photoURLs []string) (*Review, error) {
	if !revieweeType.IsValid() {
		return nil, fmt.Errorf("invalid reviewee type: %s", revieweeType)
	}
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}
	if reviewerID == revieweeID {
		return nil, fmt.Errorf("cannot review yourself")
	}

	now := time.Now().UTC()
	return &Review{
		id:           uuid.New(),
		bookingID:    bookingID,
		reviewerID:   reviewerID,
		revieweeID:   revieweeID,
		revieweeType: revieweeType,
		rating:       rating,
		comment:      comment,
		photoURLs:    photoURLs,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Reconstruct rebuilds a Review from persistence.
func Reconstruct(id, bookingID, reviewerID, revieweeID uuid.UUID, revieweeType RevieweeType, rating int, comment string, photoURLs []string, createdAt, updatedAt time.Time) *Review {
	return &Review{
		id:           id,
		bookingID:    bookingID,
		reviewerID:   reviewerID,
		revieweeID:   revieweeID,
		revieweeType: revieweeType,
		rating:       rating,
		comment:      comment,
		photoURLs:    photoURLs,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Getters.
func (r *Review) ID() uuid.UUID              { return r.id }
func (r *Review) BookingID() uuid.UUID       { return r.bookingID }
func (r *Review) ReviewerID() uuid.UUID      { return r.reviewerID }
func (r *Review) RevieweeID() uuid.UUID      { return r.revieweeID }
func (r *Review) RevieweeType() RevieweeType { return r.revieweeType }
func (r *Review) Rating() int                { return r.rating }
func (r *Review) Comment() string            { return r.comment }
func (r *Review) PhotoURLs() []string        { return r.photoURLs }
func (r *Review) CreatedAt() time.Time       { return r.createdAt }
func (r *Review) UpdatedAt() time.Time       { return r.updatedAt }
