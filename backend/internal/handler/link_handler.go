package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/backend/internal/service"
)

type LinkHandler struct {
	service *service.LinkService
}

func NewLinkHandler(svc *service.LinkService) *LinkHandler {
	return &LinkHandler{service: svc}
}

// Shorten 创建短链接
// POST /api/v1/shorten
func (h *LinkHandler) Shorten(c *gin.Context) {
	var req service.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	resp, err := h.service.CreateShortLink(c.Request.Context(), &req)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			statusCode := http.StatusInternalServerError
			switch svcErr.Type {
			case "invalid_request":
				statusCode = http.StatusBadRequest
			case "conflict":
				statusCode = http.StatusConflict
			case "not_found":
				statusCode = http.StatusNotFound
			}
			c.JSON(statusCode, gin.H{
				"error":   svcErr.Type,
				"message": svcErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "an unexpected error occurred",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Redirect 短链接重定向
// GET /{code}
func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "code is required",
		})
		return
	}

	longURL, err := h.service.GetLongURL(c.Request.Context(), code)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			if svcErr.Type == "not_found" {
				c.JSON(http.StatusNotFound, gin.H{
					"error":   "not_found",
					"message": svcErr.Message,
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "an unexpected error occurred",
		})
		return
	}

	c.Redirect(http.StatusFound, longURL)
}

// GetLinkInfo 获取短链接信息
// GET /api/v1/links/{code}
func (h *LinkHandler) GetLinkInfo(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "code is required",
		})
		return
	}

	info, err := h.service.GetLinkInfo(c.Request.Context(), code)
	if err != nil {
		if svcErr, ok := err.(*service.ServiceError); ok {
			statusCode := http.StatusInternalServerError
			if svcErr.Type == "not_found" {
				statusCode = http.StatusNotFound
			}
			c.JSON(statusCode, gin.H{
				"error":   svcErr.Type,
				"message": svcErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "an unexpected error occurred",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}
