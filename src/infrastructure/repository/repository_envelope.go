package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryEnvelope struct {
	db *gorm.DB
}

func NewRepositoryEnvelope(db *gorm.DB) *RepositoryEnvelope {
	return &RepositoryEnvelope{
		db: db,
	}
}

func (r *RepositoryEnvelope) GetByID(id int) (*entity.EntityEnvelope, error) {
	var env entity.EntityEnvelope

	err := r.db.First(&env, id).Error
	if err != nil {
		return nil, err
	}

	return &env, nil
}

func (r *RepositoryEnvelope) Create(envelope *entity.EntityEnvelope) error {
	err := r.db.Create(envelope).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryEnvelope) Update(envelope *entity.EntityEnvelope) error {
	err := r.db.Save(envelope).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryEnvelope) Delete(envelope *entity.EntityEnvelope) error {
	err := r.db.Delete(envelope).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryEnvelope) GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error) {
	var envelopes []entity.EntityEnvelope

	query := r.db.Model(&entity.EntityEnvelope{})

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

	err := query.Order("created_at DESC").Find(&envelopes).Error
	if err != nil {
		return nil, err
	}

	return envelopes, nil
}

func (r *RepositoryEnvelope) GetByClicksignKey(key string) (*entity.EntityEnvelope, error) {
	var env entity.EntityEnvelope

	err := r.db.Where("clicksign_key = ?", key).First(&env).Error
	if err != nil {
		return nil, err
	}

	return &env, nil
}
