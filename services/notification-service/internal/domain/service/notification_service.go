package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
	"solemate/services/notification-service/internal/domain/repository"
)

type NotificationService interface {
	SendNotification(ctx context.Context, request *SendNotificationRequest) (*NotificationResponse, error)
	SendBulkNotification(ctx context.Context, request *SendBulkNotificationRequest) ([]*NotificationResponse, error)
	SendTemplateNotification(ctx context.Context, request *SendTemplateNotificationRequest) (*NotificationResponse, error)
	GetNotification(ctx context.Context, id uuid.UUID) (*NotificationResponse, error)
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) (*NotificationListResponse, error)
	GetNotificationsByStatus(ctx context.Context, status entity.NotificationStatus, limit, offset int) (*NotificationListResponse, error)
	RetryFailedNotifications(ctx context.Context, request *RetryFailedRequest) (*RetryFailedResponse, error)
	CancelNotification(ctx context.Context, id uuid.UUID) error
	ProcessScheduledNotifications(ctx context.Context) error
	ProcessNotificationQueue(ctx context.Context, batchSize int) error
	GetStatistics(ctx context.Context, from, to time.Time) (*StatisticsResponse, error)
	GetDeliveryReport(ctx context.Context, from, to time.Time, groupBy string) (*DeliveryReportResponse, error)
	ProcessEvent(ctx context.Context, request *EventProcessingRequest) (*EventProcessingResponse, error)
}

type TemplateService interface {
	CreateTemplate(ctx context.Context, request *CreateTemplateRequest) (*TemplateResponse, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*TemplateResponse, error)
	GetTemplateByName(ctx context.Context, name string) (*TemplateResponse, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, request *UpdateTemplateRequest) (*TemplateResponse, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	ListTemplates(ctx context.Context, limit, offset int) (*TemplateListResponse, error)
	RenderTemplate(ctx context.Context, templateID string, data map[string]interface{}) (*RenderedTemplate, error)
}

type PreferenceService interface {
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*PreferenceResponse, error)
	UpdateUserPreferences(ctx context.Context, userID uuid.UUID, request *UpdatePreferenceRequest) (*PreferenceResponse, error)
	CheckUserConsent(ctx context.Context, userID uuid.UUID, notificationType entity.NotificationType, channel entity.NotificationChannel) (bool, error)
	CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) (*PreferenceResponse, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
	templateRepo     repository.TemplateRepository
	preferenceRepo   repository.PreferenceRepository
	logRepo          repository.LogRepository
	queueRepo        repository.QueueRepository
	eventRepo        repository.EventRepository
	userRepo         repository.UserRepository
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	templateRepo repository.TemplateRepository,
	preferenceRepo repository.PreferenceRepository,
	logRepo repository.LogRepository,
	queueRepo repository.QueueRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		preferenceRepo:   preferenceRepo,
		logRepo:          logRepo,
		queueRepo:        queueRepo,
		eventRepo:        eventRepo,
		userRepo:         userRepo,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, request *SendNotificationRequest) (*NotificationResponse, error) {
	if s.userRepo != nil {
		exists, err := s.userRepo.ValidateUserExists(ctx, request.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate user: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("user does not exist")
		}
	}

	preference, err := s.preferenceRepo.GetByUserID(ctx, request.UserID)
	if err != nil {
		preference = &entity.NotificationPreference{
			UserID:             request.UserID,
			EmailNotifications: true,
			SMSNotifications:   false,
			PushNotifications:  true,
			InAppNotifications: true,
		}
	}

	if !s.checkChannelConsent(preference, request.Channel, request.Type) {
		return nil, fmt.Errorf("user has not consented to receive %s notifications via %s", request.Type, request.Channel)
	}

	if request.Priority == "" {
		request.Priority = entity.PriorityMedium
	}

	notification := &entity.Notification{
		UserID:            request.UserID,
		Type:              request.Type,
		Channel:           request.Channel,
		Status:            entity.StatusPending,
		Priority:          request.Priority,
		Subject:           request.Subject,
		Content:           request.Content,
		HTMLContent:       request.HTMLContent,
		RecipientEmail:    request.RecipientEmail,
		RecipientPhone:    request.RecipientPhone,
		TemplateID:        request.TemplateID,
		TemplateData:      request.TemplateData,
		Metadata:          request.Metadata,
		RelatedEntityID:   request.RelatedEntityID,
		RelatedEntityType: request.RelatedEntityType,
		ScheduledAt:       request.ScheduledAt,
		MaxRetries:        3,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	queueItem := &entity.NotificationQueue{
		NotificationID: notification.ID,
		Priority:       request.Priority,
		ScheduledAt:    time.Now(),
	}

	if request.ScheduledAt != nil {
		queueItem.ScheduledAt = *request.ScheduledAt
	}

	if err := s.queueRepo.Enqueue(ctx, queueItem); err != nil {
		return nil, fmt.Errorf("failed to enqueue notification: %w", err)
	}

	return s.toNotificationResponse(notification), nil
}

func (s *notificationService) SendBulkNotification(ctx context.Context, request *SendBulkNotificationRequest) ([]*NotificationResponse, error) {
	if len(request.UserIDs) == 0 {
		return nil, fmt.Errorf("no user IDs provided")
	}

	preferences, err := s.preferenceRepo.BulkGetPreferences(ctx, request.UserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	var responses []*NotificationResponse
	var notifications []*entity.Notification

	for _, userID := range request.UserIDs {
		preference, exists := preferences[userID]
		if !exists {
			preference = &entity.NotificationPreference{
				UserID:             userID,
				EmailNotifications: true,
				SMSNotifications:   false,
				PushNotifications:  true,
				InAppNotifications: true,
			}
		}

		if !s.checkChannelConsent(preference, request.Channel, request.Type) {
			continue
		}

		notification := &entity.Notification{
			UserID:      userID,
			Type:        request.Type,
			Channel:     request.Channel,
			Status:      entity.StatusPending,
			Priority:    request.Priority,
			Subject:     request.Subject,
			Content:     request.Content,
			HTMLContent: request.HTMLContent,
			TemplateID:  request.TemplateID,
			TemplateData: request.TemplateData,
			Metadata:    request.Metadata,
			ScheduledAt: request.ScheduledAt,
			MaxRetries:  3,
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			continue
		}

		queueItem := &entity.NotificationQueue{
			NotificationID: notification.ID,
			Priority:       request.Priority,
			ScheduledAt:    time.Now(),
		}

		if request.ScheduledAt != nil {
			queueItem.ScheduledAt = *request.ScheduledAt
		}

		if err := s.queueRepo.Enqueue(ctx, queueItem); err != nil {
			continue
		}

		notifications = append(notifications, notification)
		responses = append(responses, s.toNotificationResponse(notification))
	}

	return responses, nil
}

func (s *notificationService) SendTemplateNotification(ctx context.Context, request *SendTemplateNotificationRequest) (*NotificationResponse, error) {
	template, err := s.templateRepo.GetByName(ctx, request.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active")
	}

	renderedContent, err := s.renderTemplateContent(template.Content, request.TemplateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render template content: %w", err)
	}

	renderedSubject, err := s.renderTemplateContent(template.Subject, request.TemplateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render template subject: %w", err)
	}

	var renderedHTML *string
	if template.HTMLContent != nil {
		html, err := s.renderTemplateContent(*template.HTMLContent, request.TemplateData)
		if err != nil {
			return nil, fmt.Errorf("failed to render template HTML: %w", err)
		}
		renderedHTML = &html
	}

	sendRequest := &SendNotificationRequest{
		UserID:            request.UserID,
		Type:              template.Type,
		Channel:           request.Channel,
		Priority:          request.Priority,
		Subject:           renderedSubject,
		Content:           renderedContent,
		HTMLContent:       renderedHTML,
		RecipientEmail:    request.RecipientEmail,
		RecipientPhone:    request.RecipientPhone,
		TemplateID:        &request.TemplateID,
		TemplateData:      request.TemplateData,
		Metadata:          request.Metadata,
		RelatedEntityID:   request.RelatedEntityID,
		RelatedEntityType: request.RelatedEntityType,
		ScheduledAt:       request.ScheduledAt,
	}

	return s.SendNotification(ctx, sendRequest)
}

func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return s.toNotificationResponse(notification), nil
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) (*NotificationListResponse, error) {
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}

	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = s.toNotificationResponse(notification)
	}

	return &NotificationListResponse{
		Notifications: responses,
		Total:         int64(len(responses)),
		Limit:         limit,
		Offset:        offset,
	}, nil
}

func (s *notificationService) GetNotificationsByStatus(ctx context.Context, status entity.NotificationStatus, limit, offset int) (*NotificationListResponse, error) {
	notifications, err := s.notificationRepo.GetByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by status: %w", err)
	}

	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = s.toNotificationResponse(notification)
	}

	return &NotificationListResponse{
		Notifications: responses,
		Total:         int64(len(responses)),
		Limit:         limit,
		Offset:        offset,
	}, nil
}

func (s *notificationService) RetryFailedNotifications(ctx context.Context, request *RetryFailedRequest) (*RetryFailedResponse, error) {
	processed := 0
	failed := 0
	startedAt := time.Now()

	for _, notificationID := range request.NotificationIDs {
		notification, err := s.notificationRepo.GetByID(ctx, notificationID)
		if err != nil {
			failed++
			continue
		}

		if notification.Status != entity.StatusFailed {
			failed++
			continue
		}

		maxRetries := notification.MaxRetries
		if request.MaxRetries != nil {
			maxRetries = *request.MaxRetries
		}

		if notification.RetryCount >= maxRetries {
			failed++
			continue
		}

		notification.Status = entity.StatusPending
		if err := s.notificationRepo.Update(ctx, notification); err != nil {
			failed++
			continue
		}

		queueItem := &entity.NotificationQueue{
			NotificationID: notification.ID,
			Priority:       notification.Priority,
			ScheduledAt:    time.Now(),
		}

		if err := s.queueRepo.Enqueue(ctx, queueItem); err != nil {
			failed++
			continue
		}

		processed++
	}

	return &RetryFailedResponse{
		Processed: processed,
		Failed:    failed,
		StartedAt: startedAt,
	}, nil
}

func (s *notificationService) CancelNotification(ctx context.Context, id uuid.UUID) error {
	return s.notificationRepo.UpdateStatus(ctx, id, entity.StatusCancelled)
}

func (s *notificationService) ProcessScheduledNotifications(ctx context.Context) error {
	now := time.Now()
	notifications, err := s.notificationRepo.GetScheduledNotifications(ctx, now, 100)
	if err != nil {
		return fmt.Errorf("failed to get scheduled notifications: %w", err)
	}

	for _, notification := range notifications {
		queueItem := &entity.NotificationQueue{
			NotificationID: notification.ID,
			Priority:       notification.Priority,
			ScheduledAt:    now,
		}

		if err := s.queueRepo.Enqueue(ctx, queueItem); err != nil {
			continue
		}
	}

	return nil
}

func (s *notificationService) ProcessNotificationQueue(ctx context.Context, batchSize int) error {
	priorities := []entity.NotificationPriority{
		entity.PriorityCritical,
		entity.PriorityHigh,
		entity.PriorityMedium,
		entity.PriorityLow,
	}

	for _, priority := range priorities {
		queueItems, err := s.queueRepo.Dequeue(ctx, priority, batchSize)
		if err != nil {
			continue
		}

		for _, queueItem := range queueItems {
			if err := s.processQueueItem(ctx, queueItem); err != nil {
				continue
			}

			processedAt := time.Now()
			if err := s.queueRepo.MarkAsProcessed(ctx, queueItem.ID, processedAt); err != nil {
				continue
			}
		}

		if len(queueItems) > 0 {
			break
		}
	}

	return nil
}

func (s *notificationService) GetStatistics(ctx context.Context, from, to time.Time) (*StatisticsResponse, error) {
	stats, err := s.notificationRepo.GetStatistics(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &StatisticsResponse{
		TotalSent:      stats.TotalSent,
		TotalDelivered: stats.TotalDelivered,
		TotalFailed:    stats.TotalFailed,
		TotalPending:   stats.TotalPending,
		DeliveryRate:   stats.DeliveryRate,
		FailureRate:    stats.FailureRate,
		ByChannel:      s.convertChannelStats(stats.ByChannel),
		ByType:         s.convertTypeStats(stats.ByType),
		Period:         fmt.Sprintf("%s to %s", from.Format("2006-01-02"), to.Format("2006-01-02")),
		From:           from,
		To:             to,
	}, nil
}

func (s *notificationService) GetDeliveryReport(ctx context.Context, from, to time.Time, groupBy string) (*DeliveryReportResponse, error) {
	reports, err := s.notificationRepo.GetDeliveryReport(ctx, from, to, groupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery report: %w", err)
	}

	reportItems := make([]*DeliveryReportItem, len(reports))
	var totalSent, totalDelivered, totalFailed int64

	for i, report := range reports {
		reportItems[i] = &DeliveryReportItem{
			Period:         report.Period,
			TotalSent:      report.TotalSent,
			TotalDelivered: report.TotalDelivered,
			TotalFailed:    report.TotalFailed,
			DeliveryRate:   report.DeliveryRate,
		}

		totalSent += report.TotalSent
		totalDelivered += report.TotalDelivered
		totalFailed += report.TotalFailed
	}

	overallRate := float64(0)
	if totalSent > 0 {
		overallRate = float64(totalDelivered) / float64(totalSent) * 100
	}

	summary := &DeliveryReportSummary{
		TotalSent:      totalSent,
		TotalDelivered: totalDelivered,
		TotalFailed:    totalFailed,
		OverallRate:    overallRate,
	}

	return &DeliveryReportResponse{
		Reports: reportItems,
		Summary: summary,
		Period:  groupBy,
		From:    from,
		To:      to,
	}, nil
}

func (s *notificationService) ProcessEvent(ctx context.Context, request *EventProcessingRequest) (*EventProcessingResponse, error) {
	event := &entity.NotificationEvent{
		EventType:  request.EventType,
		EntityID:   request.EntityID,
		EntityType: request.EntityType,
		UserID:     request.UserID,
		Payload:    request.Payload,
		Processed:  false,
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	notificationsCreated := 0

	switch request.EventType {
	case "order.created", "order.confirmed", "order.shipped", "order.delivered", "order.cancelled":
		if request.UserID != nil {
			notificationType := s.mapOrderEventToNotificationType(request.EventType)
			if err := s.createOrderNotification(ctx, *request.UserID, notificationType, request.Payload); err == nil {
				notificationsCreated++
			}
		}
	case "payment.successful", "payment.failed":
		if request.UserID != nil {
			notificationType := s.mapPaymentEventToNotificationType(request.EventType)
			if err := s.createPaymentNotification(ctx, *request.UserID, notificationType, request.Payload); err == nil {
				notificationsCreated++
			}
		}
	}

	if err := s.eventRepo.MarkAsProcessed(ctx, event.ID); err != nil {
		return nil, fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return &EventProcessingResponse{
		EventID:              event.ID,
		NotificationsCreated: notificationsCreated,
		ProcessedAt:          time.Now(),
	}, nil
}

func (s *notificationService) checkChannelConsent(preference *entity.NotificationPreference, channel entity.NotificationChannel, notificationType entity.NotificationType) bool {
	switch channel {
	case entity.ChannelEmail:
		if !preference.EmailNotifications {
			return false
		}
	case entity.ChannelSMS:
		if !preference.SMSNotifications {
			return false
		}
	case entity.ChannelPush:
		if !preference.PushNotifications {
			return false
		}
	case entity.ChannelInApp:
		if !preference.InAppNotifications {
			return false
		}
	}

	switch notificationType {
	case entity.NotificationTypeOrderCreated, entity.NotificationTypeOrderConfirmed,
		 entity.NotificationTypeOrderShipped, entity.NotificationTypeOrderDelivered, entity.NotificationTypeOrderCancelled:
		return preference.OrderUpdates
	case entity.NotificationTypePaymentSuccessful, entity.NotificationTypePaymentFailed:
		return preference.PaymentUpdates
	case entity.NotificationTypePromotion:
		return preference.PromotionalEmails
	case entity.NotificationTypeNewsletter:
		return preference.Newsletter
	case entity.NotificationTypeStockAlert:
		return preference.StockAlerts
	default:
		return true
	}
}

func (s *notificationService) renderTemplateContent(template string, data map[string]interface{}) (string, error) {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	return result, nil
}

func (s *notificationService) processQueueItem(ctx context.Context, queueItem *entity.NotificationQueue) error {
	return nil
}

func (s *notificationService) mapOrderEventToNotificationType(eventType string) entity.NotificationType {
	switch eventType {
	case "order.created":
		return entity.NotificationTypeOrderCreated
	case "order.confirmed":
		return entity.NotificationTypeOrderConfirmed
	case "order.shipped":
		return entity.NotificationTypeOrderShipped
	case "order.delivered":
		return entity.NotificationTypeOrderDelivered
	case "order.cancelled":
		return entity.NotificationTypeOrderCancelled
	default:
		return entity.NotificationTypeOrderCreated
	}
}

func (s *notificationService) mapPaymentEventToNotificationType(eventType string) entity.NotificationType {
	switch eventType {
	case "payment.successful":
		return entity.NotificationTypePaymentSuccessful
	case "payment.failed":
		return entity.NotificationTypePaymentFailed
	default:
		return entity.NotificationTypePaymentSuccessful
	}
}

func (s *notificationService) createOrderNotification(ctx context.Context, userID uuid.UUID, notificationType entity.NotificationType, payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	orderTotal, _ := payload["total"].(float64)

	subject := fmt.Sprintf("Order Update - %s", notificationType)
	content := fmt.Sprintf("Your order %s has been updated. Status: %s", orderID, notificationType)

	request := &SendNotificationRequest{
		UserID:   userID,
		Type:     notificationType,
		Channel:  entity.ChannelEmail,
		Priority: entity.PriorityHigh,
		Subject:  subject,
		Content:  content,
		Metadata: map[string]interface{}{
			"order_id": orderID,
			"total":    orderTotal,
		},
	}

	_, err := s.SendNotification(ctx, request)
	return err
}

func (s *notificationService) createPaymentNotification(ctx context.Context, userID uuid.UUID, notificationType entity.NotificationType, payload map[string]interface{}) error {
	paymentID, _ := payload["payment_id"].(string)
	amount, _ := payload["amount"].(float64)

	subject := fmt.Sprintf("Payment Update - %s", notificationType)
	content := fmt.Sprintf("Your payment %s has been updated. Status: %s", paymentID, notificationType)

	request := &SendNotificationRequest{
		UserID:   userID,
		Type:     notificationType,
		Channel:  entity.ChannelEmail,
		Priority: entity.PriorityHigh,
		Subject:  subject,
		Content:  content,
		Metadata: map[string]interface{}{
			"payment_id": paymentID,
			"amount":     amount,
		},
	}

	_, err := s.SendNotification(ctx, request)
	return err
}

func (s *notificationService) convertChannelStats(stats map[entity.NotificationChannel]repository.ChannelStats) map[entity.NotificationChannel]ChannelStats {
	result := make(map[entity.NotificationChannel]ChannelStats)
	for channel, stat := range stats {
		result[channel] = ChannelStats{
			Sent:      stat.Sent,
			Delivered: stat.Delivered,
			Failed:    stat.Failed,
			Rate:      stat.Rate,
		}
	}
	return result
}

func (s *notificationService) convertTypeStats(stats map[entity.NotificationType]repository.TypeStats) map[entity.NotificationType]TypeStats {
	result := make(map[entity.NotificationType]TypeStats)
	for notificationType, stat := range stats {
		result[notificationType] = TypeStats{
			Sent:      stat.Sent,
			Delivered: stat.Delivered,
			Failed:    stat.Failed,
			Rate:      stat.Rate,
		}
	}
	return result
}

func (s *notificationService) toNotificationResponse(notification *entity.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:                notification.ID,
		UserID:            notification.UserID,
		Type:              notification.Type,
		Channel:           notification.Channel,
		Status:            notification.Status,
		Priority:          notification.Priority,
		Subject:           notification.Subject,
		Content:           notification.Content,
		RecipientEmail:    notification.RecipientEmail,
		RecipientPhone:    notification.RecipientPhone,
		RelatedEntityID:   notification.RelatedEntityID,
		RelatedEntityType: notification.RelatedEntityType,
		ScheduledAt:       notification.ScheduledAt,
		SentAt:            notification.SentAt,
		DeliveredAt:       notification.DeliveredAt,
		FailedAt:          notification.FailedAt,
		RetryCount:        notification.RetryCount,
		ErrorMessage:      notification.ErrorMessage,
		CreatedAt:         notification.CreatedAt,
		UpdatedAt:         notification.UpdatedAt,
	}
}

type RenderedTemplate struct {
	Subject     string  `json:"subject"`
	Content     string  `json:"content"`
	HTMLContent *string `json:"html_content"`
}