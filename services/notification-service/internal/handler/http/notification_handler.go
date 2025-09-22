package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
	"solemate/services/notification-service/internal/domain/service"
)

type NotificationHandler struct {
	notificationService service.NotificationService
	templateService     service.TemplateService
	preferenceService   service.PreferenceService
}

func NewNotificationHandler(
	notificationService service.NotificationService,
	templateService service.TemplateService,
	preferenceService service.PreferenceService,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		templateService:     templateService,
		preferenceService:   preferenceService,
	}
}

func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	notifications := router.Group("/notifications")
	notifications.Use(jwtMiddleware)
	{
		notifications.POST("/send", h.SendNotification)
		notifications.POST("/send-bulk", adminMiddleware, h.SendBulkNotification)
		notifications.POST("/send-template", h.SendTemplateNotification)
		notifications.GET("/:id", h.GetNotification)
		notifications.GET("/user/:userId", h.GetUserNotifications)
		notifications.POST("/process-event", h.ProcessEvent)
		notifications.POST("/retry-failed", adminMiddleware, h.RetryFailedNotifications)
		notifications.PATCH("/:id/cancel", h.CancelNotification)

		notifications.GET("/admin/by-status/:status", adminMiddleware, h.GetNotificationsByStatus)
		notifications.GET("/admin/statistics", adminMiddleware, h.GetStatistics)
		notifications.GET("/admin/delivery-report", adminMiddleware, h.GetDeliveryReport)
		notifications.POST("/admin/process-queue", adminMiddleware, h.ProcessNotificationQueue)
		notifications.POST("/admin/process-scheduled", adminMiddleware, h.ProcessScheduledNotifications)
	}

	templates := router.Group("/notification-templates")
	templates.Use(jwtMiddleware)
	{
		templates.POST("/", adminMiddleware, h.CreateTemplate)
		templates.GET("/:id", h.GetTemplate)
		templates.GET("/name/:name", h.GetTemplateByName)
		templates.PUT("/:id", adminMiddleware, h.UpdateTemplate)
		templates.DELETE("/:id", adminMiddleware, h.DeleteTemplate)
		templates.GET("/", h.ListTemplates)
		templates.POST("/:id/render", h.RenderTemplate)
	}

	preferences := router.Group("/notification-preferences")
	preferences.Use(jwtMiddleware)
	{
		preferences.GET("/user/:userId", h.GetUserPreferences)
		preferences.PUT("/user/:userId", h.UpdateUserPreferences)
		preferences.POST("/user/:userId/check-consent", h.CheckUserConsent)
	}
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var request service.SendNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			request.UserID = uid
		}
	}

	response, err := h.notificationService.SendNotification(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *NotificationHandler) SendBulkNotification(c *gin.Context) {
	var request service.SendBulkNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	responses, err := h.notificationService.SendBulkNotification(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"notifications": responses,
		"count":         len(responses),
	})
}

func (h *NotificationHandler) SendTemplateNotification(c *gin.Context) {
	var request service.SendTemplateNotificationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if exists {
		if uid, ok := userID.(uuid.UUID); ok {
			request.UserID = uid
		}
	}

	response, err := h.notificationService.SendTemplateNotification(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	response, err := h.notificationService.GetNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.notificationService.GetUserNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetNotificationsByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	status := entity.NotificationStatus(statusStr)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.notificationService.GetNotificationsByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) ProcessEvent(c *gin.Context) {
	var request service.EventProcessingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.notificationService.ProcessEvent(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) RetryFailedNotifications(c *gin.Context) {
	var request service.RetryFailedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.notificationService.RetryFailedNotifications(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) CancelNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notificationService.CancelNotification(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification cancelled successfully"})
}

func (h *NotificationHandler) GetStatistics(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format (YYYY-MM-DD)"})
			return
		}
	} else {
		from = time.Now().AddDate(0, 0, -30)
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format (YYYY-MM-DD)"})
			return
		}
	} else {
		to = time.Now()
	}

	response, err := h.notificationService.GetStatistics(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetDeliveryReport(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	groupBy := c.DefaultQuery("groupBy", "day")

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format (YYYY-MM-DD)"})
			return
		}
	} else {
		from = time.Now().AddDate(0, 0, -30)
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format (YYYY-MM-DD)"})
			return
		}
	} else {
		to = time.Now()
	}

	response, err := h.notificationService.GetDeliveryReport(c.Request.Context(), from, to, groupBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) ProcessNotificationQueue(c *gin.Context) {
	batchSize, _ := strconv.Atoi(c.DefaultQuery("batchSize", "10"))

	if err := h.notificationService.ProcessNotificationQueue(c.Request.Context(), batchSize); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Notification queue processed successfully",
		"batch_size": batchSize,
	})
}

func (h *NotificationHandler) ProcessScheduledNotifications(c *gin.Context) {
	if err := h.notificationService.ProcessScheduledNotifications(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled notifications processed successfully"})
}

func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var request service.CreateTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.templateService.CreateTemplate(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	response, err := h.templateService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetTemplateByName(c *gin.Context) {
	name := c.Param("name")

	response, err := h.templateService.GetTemplateByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var request service.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.templateService.UpdateTemplate(c.Request.Context(), id, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.templateService.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.templateService.ListTemplates(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) RenderTemplate(c *gin.Context) {
	templateID := c.Param("id")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.templateService.RenderTemplate(c.Request.Context(), templateID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) GetUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	response, err := h.preferenceService.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) UpdateUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request service.UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.preferenceService.UpdateUserPreferences(c.Request.Context(), userID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) CheckUserConsent(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var request struct {
		Type    entity.NotificationType    `json:"type" binding:"required"`
		Channel entity.NotificationChannel `json:"channel" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consent, err := h.preferenceService.CheckUserConsent(c.Request.Context(), userID, request.Type, request.Channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"type":    request.Type,
		"channel": request.Channel,
		"consent": consent,
	})
}

func (h *NotificationHandler) GetHealth(c *gin.Context) {
	healthResponse := service.HealthResponse{
		Status:        "healthy",
		Service:       "notification-service",
		Database:      "connected",
		Redis:         "connected",
		EmailProvider: "available",
		SMSProvider:   "available",
		QueueStatus: map[string]interface{}{
			"pending": 0,
			"processing": 0,
		},
		PendingCount: 0,
		Timestamp:    time.Now(),
	}

	c.JSON(http.StatusOK, healthResponse)
}