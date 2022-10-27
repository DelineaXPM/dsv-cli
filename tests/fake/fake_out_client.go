// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/format"
)

type FakeOutClient struct {
	FailStub        func(error)
	failMutex       sync.RWMutex
	failArgsForCall []struct {
		arg1 error
	}
	FailEStub        func(*errors.ApiError)
	failEMutex       sync.RWMutex
	failEArgsForCall []struct {
		arg1 *errors.ApiError
	}
	FailFStub        func(string, ...interface{})
	failFMutex       sync.RWMutex
	failFArgsForCall []struct {
		arg1 string
		arg2 []interface{}
	}
	FailSStub        func(string)
	failSMutex       sync.RWMutex
	failSArgsForCall []struct {
		arg1 string
	}
	WriteResponseStub        func([]byte, *errors.ApiError)
	writeResponseMutex       sync.RWMutex
	writeResponseArgsForCall []struct {
		arg1 []byte
		arg2 *errors.ApiError
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOutClient) Fail(arg1 error) {
	fake.failMutex.Lock()
	fake.failArgsForCall = append(fake.failArgsForCall, struct {
		arg1 error
	}{arg1})
	stub := fake.FailStub
	fake.recordInvocation("Fail", []interface{}{arg1})
	fake.failMutex.Unlock()
	if stub != nil {
		fake.FailStub(arg1)
	}
}

func (fake *FakeOutClient) FailCallCount() int {
	fake.failMutex.RLock()
	defer fake.failMutex.RUnlock()
	return len(fake.failArgsForCall)
}

func (fake *FakeOutClient) FailCalls(stub func(error)) {
	fake.failMutex.Lock()
	defer fake.failMutex.Unlock()
	fake.FailStub = stub
}

func (fake *FakeOutClient) FailArgsForCall(i int) error {
	fake.failMutex.RLock()
	defer fake.failMutex.RUnlock()
	argsForCall := fake.failArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOutClient) FailE(arg1 *errors.ApiError) {
	fake.failEMutex.Lock()
	fake.failEArgsForCall = append(fake.failEArgsForCall, struct {
		arg1 *errors.ApiError
	}{arg1})
	stub := fake.FailEStub
	fake.recordInvocation("FailE", []interface{}{arg1})
	fake.failEMutex.Unlock()
	if stub != nil {
		fake.FailEStub(arg1)
	}
}

func (fake *FakeOutClient) FailECallCount() int {
	fake.failEMutex.RLock()
	defer fake.failEMutex.RUnlock()
	return len(fake.failEArgsForCall)
}

func (fake *FakeOutClient) FailECalls(stub func(*errors.ApiError)) {
	fake.failEMutex.Lock()
	defer fake.failEMutex.Unlock()
	fake.FailEStub = stub
}

func (fake *FakeOutClient) FailEArgsForCall(i int) *errors.ApiError {
	fake.failEMutex.RLock()
	defer fake.failEMutex.RUnlock()
	argsForCall := fake.failEArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOutClient) FailF(arg1 string, arg2 ...interface{}) {
	fake.failFMutex.Lock()
	fake.failFArgsForCall = append(fake.failFArgsForCall, struct {
		arg1 string
		arg2 []interface{}
	}{arg1, arg2})
	stub := fake.FailFStub
	fake.recordInvocation("FailF", []interface{}{arg1, arg2})
	fake.failFMutex.Unlock()
	if stub != nil {
		fake.FailFStub(arg1, arg2...)
	}
}

func (fake *FakeOutClient) FailFCallCount() int {
	fake.failFMutex.RLock()
	defer fake.failFMutex.RUnlock()
	return len(fake.failFArgsForCall)
}

func (fake *FakeOutClient) FailFCalls(stub func(string, ...interface{})) {
	fake.failFMutex.Lock()
	defer fake.failFMutex.Unlock()
	fake.FailFStub = stub
}

func (fake *FakeOutClient) FailFArgsForCall(i int) (string, []interface{}) {
	fake.failFMutex.RLock()
	defer fake.failFMutex.RUnlock()
	argsForCall := fake.failFArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeOutClient) FailS(arg1 string) {
	fake.failSMutex.Lock()
	fake.failSArgsForCall = append(fake.failSArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.FailSStub
	fake.recordInvocation("FailS", []interface{}{arg1})
	fake.failSMutex.Unlock()
	if stub != nil {
		fake.FailSStub(arg1)
	}
}

func (fake *FakeOutClient) FailSCallCount() int {
	fake.failSMutex.RLock()
	defer fake.failSMutex.RUnlock()
	return len(fake.failSArgsForCall)
}

func (fake *FakeOutClient) FailSCalls(stub func(string)) {
	fake.failSMutex.Lock()
	defer fake.failSMutex.Unlock()
	fake.FailSStub = stub
}

func (fake *FakeOutClient) FailSArgsForCall(i int) string {
	fake.failSMutex.RLock()
	defer fake.failSMutex.RUnlock()
	argsForCall := fake.failSArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOutClient) WriteResponse(arg1 []byte, arg2 *errors.ApiError) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.writeResponseMutex.Lock()
	fake.writeResponseArgsForCall = append(fake.writeResponseArgsForCall, struct {
		arg1 []byte
		arg2 *errors.ApiError
	}{arg1Copy, arg2})
	stub := fake.WriteResponseStub
	fake.recordInvocation("WriteResponse", []interface{}{arg1Copy, arg2})
	fake.writeResponseMutex.Unlock()
	if stub != nil {
		fake.WriteResponseStub(arg1, arg2)
	}
}

func (fake *FakeOutClient) WriteResponseCallCount() int {
	fake.writeResponseMutex.RLock()
	defer fake.writeResponseMutex.RUnlock()
	return len(fake.writeResponseArgsForCall)
}

func (fake *FakeOutClient) WriteResponseCalls(stub func([]byte, *errors.ApiError)) {
	fake.writeResponseMutex.Lock()
	defer fake.writeResponseMutex.Unlock()
	fake.WriteResponseStub = stub
}

func (fake *FakeOutClient) WriteResponseArgsForCall(i int) ([]byte, *errors.ApiError) {
	fake.writeResponseMutex.RLock()
	defer fake.writeResponseMutex.RUnlock()
	argsForCall := fake.writeResponseArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeOutClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.failMutex.RLock()
	defer fake.failMutex.RUnlock()
	fake.failEMutex.RLock()
	defer fake.failEMutex.RUnlock()
	fake.failFMutex.RLock()
	defer fake.failFMutex.RUnlock()
	fake.failSMutex.RLock()
	defer fake.failSMutex.RUnlock()
	fake.writeResponseMutex.RLock()
	defer fake.writeResponseMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeOutClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ format.OutClient = new(FakeOutClient)
