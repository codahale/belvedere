// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/codahale/belvedere/pkg/belvedere/internal/deployments (interfaces: Manager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	deployments "github.com/codahale/belvedere/pkg/belvedere/internal/deployments"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// DeploymentsManager is a mock of Manager interface
type DeploymentsManager struct {
	ctrl     *gomock.Controller
	recorder *DeploymentsManagerMockRecorder
}

// DeploymentsManagerMockRecorder is the mock recorder for DeploymentsManager
type DeploymentsManagerMockRecorder struct {
	mock *DeploymentsManager
}

// NewDeploymentsManager creates a new mock instance
func NewDeploymentsManager(ctrl *gomock.Controller) *DeploymentsManager {
	mock := &DeploymentsManager{ctrl: ctrl}
	mock.recorder = &DeploymentsManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *DeploymentsManager) EXPECT() *DeploymentsManagerMockRecorder {
	return m.recorder
}

// Delete mocks base method
func (m *DeploymentsManager) Delete(arg0 context.Context, arg1, arg2 string, arg3, arg4 bool, arg5 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *DeploymentsManagerMockRecorder) Delete(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*DeploymentsManager)(nil).Delete), arg0, arg1, arg2, arg3, arg4, arg5)
}

// Get mocks base method
func (m *DeploymentsManager) Get(arg0 context.Context, arg1, arg2 string) (*deployments.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2)
	ret0, _ := ret[0].(*deployments.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *DeploymentsManagerMockRecorder) Get(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*DeploymentsManager)(nil).Get), arg0, arg1, arg2)
}

// Insert mocks base method
func (m *DeploymentsManager) Insert(arg0 context.Context, arg1, arg2 string, arg3 []deployments.Resource, arg4 deployments.Labels, arg5 bool, arg6 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert
func (mr *DeploymentsManagerMockRecorder) Insert(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*DeploymentsManager)(nil).Insert), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// List mocks base method
func (m *DeploymentsManager) List(arg0 context.Context, arg1, arg2 string) ([]deployments.Deployment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]deployments.Deployment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *DeploymentsManagerMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*DeploymentsManager)(nil).List), arg0, arg1, arg2)
}

// Update mocks base method
func (m *DeploymentsManager) Update(arg0 context.Context, arg1, arg2 string, arg3 []deployments.Resource, arg4 bool, arg5 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *DeploymentsManagerMockRecorder) Update(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*DeploymentsManager)(nil).Update), arg0, arg1, arg2, arg3, arg4, arg5)
}