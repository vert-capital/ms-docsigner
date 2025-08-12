package repository

import (
	"app/entity"
	"fmt"

	"gorm.io/gorm"
)

type RepositoryWebhook struct {
	db *gorm.DB
}

func NewRepositoryWebhook(db *gorm.DB) *RepositoryWebhook {
	return &RepositoryWebhook{db: db}
}

// Create cria um novo webhook
func (r *RepositoryWebhook) Create(webhook *entity.EntityWebhook) error {
	result := r.db.Create(webhook)
	if result.Error != nil {
		return fmt.Errorf("failed to create webhook: %w", result.Error)
	}
	return nil
}

// GetByID busca um webhook por ID
func (r *RepositoryWebhook) GetByID(id int) (*entity.EntityWebhook, error) {
	var webhook entity.EntityWebhook
	result := r.db.First(&webhook, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get webhook: %w", result.Error)
	}
	return &webhook, nil
}

// GetByDocumentKey busca webhooks por document key
func (r *RepositoryWebhook) GetByDocumentKey(documentKey string) ([]entity.EntityWebhook, error) {
	var webhooks []entity.EntityWebhook
	result := r.db.Where("document_key = ?", documentKey).Order("created_at DESC").Find(&webhooks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get webhooks by document key: %w", result.Error)
	}
	return webhooks, nil
}

// GetByAccountKey busca webhooks por account key
func (r *RepositoryWebhook) GetByAccountKey(accountKey string) ([]entity.EntityWebhook, error) {
	var webhooks []entity.EntityWebhook
	result := r.db.Where("account_key = ?", accountKey).Order("created_at DESC").Find(&webhooks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get webhooks by account key: %w", result.Error)
	}
	return webhooks, nil
}

// GetByEventName busca webhooks por nome do evento
func (r *RepositoryWebhook) GetByEventName(eventName string) ([]entity.EntityWebhook, error) {
	var webhooks []entity.EntityWebhook
	result := r.db.Where("event_name = ?", eventName).Order("created_at DESC").Find(&webhooks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get webhooks by event name: %w", result.Error)
	}
	return webhooks, nil
}

// GetByStatus busca webhooks por status
func (r *RepositoryWebhook) GetByStatus(status string) ([]entity.EntityWebhook, error) {
	var webhooks []entity.EntityWebhook
	result := r.db.Where("status = ?", status).Order("created_at DESC").Find(&webhooks)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get webhooks by status: %w", result.Error)
	}
	return webhooks, nil
}

// GetPending busca webhooks pendentes
func (r *RepositoryWebhook) GetPending() ([]entity.EntityWebhook, error) {
	return r.GetByStatus("pending")
}

// GetFailed busca webhooks que falharam
func (r *RepositoryWebhook) GetFailed() ([]entity.EntityWebhook, error) {
	return r.GetByStatus("failed")
}

// GetAll busca todos os webhooks com paginação
func (r *RepositoryWebhook) GetAll(page, limit int) ([]entity.EntityWebhook, int64, error) {
	var webhooks []entity.EntityWebhook
	var total int64

	// Contar total
	if err := r.db.Model(&entity.EntityWebhook{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	// Buscar com paginação
	offset := (page - 1) * limit
	result := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&webhooks)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to get webhooks: %w", result.Error)
	}

	return webhooks, total, nil
}

// Update atualiza um webhook
func (r *RepositoryWebhook) Update(webhook *entity.EntityWebhook) error {
	result := r.db.Save(webhook)
	if result.Error != nil {
		return fmt.Errorf("failed to update webhook: %w", result.Error)
	}
	return nil
}

// Delete deleta um webhook
func (r *RepositoryWebhook) Delete(id int) error {
	result := r.db.Delete(&entity.EntityWebhook{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete webhook: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook not found with id: %d", id)
	}
	return nil
}

// MarkAsProcessed marca um webhook como processado
func (r *RepositoryWebhook) MarkAsProcessed(id int) error {
	result := r.db.Model(&entity.EntityWebhook{}).Where("id = ?", id).Update("status", "processed")
	if result.Error != nil {
		return fmt.Errorf("failed to mark webhook as processed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook not found with id: %d", id)
	}
	return nil
}

// MarkAsFailed marca um webhook como falhou
func (r *RepositoryWebhook) MarkAsFailed(id int, errorMsg string) error {
	result := r.db.Model(&entity.EntityWebhook{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": "failed",
		"error":  errorMsg,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to mark webhook as failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook not found with id: %d", id)
	}
	return nil
}

// GetByFilters busca webhooks com filtros
func (r *RepositoryWebhook) GetByFilters(eventName, documentKey, accountKey, status string, page, limit int) ([]entity.EntityWebhook, int64, error) {
	var webhooks []entity.EntityWebhook
	var total int64

	query := r.db.Model(&entity.EntityWebhook{})

	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}
	if documentKey != "" {
		query = query.Where("document_key = ?", documentKey)
	}
	if accountKey != "" {
		query = query.Where("account_key = ?", accountKey)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Contar total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count webhooks with filters: %w", err)
	}

	// Buscar com paginação
	offset := (page - 1) * limit
	result := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&webhooks)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to get webhooks with filters: %w", result.Error)
	}

	return webhooks, total, nil
}
