// Code generated by counterfeiter. DO NOT EDIT.
package clientfakes

import (
	"context"
	"sync"

	"github.com/unweave/cli/client"
	"github.com/unweave/unweave/api/types"
)

type FakeExecer struct {
	CreateStub        func(context.Context, string, string, types.ExecCreateParams) (*types.Exec, error)
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 types.ExecCreateParams
	}
	createReturns struct {
		result1 *types.Exec
		result2 error
	}
	createReturnsOnCall map[int]struct {
		result1 *types.Exec
		result2 error
	}
	ExecStub        func(context.Context, []string, string, *string) (*types.Exec, error)
	execMutex       sync.RWMutex
	execArgsForCall []struct {
		arg1 context.Context
		arg2 []string
		arg3 string
		arg4 *string
	}
	execReturns struct {
		result1 *types.Exec
		result2 error
	}
	execReturnsOnCall map[int]struct {
		result1 *types.Exec
		result2 error
	}
	GetStub        func(context.Context, string, string, string) (*types.Exec, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 string
	}
	getReturns struct {
		result1 *types.Exec
		result2 error
	}
	getReturnsOnCall map[int]struct {
		result1 *types.Exec
		result2 error
	}
	ListStub        func(context.Context, string, string, bool) ([]types.Exec, error)
	listMutex       sync.RWMutex
	listArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 bool
	}
	listReturns struct {
		result1 []types.Exec
		result2 error
	}
	listReturnsOnCall map[int]struct {
		result1 []types.Exec
		result2 error
	}
	TerminateStub        func(context.Context, string, string, string) error
	terminateMutex       sync.RWMutex
	terminateArgsForCall []struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 string
	}
	terminateReturns struct {
		result1 error
	}
	terminateReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeExecer) Create(arg1 context.Context, arg2 string, arg3 string, arg4 types.ExecCreateParams) (*types.Exec, error) {
	fake.createMutex.Lock()
	ret, specificReturn := fake.createReturnsOnCall[len(fake.createArgsForCall)]
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 types.ExecCreateParams
	}{arg1, arg2, arg3, arg4})
	stub := fake.CreateStub
	fakeReturns := fake.createReturns
	fake.recordInvocation("Create", []interface{}{arg1, arg2, arg3, arg4})
	fake.createMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeExecer) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeExecer) CreateCalls(stub func(context.Context, string, string, types.ExecCreateParams) (*types.Exec, error)) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = stub
}

func (fake *FakeExecer) CreateArgsForCall(i int) (context.Context, string, string, types.ExecCreateParams) {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	argsForCall := fake.createArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeExecer) CreateReturns(result1 *types.Exec, result2 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) CreateReturnsOnCall(i int, result1 *types.Exec, result2 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	if fake.createReturnsOnCall == nil {
		fake.createReturnsOnCall = make(map[int]struct {
			result1 *types.Exec
			result2 error
		})
	}
	fake.createReturnsOnCall[i] = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) Exec(arg1 context.Context, arg2 []string, arg3 string, arg4 *string) (*types.Exec, error) {
	var arg2Copy []string
	if arg2 != nil {
		arg2Copy = make([]string, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.execMutex.Lock()
	ret, specificReturn := fake.execReturnsOnCall[len(fake.execArgsForCall)]
	fake.execArgsForCall = append(fake.execArgsForCall, struct {
		arg1 context.Context
		arg2 []string
		arg3 string
		arg4 *string
	}{arg1, arg2Copy, arg3, arg4})
	stub := fake.ExecStub
	fakeReturns := fake.execReturns
	fake.recordInvocation("Exec", []interface{}{arg1, arg2Copy, arg3, arg4})
	fake.execMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeExecer) ExecCallCount() int {
	fake.execMutex.RLock()
	defer fake.execMutex.RUnlock()
	return len(fake.execArgsForCall)
}

func (fake *FakeExecer) ExecCalls(stub func(context.Context, []string, string, *string) (*types.Exec, error)) {
	fake.execMutex.Lock()
	defer fake.execMutex.Unlock()
	fake.ExecStub = stub
}

func (fake *FakeExecer) ExecArgsForCall(i int) (context.Context, []string, string, *string) {
	fake.execMutex.RLock()
	defer fake.execMutex.RUnlock()
	argsForCall := fake.execArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeExecer) ExecReturns(result1 *types.Exec, result2 error) {
	fake.execMutex.Lock()
	defer fake.execMutex.Unlock()
	fake.ExecStub = nil
	fake.execReturns = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) ExecReturnsOnCall(i int, result1 *types.Exec, result2 error) {
	fake.execMutex.Lock()
	defer fake.execMutex.Unlock()
	fake.ExecStub = nil
	if fake.execReturnsOnCall == nil {
		fake.execReturnsOnCall = make(map[int]struct {
			result1 *types.Exec
			result2 error
		})
	}
	fake.execReturnsOnCall[i] = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) Get(arg1 context.Context, arg2 string, arg3 string, arg4 string) (*types.Exec, error) {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 string
	}{arg1, arg2, arg3, arg4})
	stub := fake.GetStub
	fakeReturns := fake.getReturns
	fake.recordInvocation("Get", []interface{}{arg1, arg2, arg3, arg4})
	fake.getMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeExecer) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeExecer) GetCalls(stub func(context.Context, string, string, string) (*types.Exec, error)) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *FakeExecer) GetArgsForCall(i int) (context.Context, string, string, string) {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeExecer) GetReturns(result1 *types.Exec, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) GetReturnsOnCall(i int, result1 *types.Exec, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 *types.Exec
			result2 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 *types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) List(arg1 context.Context, arg2 string, arg3 string, arg4 bool) ([]types.Exec, error) {
	fake.listMutex.Lock()
	ret, specificReturn := fake.listReturnsOnCall[len(fake.listArgsForCall)]
	fake.listArgsForCall = append(fake.listArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 bool
	}{arg1, arg2, arg3, arg4})
	stub := fake.ListStub
	fakeReturns := fake.listReturns
	fake.recordInvocation("List", []interface{}{arg1, arg2, arg3, arg4})
	fake.listMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeExecer) ListCallCount() int {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	return len(fake.listArgsForCall)
}

func (fake *FakeExecer) ListCalls(stub func(context.Context, string, string, bool) ([]types.Exec, error)) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = stub
}

func (fake *FakeExecer) ListArgsForCall(i int) (context.Context, string, string, bool) {
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	argsForCall := fake.listArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeExecer) ListReturns(result1 []types.Exec, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	fake.listReturns = struct {
		result1 []types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) ListReturnsOnCall(i int, result1 []types.Exec, result2 error) {
	fake.listMutex.Lock()
	defer fake.listMutex.Unlock()
	fake.ListStub = nil
	if fake.listReturnsOnCall == nil {
		fake.listReturnsOnCall = make(map[int]struct {
			result1 []types.Exec
			result2 error
		})
	}
	fake.listReturnsOnCall[i] = struct {
		result1 []types.Exec
		result2 error
	}{result1, result2}
}

func (fake *FakeExecer) Terminate(arg1 context.Context, arg2 string, arg3 string, arg4 string) error {
	fake.terminateMutex.Lock()
	ret, specificReturn := fake.terminateReturnsOnCall[len(fake.terminateArgsForCall)]
	fake.terminateArgsForCall = append(fake.terminateArgsForCall, struct {
		arg1 context.Context
		arg2 string
		arg3 string
		arg4 string
	}{arg1, arg2, arg3, arg4})
	stub := fake.TerminateStub
	fakeReturns := fake.terminateReturns
	fake.recordInvocation("Terminate", []interface{}{arg1, arg2, arg3, arg4})
	fake.terminateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeExecer) TerminateCallCount() int {
	fake.terminateMutex.RLock()
	defer fake.terminateMutex.RUnlock()
	return len(fake.terminateArgsForCall)
}

func (fake *FakeExecer) TerminateCalls(stub func(context.Context, string, string, string) error) {
	fake.terminateMutex.Lock()
	defer fake.terminateMutex.Unlock()
	fake.TerminateStub = stub
}

func (fake *FakeExecer) TerminateArgsForCall(i int) (context.Context, string, string, string) {
	fake.terminateMutex.RLock()
	defer fake.terminateMutex.RUnlock()
	argsForCall := fake.terminateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeExecer) TerminateReturns(result1 error) {
	fake.terminateMutex.Lock()
	defer fake.terminateMutex.Unlock()
	fake.TerminateStub = nil
	fake.terminateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeExecer) TerminateReturnsOnCall(i int, result1 error) {
	fake.terminateMutex.Lock()
	defer fake.terminateMutex.Unlock()
	fake.TerminateStub = nil
	if fake.terminateReturnsOnCall == nil {
		fake.terminateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.terminateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeExecer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.execMutex.RLock()
	defer fake.execMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.listMutex.RLock()
	defer fake.listMutex.RUnlock()
	fake.terminateMutex.RLock()
	defer fake.terminateMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeExecer) recordInvocation(key string, args []interface{}) {
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

var _ client.Execer = new(FakeExecer)