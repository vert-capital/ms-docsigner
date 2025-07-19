package repository

import (
	"app/entity"
	"app/usecase/signatory"

	"gorm.io/gorm"
)

type RepositorySignatory struct {
	db *gorm.DB
}

func NewRepositorySignatory(db *gorm.DB) signatory.IRepositorySignatory {
	return &RepositorySignatory{
		db: db,
	}
}

func (r *RepositorySignatory) GetByID(id int) (*entity.EntitySignatory, error) {
	var sig entity.EntitySignatory

	err := r.db.First(&sig, id).Error
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func (r *RepositorySignatory) Create(signatory *entity.EntitySignatory) error {
	err := r.db.Create(signatory).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositorySignatory) Update(signatory *entity.EntitySignatory) error {
	err := r.db.Save(signatory).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositorySignatory) Delete(signatory *entity.EntitySignatory) error {
	err := r.db.Delete(signatory).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositorySignatory) GetSignatories(filters entity.EntitySignatoryFilters) ([]entity.EntitySignatory, error) {
	var signatories []entity.EntitySignatory

	query := r.db.Model(&entity.EntitySignatory{})

	if len(filters.IDs) > 0 {
		query = query.Where("id IN ?", filters.IDs)
	}

	if filters.EnvelopeID != 0 {
		query = query.Where("envelope_id = ?", filters.EnvelopeID)
	}

	if filters.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filters.Email+"%")
	}

	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	err := query.Order("created_at DESC").Find(&signatories).Error
	if err != nil {
		return nil, err
	}

	return signatories, nil
}

func (r *RepositorySignatory) GetByEnvelopeID(envelopeID int) ([]entity.EntitySignatory, error) {
	var signatories []entity.EntitySignatory

	err := r.db.Where("envelope_id = ?", envelopeID).Order("created_at DESC").Find(&signatories).Error
	if err != nil {
		return nil, err
	}

	return signatories, nil
}

func (r *RepositorySignatory) GetByEmail(email string) (*entity.EntitySignatory, error) {
	var sig entity.EntitySignatory

	err := r.db.Where("email = ?", email).First(&sig).Error
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func (r *RepositorySignatory) GetByEmailAndEnvelopeID(email string, envelopeID int) (*entity.EntitySignatory, error) {
	var sig entity.EntitySignatory

	err := r.db.Where("email = ? AND envelope_id = ?", email, envelopeID).First(&sig).Error
	if err != nil {
		return nil, err
	}

	return &sig, nil
}