// Code generated by MockGen. DO NOT EDIT.
// Source: app/usecase/clicksign (interfaces: ClicksignClientInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockClicksignClientInterface is a mock of ClicksignClientInterface interface.
type MockClicksignClientInterface struct {
	ctrl     *gomock.Controller
	recorder *MockClicksignClientInterfaceMockRecorder
}

// MockClicksignClientInterfaceMockRecorder is the mock recorder for MockClicksignClientInterface.
type MockClicksignClientInterfaceMockRecorder struct {
	mock *MockClicksignClientInterface
}

// NewMockClicksignClientInterface creates a new mock instance.
func NewMockClicksignClientInterface(ctrl *gomock.Controller) *MockClicksignClientInterface {
	mock := &MockClicksignClientInterface{ctrl: ctrl}
	mock.recorder = &MockClicksignClientInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClicksignClientInterface) EXPECT() *MockClicksignClientInterfaceMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockClicksignClientInterface) Delete(arg0 context.Context, arg1 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockClicksignClientInterfaceMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockClicksignClientInterface)(nil).Delete), arg0, arg1)
}

// Get mocks base method.
func (m *MockClicksignClientInterface) Get(arg0 context.Context, arg1 string) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockClicksignClientInterfaceMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockClicksignClientInterface)(nil).Get), arg0, arg1)
}

// Patch mocks base method.
func (m *MockClicksignClientInterface) Patch(arg0 context.Context, arg1 string, arg2 interface{}) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Patch", arg0, arg1, arg2)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Patch indicates an expected call of Patch.
func (mr *MockClicksignClientInterfaceMockRecorder) Patch(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Patch", reflect.TypeOf((*MockClicksignClientInterface)(nil).Patch), arg0, arg1, arg2)
}

// Post mocks base method.
func (m *MockClicksignClientInterface) Post(arg0 context.Context, arg1 string, arg2 interface{}) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Post", arg0, arg1, arg2)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Post indicates an expected call of Post.
func (mr *MockClicksignClientInterfaceMockRecorder) Post(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Post", reflect.TypeOf((*MockClicksignClientInterface)(nil).Post), arg0, arg1, arg2)
}

// Put mocks base method.
func (m *MockClicksignClientInterface) Put(arg0 context.Context, arg1 string, arg2 interface{}) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", arg0, arg1, arg2)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Put indicates an expected call of Put.
func (mr *MockClicksignClientInterfaceMockRecorder) Put(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockClicksignClientInterface)(nil).Put), arg0, arg1, arg2)
}
