// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/codahale/belvedere/pkg/belvedere/internal/setup (interfaces: Service)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	dns "google.golang.org/api/dns/v1"
	reflect "reflect"
	time "time"
)

// SetupService is a mock of Service interface
type SetupService struct {
	ctrl     *gomock.Controller
	recorder *SetupServiceMockRecorder
}

// SetupServiceMockRecorder is the mock recorder for SetupService
type SetupServiceMockRecorder struct {
	mock *SetupService
}

// NewSetupService creates a new mock instance
func NewSetupService(ctrl *gomock.Controller) *SetupService {
	mock := &SetupService{ctrl: ctrl}
	mock.recorder = &SetupServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *SetupService) EXPECT() *SetupServiceMockRecorder {
	return m.recorder
}

// EnableAPIs mocks base method
func (m *SetupService) EnableAPIs(arg0 context.Context, arg1 string, arg2 bool, arg3 time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableAPIs", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableAPIs indicates an expected call of EnableAPIs
func (mr *SetupServiceMockRecorder) EnableAPIs(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableAPIs", reflect.TypeOf((*SetupService)(nil).EnableAPIs), arg0, arg1, arg2, arg3)
}

// ManagedZone mocks base method
func (m *SetupService) ManagedZone(arg0 context.Context, arg1 string) (*dns.ManagedZone, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ManagedZone", arg0, arg1)
	ret0, _ := ret[0].(*dns.ManagedZone)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ManagedZone indicates an expected call of ManagedZone
func (mr *SetupServiceMockRecorder) ManagedZone(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ManagedZone", reflect.TypeOf((*SetupService)(nil).ManagedZone), arg0, arg1)
}

// SetDMPerms mocks base method
func (m *SetupService) SetDMPerms(arg0 context.Context, arg1 string, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDMPerms", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDMPerms indicates an expected call of SetDMPerms
func (mr *SetupServiceMockRecorder) SetDMPerms(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDMPerms", reflect.TypeOf((*SetupService)(nil).SetDMPerms), arg0, arg1, arg2)
}
