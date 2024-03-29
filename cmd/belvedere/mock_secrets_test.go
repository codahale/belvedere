// Code generated by MockGen. DO NOT EDIT.
// Source: ../../pkg/belvedere/secrets.go

// Package main is a generated GoMock package.
package main

import (
	context "context"
	io "io"
	reflect "reflect"

	belvedere "github.com/codahale/belvedere/pkg/belvedere"
	gomock "github.com/golang/mock/gomock"
)

// MockSecretsService is a mock of SecretsService interface.
type MockSecretsService struct {
	ctrl     *gomock.Controller
	recorder *MockSecretsServiceMockRecorder
}

// MockSecretsServiceMockRecorder is the mock recorder for MockSecretsService.
type MockSecretsServiceMockRecorder struct {
	mock *MockSecretsService
}

// NewMockSecretsService creates a new mock instance.
func NewMockSecretsService(ctrl *gomock.Controller) *MockSecretsService {
	mock := &MockSecretsService{ctrl: ctrl}
	mock.recorder = &MockSecretsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretsService) EXPECT() *MockSecretsServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSecretsService) Create(ctx context.Context, name string, value io.Reader, dryRun bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, name, value, dryRun)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockSecretsServiceMockRecorder) Create(ctx, name, value, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSecretsService)(nil).Create), ctx, name, value, dryRun)
}

// Delete mocks base method.
func (m *MockSecretsService) Delete(ctx context.Context, name string, dryRun bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, name, dryRun)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockSecretsServiceMockRecorder) Delete(ctx, name, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSecretsService)(nil).Delete), ctx, name, dryRun)
}

// Grant mocks base method.
func (m *MockSecretsService) Grant(ctx context.Context, name, app string, dryRun bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Grant", ctx, name, app, dryRun)
	ret0, _ := ret[0].(error)
	return ret0
}

// Grant indicates an expected call of Grant.
func (mr *MockSecretsServiceMockRecorder) Grant(ctx, name, app, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Grant", reflect.TypeOf((*MockSecretsService)(nil).Grant), ctx, name, app, dryRun)
}

// List mocks base method.
func (m *MockSecretsService) List(ctx context.Context) ([]belvedere.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx)
	ret0, _ := ret[0].([]belvedere.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockSecretsServiceMockRecorder) List(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockSecretsService)(nil).List), ctx)
}

// Revoke mocks base method.
func (m *MockSecretsService) Revoke(ctx context.Context, name, app string, dryRun bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Revoke", ctx, name, app, dryRun)
	ret0, _ := ret[0].(error)
	return ret0
}

// Revoke indicates an expected call of Revoke.
func (mr *MockSecretsServiceMockRecorder) Revoke(ctx, name, app, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Revoke", reflect.TypeOf((*MockSecretsService)(nil).Revoke), ctx, name, app, dryRun)
}

// Update mocks base method.
func (m *MockSecretsService) Update(ctx context.Context, name string, value io.Reader, dryRun bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, name, value, dryRun)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockSecretsServiceMockRecorder) Update(ctx, name, value, dryRun interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSecretsService)(nil).Update), ctx, name, value, dryRun)
}
