// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"github.com/DelineaXPM/dsv-cli/errors"
	"github.com/DelineaXPM/dsv-cli/requests"
)

type FakeClient struct {
	DoRequestStub        func(string, string, interface{}) ([]byte, *errors.ApiError)
	doRequestMutex       sync.RWMutex
	doRequestArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 interface{}
	}
	doRequestReturns struct {
		result1 []byte
		result2 *errors.ApiError
	}
	doRequestReturnsOnCall map[int]struct {
		result1 []byte
		result2 *errors.ApiError
	}
	DoRequestOutStub        func(string, string, interface{}, interface{}) *errors.ApiError
	doRequestOutMutex       sync.RWMutex
	doRequestOutArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 interface{}
		arg4 interface{}
	}
	doRequestOutReturns struct {
		result1 *errors.ApiError
	}
	doRequestOutReturnsOnCall map[int]struct {
		result1 *errors.ApiError
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) DoRequest(arg1 string, arg2 string, arg3 interface{}) ([]byte, *errors.ApiError) {
	fake.doRequestMutex.Lock()
	ret, specificReturn := fake.doRequestReturnsOnCall[len(fake.doRequestArgsForCall)]
	fake.doRequestArgsForCall = append(fake.doRequestArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 interface{}
	}{arg1, arg2, arg3})
	stub := fake.DoRequestStub
	fakeReturns := fake.doRequestReturns
	fake.recordInvocation("DoRequest", []interface{}{arg1, arg2, arg3})
	fake.doRequestMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeClient) DoRequestCallCount() int {
	fake.doRequestMutex.RLock()
	defer fake.doRequestMutex.RUnlock()
	return len(fake.doRequestArgsForCall)
}

func (fake *FakeClient) DoRequestCalls(stub func(string, string, interface{}) ([]byte, *errors.ApiError)) {
	fake.doRequestMutex.Lock()
	defer fake.doRequestMutex.Unlock()
	fake.DoRequestStub = stub
}

func (fake *FakeClient) DoRequestArgsForCall(i int) (string, string, interface{}) {
	fake.doRequestMutex.RLock()
	defer fake.doRequestMutex.RUnlock()
	argsForCall := fake.doRequestArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeClient) DoRequestReturns(result1 []byte, result2 *errors.ApiError) {
	fake.doRequestMutex.Lock()
	defer fake.doRequestMutex.Unlock()
	fake.DoRequestStub = nil
	fake.doRequestReturns = struct {
		result1 []byte
		result2 *errors.ApiError
	}{result1, result2}
}

func (fake *FakeClient) DoRequestReturnsOnCall(i int, result1 []byte, result2 *errors.ApiError) {
	fake.doRequestMutex.Lock()
	defer fake.doRequestMutex.Unlock()
	fake.DoRequestStub = nil
	if fake.doRequestReturnsOnCall == nil {
		fake.doRequestReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 *errors.ApiError
		})
	}
	fake.doRequestReturnsOnCall[i] = struct {
		result1 []byte
		result2 *errors.ApiError
	}{result1, result2}
}

func (fake *FakeClient) DoRequestOut(arg1 string, arg2 string, arg3 interface{}, arg4 interface{}) *errors.ApiError {
	fake.doRequestOutMutex.Lock()
	ret, specificReturn := fake.doRequestOutReturnsOnCall[len(fake.doRequestOutArgsForCall)]
	fake.doRequestOutArgsForCall = append(fake.doRequestOutArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 interface{}
		arg4 interface{}
	}{arg1, arg2, arg3, arg4})
	stub := fake.DoRequestOutStub
	fakeReturns := fake.doRequestOutReturns
	fake.recordInvocation("DoRequestOut", []interface{}{arg1, arg2, arg3, arg4})
	fake.doRequestOutMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeClient) DoRequestOutCallCount() int {
	fake.doRequestOutMutex.RLock()
	defer fake.doRequestOutMutex.RUnlock()
	return len(fake.doRequestOutArgsForCall)
}

func (fake *FakeClient) DoRequestOutCalls(stub func(string, string, interface{}, interface{}) *errors.ApiError) {
	fake.doRequestOutMutex.Lock()
	defer fake.doRequestOutMutex.Unlock()
	fake.DoRequestOutStub = stub
}

func (fake *FakeClient) DoRequestOutArgsForCall(i int) (string, string, interface{}, interface{}) {
	fake.doRequestOutMutex.RLock()
	defer fake.doRequestOutMutex.RUnlock()
	argsForCall := fake.doRequestOutArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeClient) DoRequestOutReturns(result1 *errors.ApiError) {
	fake.doRequestOutMutex.Lock()
	defer fake.doRequestOutMutex.Unlock()
	fake.DoRequestOutStub = nil
	fake.doRequestOutReturns = struct {
		result1 *errors.ApiError
	}{result1}
}

func (fake *FakeClient) DoRequestOutReturnsOnCall(i int, result1 *errors.ApiError) {
	fake.doRequestOutMutex.Lock()
	defer fake.doRequestOutMutex.Unlock()
	fake.DoRequestOutStub = nil
	if fake.doRequestOutReturnsOnCall == nil {
		fake.doRequestOutReturnsOnCall = make(map[int]struct {
			result1 *errors.ApiError
		})
	}
	fake.doRequestOutReturnsOnCall[i] = struct {
		result1 *errors.ApiError
	}{result1}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doRequestMutex.RLock()
	defer fake.doRequestMutex.RUnlock()
	fake.doRequestOutMutex.RLock()
	defer fake.doRequestOutMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
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

var _ requests.Client = new(FakeClient)
