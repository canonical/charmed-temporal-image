// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/canonical/charmed-temporal-image/temporal-server/authorizer (interfaces: NamespaceAccessProvider,TokenVerifier)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	authorizer "github.com/canonical/charmed-temporal-image/temporal-server/authorizer"
	gomock "go.uber.org/mock/gomock"
)

// MockNamespaceAccessProvider is a mock of NamespaceAccessProvider interface.
type MockNamespaceAccessProvider struct {
	ctrl     *gomock.Controller
	recorder *MockNamespaceAccessProviderMockRecorder
}

// MockNamespaceAccessProviderMockRecorder is the mock recorder for MockNamespaceAccessProvider.
type MockNamespaceAccessProviderMockRecorder struct {
	mock *MockNamespaceAccessProvider
}

// NewMockNamespaceAccessProvider creates a new mock instance.
func NewMockNamespaceAccessProvider(ctrl *gomock.Controller) *MockNamespaceAccessProvider {
	mock := &MockNamespaceAccessProvider{ctrl: ctrl}
	mock.recorder = &MockNamespaceAccessProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNamespaceAccessProvider) EXPECT() *MockNamespaceAccessProviderMockRecorder {
	return m.recorder
}

// GetNamespaceAccessInformation mocks base method.
func (m *MockNamespaceAccessProvider) GetNamespaceAccessInformation(arg0 context.Context, arg1 string) ([]authorizer.NamespaceAccess, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespaceAccessInformation", arg0, arg1)
	ret0, _ := ret[0].([]authorizer.NamespaceAccess)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNamespaceAccessInformation indicates an expected call of GetNamespaceAccessInformation.
func (mr *MockNamespaceAccessProviderMockRecorder) GetNamespaceAccessInformation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespaceAccessInformation", reflect.TypeOf((*MockNamespaceAccessProvider)(nil).GetNamespaceAccessInformation), arg0, arg1)
}

// MockTokenVerifier is a mock of TokenVerifier interface.
type MockTokenVerifier struct {
	ctrl     *gomock.Controller
	recorder *MockTokenVerifierMockRecorder
}

// MockTokenVerifierMockRecorder is the mock recorder for MockTokenVerifier.
type MockTokenVerifierMockRecorder struct {
	mock *MockTokenVerifier
}

// NewMockTokenVerifier creates a new mock instance.
func NewMockTokenVerifier(ctrl *gomock.Controller) *MockTokenVerifier {
	mock := &MockTokenVerifier{ctrl: ctrl}
	mock.recorder = &MockTokenVerifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenVerifier) EXPECT() *MockTokenVerifierMockRecorder {
	return m.recorder
}

// GetTokenInfo mocks base method.
func (m *MockTokenVerifier) GetTokenInfo(arg0 string) (*authorizer.TokenInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTokenInfo", arg0)
	ret0, _ := ret[0].(*authorizer.TokenInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTokenInfo indicates an expected call of GetTokenInfo.
func (mr *MockTokenVerifierMockRecorder) GetTokenInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTokenInfo", reflect.TypeOf((*MockTokenVerifier)(nil).GetTokenInfo), arg0)
}

// VerifyToken mocks base method.
func (m *MockTokenVerifier) VerifyToken(arg0 *authorizer.TokenInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyToken", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyToken indicates an expected call of VerifyToken.
func (mr *MockTokenVerifierMockRecorder) VerifyToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyToken", reflect.TypeOf((*MockTokenVerifier)(nil).VerifyToken), arg0)
}
