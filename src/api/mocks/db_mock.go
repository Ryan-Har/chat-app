// Code generated by MockGen. DO NOT EDIT.
// Source: db.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockDBQueryHandler is a mock of DBQueryHandler interface.
type MockDBQueryHandler struct {
	ctrl     *gomock.Controller
	recorder *MockDBQueryHandlerMockRecorder
}

// MockDBQueryHandlerMockRecorder is the mock recorder for MockDBQueryHandler.
type MockDBQueryHandlerMockRecorder struct {
	mock *MockDBQueryHandler
}

// NewMockDBQueryHandler creates a new mock instance.
func NewMockDBQueryHandler(ctrl *gomock.Controller) *MockDBQueryHandler {
	mock := &MockDBQueryHandler{ctrl: ctrl}
	mock.recorder = &MockDBQueryHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBQueryHandler) EXPECT() *MockDBQueryHandlerMockRecorder {
	return m.recorder
}

// AddExternalUser mocks base method.
func (m *MockDBQueryHandler) AddExternalUser(name, ip string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddExternalUser", name, ip)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddExternalUser indicates an expected call of AddExternalUser.
func (mr *MockDBQueryHandlerMockRecorder) AddExternalUser(name, ip interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddExternalUser", reflect.TypeOf((*MockDBQueryHandler)(nil).AddExternalUser), name, ip)
}

// AddInternalUser mocks base method.
func (m *MockDBQueryHandler) AddInternalUser(roleID int64, firstname, surname, email, password string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddInternalUser", roleID, firstname, surname, email, password)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddInternalUser indicates an expected call of AddInternalUser.
func (mr *MockDBQueryHandlerMockRecorder) AddInternalUser(roleID, firstname, surname, email, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddInternalUser", reflect.TypeOf((*MockDBQueryHandler)(nil).AddInternalUser), roleID, firstname, surname, email, password)
}

// AddMessageByUUID mocks base method.
func (m *MockDBQueryHandler) AddMessageByUUID(uuid string, userid int64, message, time string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMessageByUUID", uuid, userid, message, time)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddMessageByUUID indicates an expected call of AddMessageByUUID.
func (mr *MockDBQueryHandlerMockRecorder) AddMessageByUUID(uuid, userid, message, time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMessageByUUID", reflect.TypeOf((*MockDBQueryHandler)(nil).AddMessageByUUID), uuid, userid, message, time)
}

// ChatEnd mocks base method.
func (m *MockDBQueryHandler) ChatEnd(uuid, endTime string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChatEnd", uuid, endTime)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChatEnd indicates an expected call of ChatEnd.
func (mr *MockDBQueryHandlerMockRecorder) ChatEnd(uuid, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChatEnd", reflect.TypeOf((*MockDBQueryHandler)(nil).ChatEnd), uuid, endTime)
}

// ChatStart mocks base method.
func (m *MockDBQueryHandler) ChatStart(uuid, startTime string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChatStart", uuid, startTime)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChatStart indicates an expected call of ChatStart.
func (mr *MockDBQueryHandlerMockRecorder) ChatStart(uuid, startTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChatStart", reflect.TypeOf((*MockDBQueryHandler)(nil).ChatStart), uuid, startTime)
}

// GetAllChatsInProgress mocks base method.
func (m *MockDBQueryHandler) GetAllChatsInProgress() ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllChatsInProgress")
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllChatsInProgress indicates an expected call of GetAllChatsInProgress.
func (mr *MockDBQueryHandlerMockRecorder) GetAllChatsInProgress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllChatsInProgress", reflect.TypeOf((*MockDBQueryHandler)(nil).GetAllChatsInProgress))
}

// GetAllMessagesByUUID mocks base method.
func (m *MockDBQueryHandler) GetAllMessagesByUUID(uuid string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllMessagesByUUID", uuid)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllMessagesByUUID indicates an expected call of GetAllMessagesByUUID.
func (mr *MockDBQueryHandlerMockRecorder) GetAllMessagesByUUID(uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllMessagesByUUID", reflect.TypeOf((*MockDBQueryHandler)(nil).GetAllMessagesByUUID), uuid)
}

// GetExternalUser mocks base method.
func (m *MockDBQueryHandler) GetExternalUser(name, ip string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExternalUser", name, ip)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExternalUser indicates an expected call of GetExternalUser.
func (mr *MockDBQueryHandlerMockRecorder) GetExternalUser(name, ip interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExternalUser", reflect.TypeOf((*MockDBQueryHandler)(nil).GetExternalUser), name, ip)
}

// GetExternalUserByID mocks base method.
func (m *MockDBQueryHandler) GetExternalUserByID(id int64) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExternalUserByID", id)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExternalUserByID indicates an expected call of GetExternalUserByID.
func (mr *MockDBQueryHandlerMockRecorder) GetExternalUserByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExternalUserByID", reflect.TypeOf((*MockDBQueryHandler)(nil).GetExternalUserByID), id)
}

// GetInternalUserByID mocks base method.
func (m *MockDBQueryHandler) GetInternalUserByID(id int64) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInternalUserByID", id)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInternalUserByID indicates an expected call of GetInternalUserByID.
func (mr *MockDBQueryHandlerMockRecorder) GetInternalUserByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInternalUserByID", reflect.TypeOf((*MockDBQueryHandler)(nil).GetInternalUserByID), id)
}

// UpdateExternalUserByID mocks base method.
func (m *MockDBQueryHandler) UpdateExternalUserByID(id int64, name, ip, email string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateExternalUserByID", id, name, ip, email)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateExternalUserByID indicates an expected call of UpdateExternalUserByID.
func (mr *MockDBQueryHandlerMockRecorder) UpdateExternalUserByID(id, name, ip, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateExternalUserByID", reflect.TypeOf((*MockDBQueryHandler)(nil).UpdateExternalUserByID), id, name, ip, email)
}

// UpdateInternalUserByID mocks base method.
func (m *MockDBQueryHandler) UpdateInternalUserByID(id, roleID int64, firstname, surname, email, password string) ([][]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateInternalUserByID", id, roleID, firstname, surname, email, password)
	ret0, _ := ret[0].([][]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateInternalUserByID indicates an expected call of UpdateInternalUserByID.
func (mr *MockDBQueryHandlerMockRecorder) UpdateInternalUserByID(id, roleID, firstname, surname, email, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateInternalUserByID", reflect.TypeOf((*MockDBQueryHandler)(nil).UpdateInternalUserByID), id, roleID, firstname, surname, email, password)
}

// MockdbConnectionHandler is a mock of dbConnectionHandler interface.
type MockdbConnectionHandler struct {
	ctrl     *gomock.Controller
	recorder *MockdbConnectionHandlerMockRecorder
}

// MockdbConnectionHandlerMockRecorder is the mock recorder for MockdbConnectionHandler.
type MockdbConnectionHandlerMockRecorder struct {
	mock *MockdbConnectionHandler
}

// NewMockdbConnectionHandler creates a new mock instance.
func NewMockdbConnectionHandler(ctrl *gomock.Controller) *MockdbConnectionHandler {
	mock := &MockdbConnectionHandler{ctrl: ctrl}
	mock.recorder = &MockdbConnectionHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockdbConnectionHandler) EXPECT() *MockdbConnectionHandlerMockRecorder {
	return m.recorder
}

// manageConnections mocks base method.
func (m *MockdbConnectionHandler) manageConnections() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "manageConnections")
}

// manageConnections indicates an expected call of manageConnections.
func (mr *MockdbConnectionHandlerMockRecorder) manageConnections() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "manageConnections", reflect.TypeOf((*MockdbConnectionHandler)(nil).manageConnections))
}
