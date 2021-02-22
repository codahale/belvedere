// Code generated by MockGen. DO NOT EDIT.
// Source: internal/check/health.go

// Package belvedere is a generated GoMock package.
package belvedere

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockHealthChecker is a mock of HealthChecker interface.
type MockHealthChecker struct {
	ctrl     *gomock.Controller
	recorder *MockHealthCheckerMockRecorder
}

// MockHealthCheckerMockRecorder is the mock recorder for MockHealthChecker.
type MockHealthCheckerMockRecorder struct {
	mock *MockHealthChecker
}

// NewMockHealthChecker creates a new mock instance.
func NewMockHealthChecker(ctrl *gomock.Controller) *MockHealthChecker {
	mock := &MockHealthChecker{ctrl: ctrl}
	mock.recorder = &MockHealthCheckerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHealthChecker) EXPECT() *MockHealthCheckerMockRecorder {
	return m.recorder
}

// Poll mocks base method.
func (m *MockHealthChecker) Poll(ctx context.Context, project, region, backendService, instanceGroup string, interval time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Poll", ctx, project, region, backendService, instanceGroup, interval)
	ret0, _ := ret[0].(error)
	return ret0
}

// Poll indicates an expected call of Poll.
func (mr *MockHealthCheckerMockRecorder) Poll(ctx, project, region, backendService, instanceGroup, interval interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Poll", reflect.TypeOf((*MockHealthChecker)(nil).Poll), ctx, project, region, backendService, instanceGroup, interval)
}
