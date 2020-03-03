// Code generated by MockGen. DO NOT EDIT.
// Source: file_reader.go

// Package main is a generated GoMock package.
package main

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockFileReader is a mock of FileReader interface
type MockFileReader struct {
	ctrl     *gomock.Controller
	recorder *MockFileReaderMockRecorder
}

// MockFileReaderMockRecorder is the mock recorder for MockFileReader
type MockFileReaderMockRecorder struct {
	mock *MockFileReader
}

// NewMockFileReader creates a new mock instance
func NewMockFileReader(ctrl *gomock.Controller) *MockFileReader {
	mock := &MockFileReader{ctrl: ctrl}
	mock.recorder = &MockFileReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFileReader) EXPECT() *MockFileReaderMockRecorder {
	return m.recorder
}

// Read mocks base method
func (m *MockFileReader) Read(path string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", path)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockFileReaderMockRecorder) Read(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockFileReader)(nil).Read), path)
}
