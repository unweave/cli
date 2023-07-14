// Code generated by counterfeiter. DO NOT EDIT.
package clientfakes

import (
	"context"
	"sync"

	"github.com/unweave/cli/client"
	"github.com/unweave/unweave/api/types"
)

type FakeProvider struct {
	ListNodeTypesStub        func(context.Context, types.Provider, bool) ([]types.NodeType, error)
	listNodeTypesMutex       sync.RWMutex
	listNodeTypesArgsForCall []struct {
		arg1 context.Context
		arg2 types.Provider
		arg3 bool
	}
	listNodeTypesReturns struct {
		result1 []types.NodeType
		result2 error
	}
	listNodeTypesReturnsOnCall map[int]struct {
		result1 []types.NodeType
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeProvider) ListNodeTypes(arg1 context.Context, arg2 types.Provider, arg3 bool) ([]types.NodeType, error) {
	fake.listNodeTypesMutex.Lock()
	ret, specificReturn := fake.listNodeTypesReturnsOnCall[len(fake.listNodeTypesArgsForCall)]
	fake.listNodeTypesArgsForCall = append(fake.listNodeTypesArgsForCall, struct {
		arg1 context.Context
		arg2 types.Provider
		arg3 bool
	}{arg1, arg2, arg3})
	stub := fake.ListNodeTypesStub
	fakeReturns := fake.listNodeTypesReturns
	fake.recordInvocation("ListNodeTypes", []interface{}{arg1, arg2, arg3})
	fake.listNodeTypesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeProvider) ListNodeTypesCallCount() int {
	fake.listNodeTypesMutex.RLock()
	defer fake.listNodeTypesMutex.RUnlock()
	return len(fake.listNodeTypesArgsForCall)
}

func (fake *FakeProvider) ListNodeTypesCalls(stub func(context.Context, types.Provider, bool) ([]types.NodeType, error)) {
	fake.listNodeTypesMutex.Lock()
	defer fake.listNodeTypesMutex.Unlock()
	fake.ListNodeTypesStub = stub
}

func (fake *FakeProvider) ListNodeTypesArgsForCall(i int) (context.Context, types.Provider, bool) {
	fake.listNodeTypesMutex.RLock()
	defer fake.listNodeTypesMutex.RUnlock()
	argsForCall := fake.listNodeTypesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeProvider) ListNodeTypesReturns(result1 []types.NodeType, result2 error) {
	fake.listNodeTypesMutex.Lock()
	defer fake.listNodeTypesMutex.Unlock()
	fake.ListNodeTypesStub = nil
	fake.listNodeTypesReturns = struct {
		result1 []types.NodeType
		result2 error
	}{result1, result2}
}

func (fake *FakeProvider) ListNodeTypesReturnsOnCall(i int, result1 []types.NodeType, result2 error) {
	fake.listNodeTypesMutex.Lock()
	defer fake.listNodeTypesMutex.Unlock()
	fake.ListNodeTypesStub = nil
	if fake.listNodeTypesReturnsOnCall == nil {
		fake.listNodeTypesReturnsOnCall = make(map[int]struct {
			result1 []types.NodeType
			result2 error
		})
	}
	fake.listNodeTypesReturnsOnCall[i] = struct {
		result1 []types.NodeType
		result2 error
	}{result1, result2}
}

func (fake *FakeProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.listNodeTypesMutex.RLock()
	defer fake.listNodeTypesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeProvider) recordInvocation(key string, args []interface{}) {
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

var _ client.Provider = new(FakeProvider)