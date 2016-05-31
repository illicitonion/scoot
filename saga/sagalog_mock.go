// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/scootdev/scoot/sched/sagalog (interfaces: SagaLog)

package saga

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of SagaLog interface
type MockSagaLog struct {
	ctrl     *gomock.Controller
	recorder *_MockSagaLogRecorder
}

// Recorder for MockSagaLog (not exported)
type _MockSagaLogRecorder struct {
	mock *MockSagaLog
}

func NewMockSagaLog(ctrl *gomock.Controller) *MockSagaLog {
	mock := &MockSagaLog{ctrl: ctrl}
	mock.recorder = &_MockSagaLogRecorder{mock}
	return mock
}

func (_m *MockSagaLog) EXPECT() *_MockSagaLogRecorder {
	return _m.recorder
}

func (_m *MockSagaLog) GetSagaState(_param0 string) (*SagaState, error) {
	ret := _m.ctrl.Call(_m, "GetSagaState", _param0)
	ret0, _ := ret[0].(*SagaState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSagaLogRecorder) GetSagaState(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSagaState", arg0)
}

func (_m *MockSagaLog) LogMessage(_param0 sagaMessage) error {
	ret := _m.ctrl.Call(_m, "LogMessage", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSagaLogRecorder) LogMessage(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LogMessage", arg0)
}

func (_m *MockSagaLog) StartSaga(_param0 string, _param1 []byte) error {
	ret := _m.ctrl.Call(_m, "StartSaga", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSagaLogRecorder) StartSaga(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StartSaga", arg0, arg1)
}
