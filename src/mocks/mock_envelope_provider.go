// Code generated manually for testing. DO NOT EDIT.
// Source: app/infrastructure/provider (interfaces: EnvelopeProvider)

package mocks

import (
	"context"
	"reflect"

	"app/entity"
	"app/infrastructure/provider"

	"github.com/golang/mock/gomock"
)

// MockEnvelopeProvider is a mock of EnvelopeProvider interface.
type MockEnvelopeProvider struct {
	ctrl     *gomock.Controller
	recorder *MockEnvelopeProviderMockRecorder
}

// MockEnvelopeProviderMockRecorder is the mock recorder for MockEnvelopeProvider.
type MockEnvelopeProviderMockRecorder struct {
	mock *MockEnvelopeProvider
}

// NewMockEnvelopeProvider creates a new mock instance.
func NewMockEnvelopeProvider(ctrl *gomock.Controller) *MockEnvelopeProvider {
	mock := &MockEnvelopeProvider{ctrl: ctrl}
	mock.recorder = &MockEnvelopeProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEnvelopeProvider) EXPECT() *MockEnvelopeProviderMockRecorder {
	return m.recorder
}

// CreateEnvelope mocks base method.
func (m *MockEnvelopeProvider) CreateEnvelope(ctx context.Context, envelope *entity.EntityEnvelope) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEnvelope", ctx, envelope)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateEnvelope indicates an expected call of CreateEnvelope.
func (mr *MockEnvelopeProviderMockRecorder) CreateEnvelope(ctx, envelope interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEnvelope", reflect.TypeOf((*MockEnvelopeProvider)(nil).CreateEnvelope), ctx, envelope)
}

// CreateDocument mocks base method.
func (m *MockEnvelopeProvider) CreateDocument(ctx context.Context, envelopeKey string, document *entity.EntityDocument, internalEnvelopeID int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDocument", ctx, envelopeKey, document, internalEnvelopeID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDocument indicates an expected call of CreateDocument.
func (mr *MockEnvelopeProviderMockRecorder) CreateDocument(ctx, envelopeKey, document, internalEnvelopeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDocument", reflect.TypeOf((*MockEnvelopeProvider)(nil).CreateDocument), ctx, envelopeKey, document, internalEnvelopeID)
}

// CreateSigner mocks base method.
func (m *MockEnvelopeProvider) CreateSigner(ctx context.Context, envelopeKey string, signerData provider.SignerData) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSigner", ctx, envelopeKey, signerData)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSigner indicates an expected call of CreateSigner.
func (mr *MockEnvelopeProviderMockRecorder) CreateSigner(ctx, envelopeKey, signerData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSigner", reflect.TypeOf((*MockEnvelopeProvider)(nil).CreateSigner), ctx, envelopeKey, signerData)
}

// CreateRequirement mocks base method.
func (m *MockEnvelopeProvider) CreateRequirement(ctx context.Context, envelopeKey string, reqData provider.RequirementData) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRequirement", ctx, envelopeKey, reqData)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateRequirement indicates an expected call of CreateRequirement.
func (mr *MockEnvelopeProviderMockRecorder) CreateRequirement(ctx, envelopeKey, reqData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRequirement", reflect.TypeOf((*MockEnvelopeProvider)(nil).CreateRequirement), ctx, envelopeKey, reqData)
}

// ActivateEnvelope mocks base method.
func (m *MockEnvelopeProvider) ActivateEnvelope(ctx context.Context, envelopeKey string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ActivateEnvelope", ctx, envelopeKey)
	ret0, _ := ret[0].(error)
	return ret0
}

// ActivateEnvelope indicates an expected call of ActivateEnvelope.
func (mr *MockEnvelopeProviderMockRecorder) ActivateEnvelope(ctx, envelopeKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ActivateEnvelope", reflect.TypeOf((*MockEnvelopeProvider)(nil).ActivateEnvelope), ctx, envelopeKey)
}

// NotifyEnvelope mocks base method.
func (m *MockEnvelopeProvider) NotifyEnvelope(ctx context.Context, envelopeKey string, message string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NotifyEnvelope", ctx, envelopeKey, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// NotifyEnvelope indicates an expected call of NotifyEnvelope.
func (mr *MockEnvelopeProviderMockRecorder) NotifyEnvelope(ctx, envelopeKey, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyEnvelope", reflect.TypeOf((*MockEnvelopeProvider)(nil).NotifyEnvelope), ctx, envelopeKey, message)
}

