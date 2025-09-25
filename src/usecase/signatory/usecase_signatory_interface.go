package signatory

import "app/entity"

//go:generate mockgen -destination=../../mocks/mock_usecase_repository_signatory.go -package=mocks app/usecase/signatory IRepositorySignatory
type IRepositorySignatory interface {
	GetByID(id int) (*entity.EntitySignatory, error)
	Create(signatory *entity.EntitySignatory) error
	Update(signatory *entity.EntitySignatory) error
	Delete(signatory *entity.EntitySignatory) error
	GetSignatories(filters entity.EntitySignatoryFilters) ([]entity.EntitySignatory, error)
	GetByEnvelopeID(envelopeID int) ([]entity.EntitySignatory, error)
	GetByEmail(email string) (*entity.EntitySignatory, error)
	GetByEmailAndEnvelopeID(email string, envelopeID int) (*entity.EntitySignatory, error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_signatory.go -package=mocks app/usecase/signatory IUsecaseSignatory
type IUsecaseSignatory interface {
	CreateSignatory(signatory *entity.EntitySignatory) (*entity.EntitySignatory, error)
	GetSignatory(id int) (*entity.EntitySignatory, error)
	GetSignatories(filters entity.EntitySignatoryFilters) ([]entity.EntitySignatory, error)
	GetSignatoriesByEnvelope(envelopeID int) ([]entity.EntitySignatory, error)
	UpdateSignatory(signatory *entity.EntitySignatory) error
	DeleteSignatory(id int) error
	AssociateToEnvelope(signatoryID int, envelopeID int) error
}
