// Code generated by MockGen. DO NOT EDIT.
// Source: internal/deployments/deployments.go

// Package belvedere is a generated GoMock package.
package belvedere

import (
	context "context"
	reflect "reflect"
	time "time"

	deployments "github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	gomock "github.com/golang/mock/gomock"
)

// DeploymentsManager is a mock of Manager interface.
type DeploymentsManager struct {
	ctrl     *gomock.Controller
	recorder *DeploymentsManagerMockRecorder
}

// DeploymentsManagerMockRecorder is the mock recorder for DeploymentsManager.
type DeploymentsManagerMockRecorder struct {
	mock *DeploymentsManager
}

// NewDeploymentsManager creates a new mock instance.
func NewDeploymentsManager(ctrl *gomock.Controller) *DeploymentsManager {
	mock := &DeploymentsManager{ctrl: ctrl}
	mock.recorder = &DeploymentsManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *DeploymentsManager) EXPECT() *DeploymentsManagerMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *DeploymentsManager) Delete(ctx context.Context, project, name string, dryRun, async bool, interval time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, project, name, dryRun, async, interval)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *DeploymentsManagerMockRecorder) Delete(ctx, project, name, dryRun, async, interval interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*DeploymentsManager)(nil).Delete), ctx, project, name, dryRun, async, interval)
}

// Get mocks base method.
func (m *DeploymentsManager) Get(ctx context.Context, project, name string) (*deployments.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, project, name)
	ret0, _ := ret[0].(*deployments.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *DeploymentsManagerMockRecorder) Get(ctx, project, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*DeploymentsManager)(nil).Get), ctx, project, name)
}

// Insert mocks base method.
func (m *DeploymentsManager) Insert(ctx context.Context, project, name string, resources []deployments.Resource, labels deployments.Labels, dryRun bool, interval time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, project, name, resources, labels, dryRun, interval)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *DeploymentsManagerMockRecorder) Insert(ctx, project, name, resources, labels, dryRun, interval interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*DeploymentsManager)(nil).Insert), ctx, project, name, resources, labels, dryRun, interval)
}

// List mocks base method.
func (m *DeploymentsManager) List(ctx context.Context, project, filter string) ([]deployments.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, project, filter)
	ret0, _ := ret[0].([]deployments.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *DeploymentsManagerMockRecorder) List(ctx, project, filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*DeploymentsManager)(nil).List), ctx, project, filter)
}

// Update mocks base method.
func (m *DeploymentsManager) Update(ctx context.Context, project, name string, resources []deployments.Resource, dryRun bool, interval time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, project, name, resources, dryRun, interval)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *DeploymentsManagerMockRecorder) Update(ctx, project, name, resources, dryRun, interval interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*DeploymentsManager)(nil).Update), ctx, project, name, resources, dryRun, interval)
}
