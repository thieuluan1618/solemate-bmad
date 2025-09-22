package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
	"solemate/services/notification-service/internal/domain/repository"
)

type preferenceService struct {
	preferenceRepo repository.PreferenceRepository
	userRepo       repository.UserRepository
}

func NewPreferenceService(
	preferenceRepo repository.PreferenceRepository,
	userRepo repository.UserRepository,
) PreferenceService {
	return &preferenceService{
		preferenceRepo: preferenceRepo,
		userRepo:       userRepo,
	}
}

func (s *preferenceService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*PreferenceResponse, error) {
	if s.userRepo != nil {
		exists, err := s.userRepo.ValidateUserExists(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate user: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("user does not exist")
		}
	}

	preference, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		defaultPreferences, createErr := s.CreateDefaultPreferences(ctx, userID)
		if createErr != nil {
			return nil, fmt.Errorf("failed to get or create preferences: %w", createErr)
		}
		return defaultPreferences, nil
	}

	return s.toPreferenceResponse(preference), nil
}

func (s *preferenceService) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, request *UpdatePreferenceRequest) (*PreferenceResponse, error) {
	if s.userRepo != nil {
		exists, err := s.userRepo.ValidateUserExists(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate user: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("user does not exist")
		}
	}

	preference, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		preference = &entity.NotificationPreference{
			UserID:             userID,
			EmailNotifications: true,
			SMSNotifications:   false,
			PushNotifications:  true,
			InAppNotifications: true,
			OrderUpdates:       true,
			PaymentUpdates:     true,
			PromotionalEmails:  false,
			Newsletter:         false,
			SecurityAlerts:     true,
			StockAlerts:        false,
			PreferredChannel:   entity.ChannelEmail,
			TimeZone:           "UTC",
		}

		if createErr := s.preferenceRepo.Create(ctx, preference); createErr != nil {
			return nil, fmt.Errorf("failed to create default preferences: %w", createErr)
		}
	}

	if request.EmailNotifications != nil {
		preference.EmailNotifications = *request.EmailNotifications
	}
	if request.SMSNotifications != nil {
		preference.SMSNotifications = *request.SMSNotifications
	}
	if request.PushNotifications != nil {
		preference.PushNotifications = *request.PushNotifications
	}
	if request.InAppNotifications != nil {
		preference.InAppNotifications = *request.InAppNotifications
	}
	if request.OrderUpdates != nil {
		preference.OrderUpdates = *request.OrderUpdates
	}
	if request.PaymentUpdates != nil {
		preference.PaymentUpdates = *request.PaymentUpdates
	}
	if request.PromotionalEmails != nil {
		preference.PromotionalEmails = *request.PromotionalEmails
	}
	if request.Newsletter != nil {
		preference.Newsletter = *request.Newsletter
	}
	if request.SecurityAlerts != nil {
		preference.SecurityAlerts = *request.SecurityAlerts
	}
	if request.StockAlerts != nil {
		preference.StockAlerts = *request.StockAlerts
	}
	if request.PreferredChannel != nil {
		preference.PreferredChannel = *request.PreferredChannel
	}
	if request.TimeZone != nil {
		preference.TimeZone = *request.TimeZone
	}
	if request.QuietHoursStart != nil {
		preference.QuietHoursStart = request.QuietHoursStart
	}
	if request.QuietHoursEnd != nil {
		preference.QuietHoursEnd = request.QuietHoursEnd
	}

	if err := s.preferenceRepo.Update(ctx, preference); err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	return s.toPreferenceResponse(preference), nil
}

func (s *preferenceService) CheckUserConsent(ctx context.Context, userID uuid.UUID, notificationType entity.NotificationType, channel entity.NotificationChannel) (bool, error) {
	preference, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user preferences: %w", err)
	}

	channelConsent := s.checkChannelConsent(preference, channel)
	if !channelConsent {
		return false, nil
	}

	typeConsent := s.checkTypeConsent(preference, notificationType)
	return typeConsent, nil
}

func (s *preferenceService) CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) (*PreferenceResponse, error) {
	preference := &entity.NotificationPreference{
		UserID:             userID,
		EmailNotifications: true,
		SMSNotifications:   false,
		PushNotifications:  true,
		InAppNotifications: true,
		OrderUpdates:       true,
		PaymentUpdates:     true,
		PromotionalEmails:  false,
		Newsletter:         false,
		SecurityAlerts:     true,
		StockAlerts:        false,
		PreferredChannel:   entity.ChannelEmail,
		TimeZone:           "UTC",
	}

	if err := s.preferenceRepo.Create(ctx, preference); err != nil {
		return nil, fmt.Errorf("failed to create default preferences: %w", err)
	}

	return s.toPreferenceResponse(preference), nil
}

func (s *preferenceService) checkChannelConsent(preference *entity.NotificationPreference, channel entity.NotificationChannel) bool {
	switch channel {
	case entity.ChannelEmail:
		return preference.EmailNotifications
	case entity.ChannelSMS:
		return preference.SMSNotifications
	case entity.ChannelPush:
		return preference.PushNotifications
	case entity.ChannelInApp:
		return preference.InAppNotifications
	default:
		return false
	}
}

func (s *preferenceService) checkTypeConsent(preference *entity.NotificationPreference, notificationType entity.NotificationType) bool {
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
	case entity.NotificationTypeWelcome, entity.NotificationTypePasswordReset:
		return preference.SecurityAlerts
	default:
		return true
	}
}

func (s *preferenceService) toPreferenceResponse(preference *entity.NotificationPreference) *PreferenceResponse {
	return &PreferenceResponse{
		ID:                    preference.ID,
		UserID:                preference.UserID,
		EmailNotifications:    preference.EmailNotifications,
		SMSNotifications:      preference.SMSNotifications,
		PushNotifications:     preference.PushNotifications,
		InAppNotifications:    preference.InAppNotifications,
		OrderUpdates:          preference.OrderUpdates,
		PaymentUpdates:        preference.PaymentUpdates,
		PromotionalEmails:     preference.PromotionalEmails,
		Newsletter:            preference.Newsletter,
		SecurityAlerts:        preference.SecurityAlerts,
		StockAlerts:           preference.StockAlerts,
		PreferredChannel:      preference.PreferredChannel,
		TimeZone:              preference.TimeZone,
		QuietHoursStart:       preference.QuietHoursStart,
		QuietHoursEnd:         preference.QuietHoursEnd,
		CreatedAt:             preference.CreatedAt,
		UpdatedAt:             preference.UpdatedAt,
	}
}