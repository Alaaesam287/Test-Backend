package handlers

import (
	"net/http"

	"github.com/Secure-Website-Builder/Backend/internal/services/auth"
	"github.com/gin-gonic/gin"
)

type AdminAuthHandler struct {
	service *auth.Service
}

func NewAdminAuthHandler(service *auth.Service) *AdminAuthHandler {
	return &AdminAuthHandler{service: service}
}

type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.AdminLogin(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
