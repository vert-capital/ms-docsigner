package repository

import (
	"app/entity"
	"app/usecase/document"

	"gorm.io/gorm"
)

type RepositoryDocument struct {
	db *gorm.DB
}

func NewRepositoryDocument(db *gorm.DB) document.IRepositoryDocument {
	return &RepositoryDocument{
		db: db,
	}
}

func (r *RepositoryDocument) GetByID(id int) (*entity.EntityDocument, error) {
	var doc entity.EntityDocument

	err := r.db.First(&doc, id).Error
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *RepositoryDocument) Create(document *entity.EntityDocument) error {
	err := r.db.Create(document).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryDocument) Update(document *entity.EntityDocument) error {
	err := r.db.Save(document).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryDocument) Delete(document *entity.EntityDocument) error {
	err := r.db.Delete(document).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryDocument) GetDocuments(filters entity.EntityDocumentFilters) ([]entity.EntityDocument, error) {
	var documents []entity.EntityDocument

	query := r.db.Model(&entity.EntityDocument{})

	if len(filters.IDs) > 0 {
		query = query.Where("id IN ?", filters.IDs)
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.ClicksignKey != "" {
		query = query.Where("clicksign_key = ?", filters.ClicksignKey)
	}

	err := query.Order("created_at DESC").Find(&documents).Error
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *RepositoryDocument) GetByClicksignKey(key string) (*entity.EntityDocument, error) {
	var doc entity.EntityDocument

	err := r.db.Where("clicksign_key = ?", key).First(&doc).Error
	if err != nil {
		return nil, err
	}

	return &doc, nil
}
