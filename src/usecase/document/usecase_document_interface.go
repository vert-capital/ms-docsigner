package document

import "app/entity"

//go:generate mockgen -destination=../../mocks/mock_usecase_repository_document.go -package=mocks app/usecase/document IRepositoryDocument
type IRepositoryDocument interface {
	GetByID(id int) (*entity.EntityDocument, error)
	Create(document *entity.EntityDocument) error
	Update(document *entity.EntityDocument) error
	Delete(document *entity.EntityDocument) error
	GetDocuments(filters entity.EntityDocumentFilters) ([]entity.EntityDocument, error)
	GetByClicksignKey(key string) (*entity.EntityDocument, error)
}

//go:generate mockgen -destination=../../mocks/mock_usecase_document.go -package=mocks app/usecase/document IUsecaseDocument
type IUsecaseDocument interface {
	Create(document *entity.EntityDocument) error
	Update(document *entity.EntityDocument) error
	Delete(document *entity.EntityDocument) error
	GetDocument(id int) (*entity.EntityDocument, error)
	GetDocuments(filters entity.EntityDocumentFilters) ([]entity.EntityDocument, error)
	PrepareForSigning(id int) (*entity.EntityDocument, error)
}
