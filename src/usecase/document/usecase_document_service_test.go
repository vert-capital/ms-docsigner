package document

import (
	"app/entity"
	"app/mocks"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsecaseDocumentService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_document*.pdf")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("test content"))
	require.NoError(t, err)
	tempFile.Close()

	fileInfo, err := os.Stat(tempFile.Name())
	require.NoError(t, err)

	tests := []struct {
		name          string
		document      *entity.EntityDocument
		mockSetup     func()
		expectedError string
	}{
		{
			name: "Successful creation",
			document: &entity.EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "draft",
				Description: "Test description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func() {
				mockRepo.EXPECT().Create(gomock.Any()).Return(nil).Times(1)
			},
			expectedError: "",
		},
		{
			name: "Validation error",
			document: &entity.EntityDocument{
				Name:        "", // Invalid name
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "draft",
				Description: "Test description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func() {
				// No mock setup needed as validation fails before repo call
			},
			expectedError: "document validation failed",
		},
		{
			name: "Repository error",
			document: &entity.EntityDocument{
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "draft",
				Description: "Test description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func() {
				mockRepo.EXPECT().Create(gomock.Any()).Return(errors.New("database error")).Times(1)
			},
			expectedError: "failed to create document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Create(tt.document)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUsecaseDocumentService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_document*.pdf")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte("test content"))
	require.NoError(t, err)
	tempFile.Close()

	fileInfo, err := os.Stat(tempFile.Name())
	require.NoError(t, err)

	validDoc := &entity.EntityDocument{
		ID:          1,
		Name:        "Test Document",
		FilePath:    tempFile.Name(),
		FileSize:    fileInfo.Size(),
		MimeType:    "application/pdf",
		Status:      "draft",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		document      *entity.EntityDocument
		mockSetup     func()
		expectedError string
	}{
		{
			name:     "Successful update",
			document: validDoc,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(1).Return(validDoc, nil).Times(1)
				mockRepo.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
			},
			expectedError: "",
		},
		{
			name: "Document not found",
			document: &entity.EntityDocument{
				ID:          999,
				Name:        "Test Document",
				FilePath:    tempFile.Name(),
				FileSize:    fileInfo.Size(),
				MimeType:    "application/pdf",
				Status:      "draft",
				Description: "Test description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(999).Return(nil, errors.New("not found")).Times(1)
			},
			expectedError: "document not found",
		},
		{
			name:     "Cannot update sent document",
			document: validDoc,
			mockSetup: func() {
				sentDoc := &entity.EntityDocument{
					ID:     1,
					Status: "sent",
				}
				mockRepo.EXPECT().GetByID(1).Return(sentDoc, nil).Times(1)
			},
			expectedError: "cannot update document in 'sent' status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Update(tt.document)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUsecaseDocumentService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	doc := &entity.EntityDocument{
		ID:     1,
		Name:   "Test Document",
		Status: "draft",
	}

	tests := []struct {
		name          string
		document      *entity.EntityDocument
		mockSetup     func()
		expectedError string
	}{
		{
			name:     "Successful deletion",
			document: doc,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(1).Return(doc, nil).Times(1)
				mockRepo.EXPECT().Delete(gomock.Any()).Return(nil).Times(1)
			},
			expectedError: "",
		},
		{
			name:     "Document not found",
			document: doc,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(1).Return(nil, errors.New("not found")).Times(1)
			},
			expectedError: "document not found",
		},
		{
			name:     "Cannot delete sent document",
			document: doc,
			mockSetup: func() {
				sentDoc := &entity.EntityDocument{
					ID:     1,
					Status: "sent",
				}
				mockRepo.EXPECT().GetByID(1).Return(sentDoc, nil).Times(1)
			},
			expectedError: "cannot delete document in 'sent' status",
		},
		{
			name:     "Cannot delete processing document",
			document: doc,
			mockSetup: func() {
				processingDoc := &entity.EntityDocument{
					ID:     1,
					Status: "processing",
				}
				mockRepo.EXPECT().GetByID(1).Return(processingDoc, nil).Times(1)
			},
			expectedError: "cannot delete document in 'processing' status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := service.Delete(tt.document)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUsecaseDocumentService_GetDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	doc := &entity.EntityDocument{
		ID:   1,
		Name: "Test Document",
	}

	tests := []struct {
		name          string
		id            int
		mockSetup     func()
		expectedDoc   *entity.EntityDocument
		expectedError string
	}{
		{
			name: "Successful retrieval",
			id:   1,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(1).Return(doc, nil).Times(1)
			},
			expectedDoc:   doc,
			expectedError: "",
		},
		{
			name: "Document not found",
			id:   999,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(999).Return(nil, errors.New("not found")).Times(1)
			},
			expectedDoc:   nil,
			expectedError: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			doc, err := service.GetDocument(tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDoc, doc)
			}
		})
	}
}

func TestUsecaseDocumentService_GetDocuments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	docs := []entity.EntityDocument{
		{ID: 1, Name: "Doc 1"},
		{ID: 2, Name: "Doc 2"},
	}

	filters := entity.EntityDocumentFilters{
		Search: "test",
		Status: "draft",
	}

	tests := []struct {
		name          string
		filters       entity.EntityDocumentFilters
		mockSetup     func()
		expectedDocs  []entity.EntityDocument
		expectedError string
	}{
		{
			name:    "Successful retrieval",
			filters: filters,
			mockSetup: func() {
				mockRepo.EXPECT().GetDocuments(filters).Return(docs, nil).Times(1)
			},
			expectedDocs:  docs,
			expectedError: "",
		},
		{
			name:    "Repository error",
			filters: filters,
			mockSetup: func() {
				mockRepo.EXPECT().GetDocuments(filters).Return(nil, errors.New("database error")).Times(1)
			},
			expectedDocs:  nil,
			expectedError: "failed to get documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			docs, err := service.GetDocuments(tt.filters)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, docs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDocs, docs)
			}
		})
	}
}

func TestUsecaseDocumentService_PrepareForSigning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockIRepositoryDocument(ctrl)
	service := NewUsecaseDocumentService(mockRepo)

	tests := []struct {
		name          string
		id            int
		mockSetup     func()
		expectedError string
	}{
		{
			name: "Successful preparation",
			id:   1,
			mockSetup: func() {
				doc := &entity.EntityDocument{
					ID:     1,
					Status: "ready",
				}
				mockRepo.EXPECT().GetByID(1).Return(doc, nil).Times(1)
				mockRepo.EXPECT().Update(gomock.Any()).Return(nil).Times(1)
			},
			expectedError: "",
		},
		{
			name: "Document not found",
			id:   999,
			mockSetup: func() {
				mockRepo.EXPECT().GetByID(999).Return(nil, errors.New("not found")).Times(1)
			},
			expectedError: "document not found",
		},
		{
			name: "Invalid status for signing",
			id:   1,
			mockSetup: func() {
				doc := &entity.EntityDocument{
					ID:     1,
					Status: "draft",
				}
				mockRepo.EXPECT().GetByID(1).Return(doc, nil).Times(1)
			},
			expectedError: "failed to prepare document for signing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			doc, err := service.PrepareForSigning(tt.id)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, doc)
				assert.Equal(t, "processing", doc.Status)
			}
		})
	}
}
