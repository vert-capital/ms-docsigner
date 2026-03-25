package vertc_assinaturas

import (
	"testing"
	"time"

	"app/entity"
	"app/infrastructure/provider"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuickSendService_mapToQuickSendRequest(t *testing.T) {
	service := NewQuickSendService(nil, logrus.New())
	deadline := time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC)

	request, err := service.mapToQuickSendRequest(QuickSendData{
		Envelope: &entity.EntityEnvelope{
			Name:        "Envelope VertSign",
			Description: "Mensagem do envelope",
			DeadlineAt:  &deadline,
		},
		Documents: []*entity.EntityDocument{
			{
				Name:     "Contrato.pdf",
				FilePath: "https://example.com/contrato.pdf",
				MimeType: "application/pdf",
			},
		},
		Signers: []provider.SignerData{
			{
				Name:       "Assinante Auto",
				Email:      "auto@empresa.com",
				Refusable:  false,
				AuthMethod: "auto_signature",
			},
			{
				Name:       "Assinante Email",
				Email:      "email@empresa.com",
				Refusable:  false,
				AuthMethod: "email",
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, request)
	require.Len(t, request.Signers, 2)

	assert.Equal(t, []string{"automatic_signature"}, request.Signers[0].RequiredMethods)
	assert.Equal(t, []string{"code_email"}, request.Signers[1].RequiredMethods)
}
