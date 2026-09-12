-- service-review shipped with no migrations/ directory at all, while
-- cmd/server/main.go calls RunMigrations("migrations") for every environment
-- that is not APP_ENV=development -- so the service could not boot outside
-- development. See KPD-58.
--
-- The CHECK constraints mirror the invariants NewReview already enforces in
-- internal/domain/review/review.go.

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL,
    reviewer_id UUID NOT NULL,
    reviewee_id UUID NOT NULL,
    reviewee_type VARCHAR(20) NOT NULL CHECK (reviewee_type IN ('runner', 'shop', 'owner')),
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    photo_urls JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reviews_reviewer_not_reviewee CHECK (reviewer_id <> reviewee_id),
    CONSTRAINT reviews_one_per_booking_per_reviewer UNIQUE (booking_id, reviewer_id)
);

CREATE INDEX idx_reviews_booking ON reviews(booking_id);
CREATE INDEX idx_reviews_reviewer ON reviews(reviewer_id);
CREATE INDEX idx_reviewee ON reviews(reviewee_id, reviewee_type);
