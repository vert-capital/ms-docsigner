package envelope

import (
	"app/entity"
	"context"
)

//go:generate mockgen -destination=../../mocks/mock_usecase_repository_envelope.go -package=mocks app/usecase/envelope IRepositoryEnvelope
type IRepositoryEnvelope interface {
	GetByID(id int) (*entity.EntityEnvelope, error)
	Create(envelope *entity.EntityEnvelope) error
	Update(envelope *entity.EntityEnvelope) error
	Delete(envelope *entity.EntityEnvelope) error
	GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error)
	GetByClicksignKey(key string) (*entity.EntityEnvelope, error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_envelope.go -package=mocks app/usecase/envelope IUsecaseEnvelope
type IUsecaseEnvelope interface {
	CreateDocument(ctx context.Context, envelopeID string, document *entity.EntityDocument) (string, error)
	CreateEnvelope(envelope *entity.EntityEnvelope) (*entity.EntityEnvelope, error)
	CreateEnvelopeWithDocuments(envelope *entity.EntityEnvelope, documents []*entity.EntityDocument) (*entity.EntityEnvelope, error)
	CreateEnvelopeWithRequirements(ctx context.Context, envelope *entity.EntityEnvelope, requirements []*entity.EntityRequirement) (*entity.EntityEnvelope, error)
	GetEnvelope(id int) (*entity.EntityEnvelope, error)
	GetEnvelopes(filters entity.EntityEnvelopeFilters) ([]entity.EntityEnvelope, error)
	UpdateEnvelope(envelope *entity.EntityEnvelope) error
	DeleteEnvelope(id int) error
	ActivateEnvelope(id int) (*entity.EntityEnvelope, error)
}
