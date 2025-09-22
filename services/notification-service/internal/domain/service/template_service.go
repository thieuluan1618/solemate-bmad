package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"solemate/services/notification-service/internal/domain/entity"
	"solemate/services/notification-service/internal/domain/repository"
)

type templateService struct {
	templateRepo repository.TemplateRepository
}

func NewTemplateService(templateRepo repository.TemplateRepository) TemplateService {
	return &templateService{
		templateRepo: templateRepo,
	}
}

func (s *templateService) CreateTemplate(ctx context.Context, request *CreateTemplateRequest) (*TemplateResponse, error) {
	template := &entity.NotificationTemplate{
		Name:        request.Name,
		Type:        request.Type,
		Channel:     request.Channel,
		Subject:     request.Subject,
		Content:     request.Content,
		HTMLContent: request.HTMLContent,
		Variables:   request.Variables,
		IsActive:    true,
		Version:     1,
		Description: request.Description,
		SenderEmail: request.SenderEmail,
		SenderName:  request.SenderName,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return s.toTemplateResponse(template), nil
}

func (s *templateService) GetTemplate(ctx context.Context, id uuid.UUID) (*TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return s.toTemplateResponse(template), nil
}

func (s *templateService) GetTemplateByName(ctx context.Context, name string) (*TemplateResponse, error) {
	template, err := s.templateRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get template by name: %w", err)
	}

	return s.toTemplateResponse(template), nil
}

func (s *templateService) UpdateTemplate(ctx context.Context, id uuid.UUID, request *UpdateTemplateRequest) (*TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if request.Name != nil {
		template.Name = *request.Name
	}
	if request.Subject != nil {
		template.Subject = *request.Subject
	}
	if request.Content != nil {
		template.Content = *request.Content
	}
	if request.HTMLContent != nil {
		template.HTMLContent = request.HTMLContent
	}
	if request.Variables != nil {
		template.Variables = request.Variables
	}
	if request.IsActive != nil {
		template.IsActive = *request.IsActive
	}
	if request.Description != nil {
		template.Description = request.Description
	}
	if request.SenderEmail != nil {
		template.SenderEmail = request.SenderEmail
	}
	if request.SenderName != nil {
		template.SenderName = request.SenderName
	}

	template.Version++

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return s.toTemplateResponse(template), nil
}

func (s *templateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	if err := s.templateRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

func (s *templateService) ListTemplates(ctx context.Context, limit, offset int) (*TemplateListResponse, error) {
	templates, err := s.templateRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	responses := make([]*TemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = s.toTemplateResponse(template)
	}

	return &TemplateListResponse{
		Templates: responses,
		Total:     int64(len(responses)),
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func (s *templateService) RenderTemplate(ctx context.Context, templateID string, data map[string]interface{}) (*RenderedTemplate, error) {
	template, err := s.templateRepo.GetByName(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active")
	}

	renderedSubject, err := s.renderContent(template.Subject, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	renderedContent, err := s.renderContent(template.Content, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}

	var renderedHTML *string
	if template.HTMLContent != nil {
		html, err := s.renderContent(*template.HTMLContent, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render HTML content: %w", err)
		}
		renderedHTML = &html
	}

	return &RenderedTemplate{
		Subject:     renderedSubject,
		Content:     renderedContent,
		HTMLContent: renderedHTML,
	}, nil
}

func (s *templateService) renderContent(template string, data map[string]interface{}) (string, error) {
	result := template

	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	if strings.Contains(result, "{{") && strings.Contains(result, "}}") {
		return result, fmt.Errorf("template contains unresolved variables")
	}

	return result, nil
}

func (s *templateService) toTemplateResponse(template *entity.NotificationTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Type:        template.Type,
		Channel:     template.Channel,
		Subject:     template.Subject,
		Content:     template.Content,
		HTMLContent: template.HTMLContent,
		Variables:   template.Variables,
		IsActive:    template.IsActive,
		Version:     template.Version,
		Description: template.Description,
		SenderEmail: template.SenderEmail,
		SenderName:  template.SenderName,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}