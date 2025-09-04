package repository

import (
	"app/entity"

	"gorm.io/gorm"
)

type RepositoryAutoSignatureTerm struct {
	db *gorm.DB
}

func NewRepositoryAutoSignatureTerm(db *gorm.DB) *RepositoryAutoSignatureTerm {
	return &RepositoryAutoSignatureTerm{
		db: db,
	}
}

func (r *RepositoryAutoSignatureTerm) Create(term *entity.EntityAutoSignatureTerm) error {
	return r.db.Create(term).Error
}

func (r *RepositoryAutoSignatureTerm) GetByID(id int) (*entity.EntityAutoSignatureTerm, error) {
	var term entity.EntityAutoSignatureTerm
	err := r.db.First(&term, id).Error
	if err != nil {
		return nil, err
	}
	return &term, nil
}

func (r *RepositoryAutoSignatureTerm) GetByClicksignKey(key string) (*entity.EntityAutoSignatureTerm, error) {
	var term entity.EntityAutoSignatureTerm
	err := r.db.Where("clicksign_key = ?", key).First(&term).Error
	if err != nil {
		return nil, err
	}
	return &term, nil
}

func (r *RepositoryAutoSignatureTerm) Update(term *entity.EntityAutoSignatureTerm) error {
	return r.db.Save(term).Error
}

func (r *RepositoryAutoSignatureTerm) Delete(term *entity.EntityAutoSignatureTerm) error {
	return r.db.Delete(term).Error
}

func (r *RepositoryAutoSignatureTerm) GetAll() ([]entity.EntityAutoSignatureTerm, error) {
	var terms []entity.EntityAutoSignatureTerm
	err := r.db.Find(&terms).Error
	if err != nil {
		return nil, err
	}
	return terms, nil
}
