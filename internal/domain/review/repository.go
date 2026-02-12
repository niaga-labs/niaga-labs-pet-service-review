package review

import (
	"context"

	"github.com/google/uuid"
)

// ReviewRepository defines persistence operations for reviews.
type ReviewRepository interface {
	Save(ctx context.Context, review *Review) error
	FindByID(ctx context.Context, id uuid.UUID) (*Review, error)
	FindByBookingID(ctx context.Context, bookingID uuid.UUID) ([]*Review, error)
	FindByRevieweeID(ctx context.Context, revieweeID uuid.UUID, revieweeType RevieweeType, limit, offset int) ([]*Review, int64, error)
	FindByReviewerID(ctx context.Context, reviewerID uuid.UUID, limit, offset int) ([]*Review, int64, error)
	AverageRating(ctx context.Context, revieweeID uuid.UUID, revieweeType RevieweeType) (float64, int64, error)
	ExistsForBooking(ctx context.Context, bookingID, reviewerID uuid.UUID) (bool, error)
}
