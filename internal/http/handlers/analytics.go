package handlers

import (
	"github.com/Secure-Website-Builder/Backend/internal/services/analytics"
)

type AnalyticsHandler struct {
	service *analytics.Service
}

func NewAnalyticsHandler(s *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{service: s}
}
