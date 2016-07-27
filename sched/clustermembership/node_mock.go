// Automatically generated by MockGen. DO NOT EDIT!
// Source: node.go

package cluster_membership

import (
	gomock "github.com/golang/mock/gomock"
	sched "github.com/scootdev/scoot/sched"
)

// Mock of Node interface
type MockNode struct {
	ctrl     *gomock.Controller
	recorder *_MockNodeRecorder
}

// Recorder for MockNode (not exported)
type _MockNodeRecorder struct {
	mock *MockNode
}

func NewMockNode(ctrl *gomock.Controller) *MockNode {
	mock := &MockNode{ctrl: ctrl}
	mock.recorder = &_MockNodeRecorder{mock}
	return mock
}

func (_m *MockNode) EXPECT() *_MockNodeRecorder {
	return _m.recorder
}

func (_m *MockNode) Id() string {
	ret := _m.ctrl.Call(_m, "Id")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockNodeRecorder) Id() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Id")
}

func (_m *MockNode) SendMessage(task sched.TaskDefinition) error {
	ret := _m.ctrl.Call(_m, "SendMessage", task)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockNodeRecorder) SendMessage(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendMessage", arg0)
}