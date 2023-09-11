// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/grafana/grafana/pkg/services/screenshot (interfaces: CacheService)

// Package screenshot is a generated GoMock package.
package screenshot

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCacheService is a mock of CacheService interface.
type MockCacheService struct {
	ctrl     *gomock.Controller
	recorder *MockCacheServiceMockRecorder
}

// MockCacheServiceMockRecorder is the mock recorder for MockCacheService.
type MockCacheServiceMockRecorder struct {
	mock *MockCacheService
}

// NewMockCacheService creates a new mock instance.
func NewMockCacheService(ctrl *gomock.Controller) *MockCacheService {
	mock := &MockCacheService{ctrl: ctrl}
	mock.recorder = &MockCacheServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheService) EXPECT() *MockCacheServiceMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockCacheService) Get(arg0 context.Context, arg1 ScreenshotOptions) (*Screenshot, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*Screenshot)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockCacheServiceMockRecorder) Get(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCacheService)(nil).Get), arg0, arg1)
}

// Set mocks base method.
func (m *MockCacheService) Set(arg0 context.Context, arg1 ScreenshotOptions, arg2 *Screenshot) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockCacheServiceMockRecorder) Set(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockCacheService)(nil).Set), arg0, arg1, arg2)
}