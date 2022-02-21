package interpreter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	wasm "github.com/tetratelabs/wazero/internal/wasm"
	"github.com/tetratelabs/wazero/internal/wasm/buildoptions"
	publicwasm "github.com/tetratelabs/wazero/wasm"
)

func TestInterpreter_PushFrame(t *testing.T) {
	f1 := &interpreterFrame{}
	f2 := &interpreterFrame{}

	it := interpreter{}
	require.Empty(t, it.frames)

	it.pushFrame(f1)
	require.Equal(t, []*interpreterFrame{f1}, it.frames)

	it.pushFrame(f2)
	require.Equal(t, []*interpreterFrame{f1, f2}, it.frames)
}

func TestInterpreter_PushFrame_StackOverflow(t *testing.T) {
	defer func() { callStackCeiling = buildoptions.CallStackCeiling }()

	callStackCeiling = 3

	f1 := &interpreterFrame{}
	f2 := &interpreterFrame{}
	f3 := &interpreterFrame{}
	f4 := &interpreterFrame{}

	it := interpreter{}
	it.pushFrame(f1)
	it.pushFrame(f2)
	it.pushFrame(f3)
	require.Panics(t, func() { it.pushFrame(f4) })
}

func TestInterpreter_CallHostFunc(t *testing.T) {
	t.Run("defaults to module memory when call stack empty", func(t *testing.T) {
		memory := &wasm.MemoryInstance{}
		var ctxMemory publicwasm.Memory
		hostFn := reflect.ValueOf(func(ctx publicwasm.ModuleContext) {
			ctxMemory = ctx.Memory()
		})
		module := &wasm.ModuleInstance{Memory: memory}
		it := interpreter{functions: map[wasm.FunctionAddress]*interpreterFunction{
			0: {hostFn: &hostFn, funcInstance: &wasm.FunctionInstance{
				FunctionKind: wasm.FunctionKindGoModuleContext,
				FunctionType: &wasm.TypeInstance{
					Type: &wasm.FunctionType{
						Params:  []wasm.ValueType{},
						Results: []wasm.ValueType{},
					},
				},
				ModuleInstance: module,
			},
			},
		}}

		// When calling a host func directly, there may be no stack. This ensures the module's memory is used.
		it.callHostFunc(newModuleContext(&it, module), it.functions[0])
		require.Same(t, memory, ctxMemory)
	})
}

func newModuleContext(engine wasm.Engine, module *wasm.ModuleInstance) *wasm.ModuleContext {
	ctx := wasm.NewModuleContext(&wasm.Store{
		Engine:          engine,
		ModuleInstances: map[string]*wasm.ModuleInstance{"test": module},
	}, module)
	return ctx
}