package requirement

import (
	"app/entity"
	"context"
)

//go:generate mockgen -destination=../../mocks/mock_repository_requirement.go -package=mocks app/usecase/requirement IRepositoryRequirement
type IRepositoryRequirement interface {
	Create(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error)
	GetByEnvelopeID(ctx context.Context, envelopeID int) ([]entity.EntityRequirement, error)
	GetByID(ctx context.Context, id int) (*entity.EntityRequirement, error)
	Update(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error)
	Delete(ctx context.Context, requirement *entity.EntityRequirement) error
	GetByClicksignKey(ctx context.Context, key string) (*entity.EntityRequirement, error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_requirement.go -package=mocks app/usecase/requirement IUsecaseRequirement
type IUsecaseRequirement interface {
	CreateRequirement(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error)
	GetRequirementsByEnvelopeID(ctx context.Context, envelopeID int) ([]entity.EntityRequirement, error)
	GetRequirement(ctx context.Context, id int) (*entity.EntityRequirement, error)
	UpdateRequirement(ctx context.Context, requirement *entity.EntityRequirement) (*entity.EntityRequirement, error)
	DeleteRequirement(ctx context.Context, id int) error
}