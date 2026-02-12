package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Kilat-Pet-Delivery/lib-common/auth"
	"github.com/Kilat-Pet-Delivery/lib-common/middleware"
	"github.com/Kilat-Pet-Delivery/lib-common/response"
	"github.com/Kilat-Pet-Delivery/service-review/internal/application"
)

// ReviewHandler handles HTTP requests for review operations.
type ReviewHandler struct {
	service *application.ReviewService
}

// NewReviewHandler creates a new ReviewHandler.
func NewReviewHandler(service *application.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

// RegisterRoutes registers all review routes.
func (h *ReviewHandler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager) {
	authMW := middleware.AuthMiddleware(jwtManager)

	reviews := r.Group("/api/v1/reviews")
	{
		reviews.POST("", authMW, h.CreateReview)
		reviews.GET("/me", authMW, h.GetMyReviews)
		reviews.GET("/runner/:id", h.GetRunnerReviews)
		reviews.GET("/shop/:id", h.GetShopReviews)
		reviews.GET("/rating/:type/:id", h.GetRatingSummary)
	}
}

// CreateReview handles POST /api/v1/reviews.
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req application.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.service.CreateReview(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// GetMyReviews handles GET /api/v1/reviews/me.
func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, limit := parsePagination(c)

	reviews, total, err := h.service.GetMyReviews(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, reviews, total, page, limit)
}

// GetRunnerReviews handles GET /api/v1/reviews/runner/:id.
func (h *ReviewHandler) GetRunnerReviews(c *gin.Context) {
	runnerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid runner ID")
		return
	}

	page, limit := parsePagination(c)

	reviews, total, err := h.service.GetReviewsByReviewee(c.Request.Context(), runnerID, "runner", page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, reviews, total, page, limit)
}

// GetShopReviews handles GET /api/v1/reviews/shop/:id.
func (h *ReviewHandler) GetShopReviews(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid shop ID")
		return
	}

	page, limit := parsePagination(c)

	reviews, total, err := h.service.GetReviewsByReviewee(c.Request.Context(), shopID, "shop", page, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, reviews, total, page, limit)
}

// GetRatingSummary handles GET /api/v1/reviews/rating/:type/:id.
func (h *ReviewHandler) GetRatingSummary(c *gin.Context) {
	revieweeType := c.Param("type")
	revieweeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ID")
		return
	}

	result, err := h.service.GetRatingSummary(c.Request.Context(), revieweeID, revieweeType)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
