package repository

import (
	"app/entity"
	"app/usecase/requirement"
	"context"

	"gorm.io/gorm"
)

type RepositoryRequirement struct {
	db *gorm.DB
}

func NewRepositoryRequirement(db *gorm.DB) requirement.IRepositoryRequirement {
	return &RepositoryRequirement{
		db: db,
	}
}

func (r *RepositoryRequirement) Create(ctx context.Context, req *entity.EntityRequirement) (*entity.EntityRequirement, error) {
	err := r.db.WithContext(ctx).Create(req).Error
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (r *RepositoryRequirement) GetByID(ctx context.Context, id int) (*entity.EntityRequirement, error) {
	var req entity.EntityRequirement

	err := r.db.WithContext(ctx).First(&req, id).Error
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (r *RepositoryRequirement) GetByEnvelopeID(ctx context.Context, envelopeID int) ([]entity.EntityRequirement, error) {
	var requirements []entity.EntityRequirement

	err := r.db.WithContext(ctx).Where("envelope_id = ?", envelopeID).
		Order("created_at ASC").Find(&requirements).Error
	if err != nil {
		return nil, err
	}

	return requirements, nil
}

func (r *RepositoryRequirement) Update(ctx context.Context, req *entity.EntityRequirement) (*entity.EntityRequirement, error) {
	err := r.db.WithContext(ctx).Save(req).Error
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (r *RepositoryRequirement) Delete(ctx context.Context, req *entity.EntityRequirement) error {
	err := r.db.WithContext(ctx).Delete(req).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryRequirement) GetByClicksignKey(ctx context.Context, key string) (*entity.EntityRequirement, error) {
	var req entity.EntityRequirement

	err := r.db.WithContext(ctx).Where("clicksign_key = ?", key).First(&req).Error
	if err != nil {
		return nil, err
	}

	return &req, nil
}