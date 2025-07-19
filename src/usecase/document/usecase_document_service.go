package document

import (
	"app/entity"
	"app/infrastructure/clicksign"
	clicksignInterface "app/usecase/clicksign"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type UsecaseDocumentService struct {
	repositoryDocument IRepositoryDocument
	clicksignClient    clicksignInterface.ClicksignClientInterface
	documentService    *clicksign.DocumentService
	logger             *logrus.Logger
}

func NewUsecaseDocumentService(repositoryDocument IRepositoryDocument) IUsecaseDocument {
	return &UsecaseDocumentService{
		repositoryDocument: repositoryDocument,
	}
}

func NewUsecaseDocumentServiceWithClicksign(
	repositoryDocument IRepositoryDocument,
	clicksignClient clicksignInterface.ClicksignClientInterface,
	logger *logrus.Logger,
) IUsecaseDocument {
	documentService := clicksign.NewDocumentService(clicksignClient, logger)

	return &UsecaseDocumentService{
		repositoryDocument: repositoryDocument,
		clicksignClient:    clicksignClient,
		documentService:    documentService,
		logger:             logger,
	}
}

func (u *UsecaseDocumentService) Create(document *entity.EntityDocument) error {
	err := document.Validate()
	if err != nil {
		return fmt.Errorf("document validation failed: %w", err)
	}

	err = u.repositoryDocument.Create(document)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

func (u *UsecaseDocumentService) Update(document *entity.EntityDocument) error {
	err := document.Validate()
	if err != nil {
		return fmt.Errorf("document validation failed: %w", err)
	}

	existingDoc, err := u.repositoryDocument.GetByID(document.ID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	if existingDoc.Status == "sent" {
		return fmt.Errorf("cannot update document in 'sent' status")
	}

	err = u.repositoryDocument.Update(document)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (u *UsecaseDocumentService) Delete(document *entity.EntityDocument) error {
	existingDoc, err := u.repositoryDocument.GetByID(document.ID)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	if existingDoc.Status == "sent" || existingDoc.Status == "processing" {
		return fmt.Errorf("cannot delete document in '%s' status", existingDoc.Status)
	}

	err = u.repositoryDocument.Delete(document)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (u *UsecaseDocumentService) GetDocument(id int) (*entity.EntityDocument, error) {
	document, err := u.repositoryDocument.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	return document, nil
}

func (u *UsecaseDocumentService) GetDocuments(filters entity.EntityDocumentFilters) ([]entity.EntityDocument, error) {
	documents, err := u.repositoryDocument.GetDocuments(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return documents, nil
}

func (u *UsecaseDocumentService) PrepareForSigning(id int) (*entity.EntityDocument, error) {
	document, err := u.repositoryDocument.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	err = document.PrepareForSigning()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare document for signing: %w", err)
	}

	err = u.repositoryDocument.Update(document)
	if err != nil {
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	return document, nil
}

// UploadToClicksign faz upload do documento para o Clicksign
func (u *UsecaseDocumentService) UploadToClicksign(document *entity.EntityDocument) (string, error) {
	if u.documentService == nil {
		return "", fmt.Errorf("clicksign service not configured")
	}

	ctx := context.Background()
	correlationID := fmt.Sprintf("doc_%d_%d", document.ID, document.CreatedAt.Unix())
	ctx = context.WithValue(ctx, "correlation_id", correlationID)

	if u.logger != nil {
		u.logger.WithFields(logrus.Fields{
			"document_id":    document.ID,
			"document_name":  document.Name,
			"is_from_base64": document.IsFromBase64,
			"correlation_id": correlationID,
		}).Info("Starting document upload to Clicksign")
	}

	clicksignDocID, err := u.documentService.UploadDocument(ctx, document)
	if err != nil {
		if u.logger != nil {
			u.logger.WithFields(logrus.Fields{
				"error":          err.Error(),
				"document_id":    document.ID,
				"correlation_id": correlationID,
			}).Error("Failed to upload document to Clicksign")
		}
		return "", fmt.Errorf("failed to upload document to Clicksign: %w", err)
	}

	// Atualizar documento com a chave do Clicksign
	document.SetClicksignKey(clicksignDocID)
	if err := u.repositoryDocument.Update(document); err != nil {
		if u.logger != nil {
			u.logger.WithFields(logrus.Fields{
				"error":             err.Error(),
				"document_id":       document.ID,
				"clicksign_doc_id":  clicksignDocID,
				"correlation_id":    correlationID,
			}).Error("Failed to update document with Clicksign key")
		}
		return "", fmt.Errorf("failed to update document with Clicksign key: %w", err)
	}

	if u.logger != nil {
		u.logger.WithFields(logrus.Fields{
			"document_id":      document.ID,
			"clicksign_doc_id": clicksignDocID,
			"correlation_id":   correlationID,
		}).Info("Document uploaded to Clicksign successfully")
	}

	return clicksignDocID, nil
}
