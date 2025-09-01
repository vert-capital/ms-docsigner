package auto_signature_term

import (
	"app/entity"
	"context"
)

//go:generate mockgen -destination=../../mocks/mock_usecase_repository_auto_signature_term.go -package=mocks app/usecase/auto_signature_term IRepositoryAutoSignatureTerm
type IRepositoryAutoSignatureTerm interface {
	Create(term *entity.EntityAutoSignatureTerm) error
	GetByID(id int) (*entity.EntityAutoSignatureTerm, error)
	GetByClicksignKey(key string) (*entity.EntityAutoSignatureTerm, error)
	Update(term *entity.EntityAutoSignatureTerm) error
	Delete(term *entity.EntityAutoSignatureTerm) error
	GetAll() ([]entity.EntityAutoSignatureTerm, error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_auto_signature_term.go -package=mocks app/usecase/auto_signature_term IUsecaseAutoSignatureTerm
type IUsecaseAutoSignatureTerm interface {
	CreateAutoSignatureTerm(ctx context.Context, term *entity.EntityAutoSignatureTerm) (*entity.EntityAutoSignatureTerm, error)
	GetAutoSignatureTerm(id int) (*entity.EntityAutoSignatureTerm, error)
	GetAutoSignatureTermByClicksignKey(key string) (*entity.EntityAutoSignatureTerm, error)
	GetAllAutoSignatureTerms() ([]entity.EntityAutoSignatureTerm, error)
	UpdateAutoSignatureTerm(term *entity.EntityAutoSignatureTerm) error
	DeleteAutoSignatureTerm(id int) error
}
