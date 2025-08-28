package routes

import (
	"net/http"
	"strconv"

	"taraftar-analizi/models"
	"taraftar-analizi/services"

	"github.com/gin-gonic/gin"
)

type CommentRoutes struct {
	commentService *services.CommentService
}

func NewCommentRoutes() *CommentRoutes {
	return &CommentRoutes{
		commentService: services.NewCommentService(),
	}
}

func (cr *CommentRoutes) RegisterRoutes(router *gin.RouterGroup) {
	comments := router.Group("/comments")
	{
		comments.GET("", cr.GetComments)
		comments.POST("", cr.CreateComment)
		comments.GET("/unprocessed", cr.GetUnprocessedComments)
		comments.GET("/stats", cr.GetCommentStats)
		comments.GET("/:id", cr.GetComment)
		comments.PUT("/:id", cr.UpdateComment)
		comments.POST("/bulk/processed", cr.BulkUpdateProcessed)
	}
}

func (cr *CommentRoutes) GetComments(c *gin.Context) {
	var query models.CommentQuery
	
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"message": err.Error(),
		})
		return
	}

	comments, err := cr.commentService.GetComments(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get comments",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (cr *CommentRoutes) CreateComment(c *gin.Context) {
	var req models.CommentCreateRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	comment, err := cr.commentService.CreateComment(req)
	if err != nil {
		if err.Error() == "comment already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Duplicate comment",
				"message": "Comment already exists in database",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create comment",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment created successfully",
		"comment": comment,
	})
}

func (cr *CommentRoutes) GetUnprocessedComments(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	comments, err := cr.commentService.GetUnprocessedComments(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get unprocessed comments",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"count":    len(comments),
		"limit":    limit,
	})
}

func (cr *CommentRoutes) GetCommentStats(c *gin.Context) {
	stats, err := cr.commentService.GetCommentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get comment statistics",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (cr *CommentRoutes) GetComment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Comment ID is required",
		})
		return
	}

	// Note: We would need to add ID filtering to the service, but for now this is a placeholder
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Not implemented",
		"message": "Get comment by ID is not yet implemented",
	})
}

func (cr *CommentRoutes) UpdateComment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ID",
			"message": "Comment ID is required",
		})
		return
	}

	var req models.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	err := cr.commentService.UpdateComment(id, req)
	if err != nil {
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Comment not found",
				"message": "Comment with specified ID does not exist",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update comment",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
	})
}

func (cr *CommentRoutes) BulkUpdateProcessed(c *gin.Context) {
	var req struct {
		CommentIDs []string `json:"comment_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if len(req.CommentIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Comment IDs array cannot be empty",
		})
		return
	}

	err := cr.commentService.BulkUpdateProcessed(req.CommentIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to bulk update comments",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Comments updated successfully",
		"updated_count": len(req.CommentIDs),
	})
}

