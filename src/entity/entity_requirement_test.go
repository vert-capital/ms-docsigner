package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRequirement_Success(t *testing.T) {
	auth := "email"
	docID := "doc123"
	signerID := "signer123"

	param := EntityRequirement{
		EnvelopeID: 1,
		Action:     "sign",
		Auth:       &auth,
		DocumentID: &docID,
		SignerID:   &signerID,
	}

	requirement, err := NewRequirement(param)

	assert.NoError(t, err)
	assert.NotNil(t, requirement)
	assert.Equal(t, 1, requirement.EnvelopeID)
	assert.Equal(t, "sign", requirement.Action)
	assert.Equal(t, "sign", requirement.Role) // default value
	assert.Equal(t, "pending", requirement.Status) // default value
	assert.Equal(t, &auth, requirement.Auth)
	assert.Equal(t, &docID, requirement.DocumentID)
	assert.Equal(t, &signerID, requirement.SignerID)
	assert.False(t, requirement.CreatedAt.IsZero())
	assert.False(t, requirement.UpdatedAt.IsZero())
}

func TestNewRequirement_WithDefaults(t *testing.T) {
	param := EntityRequirement{
		EnvelopeID: 1,
		Action:     "agree",
	}

	requirement, err := NewRequirement(param)

	assert.NoError(t, err)
	assert.Equal(t, "sign", requirement.Role)
	assert.Equal(t, "pending", requirement.Status)
}

func TestNewRequirement_ValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		param EntityRequirement
	}{
		{
			name: "missing envelope_id",
			param: EntityRequirement{
				Action: "sign",
			},
		},
		{
			name: "missing action",
			param: EntityRequirement{
				EnvelopeID: 1,
			},
		},
		{
			name: "invalid action",
			param: EntityRequirement{
				EnvelopeID: 1,
				Action:     "invalid_action",
			},
		},
		{
			name: "invalid role",
			param: EntityRequirement{
				EnvelopeID: 1,
				Action:     "sign",
				Role:       "invalid_role",
			},
		},
		{
			name: "invalid auth",
			param: EntityRequirement{
				EnvelopeID: 1,
				Action:     "sign",
				Auth:       stringPtr("invalid_auth"),
			},
		},
		{
			name: "invalid status",
			param: EntityRequirement{
				EnvelopeID: 1,
				Action:     "sign",
				Status:     "invalid_status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requirement, err := NewRequirement(tt.param)
			assert.Error(t, err)
			assert.Nil(t, requirement)
		})
	}
}

func TestRequirement_BusinessRules(t *testing.T) {
	t.Run("provide_evidence requires auth", func(t *testing.T) {
		param := EntityRequirement{
			EnvelopeID: 1,
			Action:     "provide_evidence",
		}

		requirement, err := NewRequirement(param)
		assert.Error(t, err)
		assert.Nil(t, requirement)
		assert.Contains(t, err.Error(), "auth is required for action 'provide_evidence'")
	})

	t.Run("provide_evidence with empty auth", func(t *testing.T) {
		emptyAuth := ""
		param := EntityRequirement{
			EnvelopeID: 1,
			Action:     "provide_evidence",
			Auth:       &emptyAuth,
		}

		requirement, err := NewRequirement(param)
		assert.Error(t, err)
		assert.Nil(t, requirement)
	})

	t.Run("provide_evidence with valid auth", func(t *testing.T) {
		auth := "icp_brasil"
		param := EntityRequirement{
			EnvelopeID: 1,
			Action:     "provide_evidence",
			Auth:       &auth,
		}

		requirement, err := NewRequirement(param)
		assert.NoError(t, err)
		assert.NotNil(t, requirement)
	})
}

func TestRequirement_SetStatus(t *testing.T) {
	requirement := &EntityRequirement{
		Status:    "pending",
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := requirement.UpdatedAt

	err := requirement.SetStatus("completed")
	assert.NoError(t, err)
	assert.Equal(t, "completed", requirement.Status)
	assert.True(t, requirement.UpdatedAt.After(oldUpdatedAt))
}

func TestRequirement_SetStatus_Invalid(t *testing.T) {
	requirement := &EntityRequirement{Status: "pending"}

	err := requirement.SetStatus("invalid_status")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status: invalid_status")
	assert.Equal(t, "pending", requirement.Status)
}

func TestRequirement_SetClicksignKey(t *testing.T) {
	requirement := &EntityRequirement{
		ClicksignKey: "",
		UpdatedAt:    time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := requirement.UpdatedAt

	requirement.SetClicksignKey("test-key-123")
	assert.Equal(t, "test-key-123", requirement.ClicksignKey)
	assert.True(t, requirement.UpdatedAt.After(oldUpdatedAt))
}

func TestRequirement_Complete(t *testing.T) {
	requirement := &EntityRequirement{
		Status:    "pending",
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := requirement.UpdatedAt

	err := requirement.Complete()
	assert.NoError(t, err)
	assert.Equal(t, "completed", requirement.Status)
	assert.True(t, requirement.UpdatedAt.After(oldUpdatedAt))
}

func TestRequirement_Complete_InvalidStatus(t *testing.T) {
	requirement := &EntityRequirement{Status: "completed"}

	err := requirement.Complete()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requirement must be in 'pending' status to complete")
}

func TestRequirement_IsValidAction(t *testing.T) {
	requirement := &EntityRequirement{}

	assert.True(t, requirement.IsValidAction("agree"))
	assert.True(t, requirement.IsValidAction("sign"))
	assert.True(t, requirement.IsValidAction("provide_evidence"))
	assert.False(t, requirement.IsValidAction("invalid_action"))
	assert.False(t, requirement.IsValidAction(""))
}

func TestRequirement_IsValidAuth(t *testing.T) {
	requirement := &EntityRequirement{}

	assert.True(t, requirement.IsValidAuth("email"))
	assert.True(t, requirement.IsValidAuth("icp_brasil"))
	assert.False(t, requirement.IsValidAuth("invalid_auth"))
	assert.False(t, requirement.IsValidAuth(""))
}

func TestRequirement_TableName(t *testing.T) {
	requirement := EntityRequirement{}
	assert.Equal(t, "requirements", requirement.TableName())
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}