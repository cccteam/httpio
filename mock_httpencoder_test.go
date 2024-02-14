// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cccteam/httpio (interfaces: HTTPEncoder)
//
// Generated by this command:
//
//	mockgen -package httpio -destination mock_httpencoder_test.go github.com/cccteam/httpio HTTPEncoder
//

// Package httpio is a generated GoMock package.
package httpio

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockHTTPEncoder is a mock of HTTPEncoder interface.
type MockHTTPEncoder struct {
	ctrl     *gomock.Controller
	recorder *MockHTTPEncoderMockRecorder
}

// MockHTTPEncoderMockRecorder is the mock recorder for MockHTTPEncoder.
type MockHTTPEncoderMockRecorder struct {
	mock *MockHTTPEncoder
}

// NewMockHTTPEncoder creates a new mock instance.
func NewMockHTTPEncoder(ctrl *gomock.Controller) *MockHTTPEncoder {
	mock := &MockHTTPEncoder{ctrl: ctrl}
	mock.recorder = &MockHTTPEncoderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHTTPEncoder) EXPECT() *MockHTTPEncoderMockRecorder {
	return m.recorder
}

// Encode mocks base method.
func (m *MockHTTPEncoder) Encode(arg0 any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encode", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Encode indicates an expected call of Encode.
func (mr *MockHTTPEncoderMockRecorder) Encode(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encode", reflect.TypeOf((*MockHTTPEncoder)(nil).Encode), arg0)
}
