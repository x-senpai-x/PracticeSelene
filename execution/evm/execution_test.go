package evm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing
func mockLastFrameReturn(ctx *Context[any, Database], frameResult *FrameResult) EvmError {
	return EvmError{}
}

func mockExecuteFrame(frame *Frame, memory *SharedMemory, tables *InstructionTables[Context[any, Database]], ctx *Context[any, Database]) (InterpreterAction, EvmError) {
	return InterpreterAction{}, EvmError{}
}

func mockFrameCall(ctx *Context[any, Database], inputs *CallInputs) (FrameOrResult, EvmError) {
	return FrameOrResult{}, EvmError{}
}

func mockFrameCallReturn(ctx *Context[any, Database], frame *CallFrame, result InterpreterResult) (CallOutcome, EvmError) {
	return CallOutcome{}, EvmError{}
}

func mockInsertCallOutcome(ctx *Context[any, Database], frame *Frame, memory *SharedMemory, outcome CallOutcome) EvmError {
	return EvmError{}
}

func mockFrameCreate(ctx *Context[any, Database], inputs *CreateInputs) (FrameOrResult, EvmError) {
	return FrameOrResult{}, EvmError{}
}

func mockFrameCreateReturn(ctx *Context[any, Database], frame *CreateFrame, result InterpreterResult) (CreateOutcome, EvmError) {
	return CreateOutcome{}, EvmError{}
}

func mockInsertCreateOutcome(ctx *Context[any, Database], frame *Frame, outcome CreateOutcome) EvmError {
	return EvmError{}
}

func mockFrameEOFCreate(ctx *Context[any, Database], inputs EOFCreateInputs) (FrameOrResult, EvmError) {
	return FrameOrResult{}, EvmError{}
}

func mockFrameEOFCreateReturn(ctx *Context[any, Database], frame *EOFCreateFrame, result InterpreterResult) (CreateOutcome, EvmError) {
	return CreateOutcome{}, EvmError{}
}

func mockInsertEOFCreateOutcome(ctx *Context[any, Database], frame *CallFrame, result InterpreterResult) (CallOutcome, EvmError) {
	return CallOutcome{}, EvmError{}
}

// Test the ExecutionHandler with mock functions
func TestExecutionHandler(t *testing.T) {
	handler := ExecutionHandler[any, Database]{
		LastFrameReturn:        mockLastFrameReturn,
		ExecuteFrame:           mockExecuteFrame,
		Call:                   mockFrameCall,
		CallReturn:             mockFrameCallReturn,
		InsertCallOutcome:      mockInsertCallOutcome,
		Create:                 mockFrameCreate,
		CreateReturn:           mockFrameCreateReturn,
		InsertCreateOutcome:    mockInsertCreateOutcome,
		EOFCreate:              mockFrameEOFCreate,
		EOFCreateReturn:        mockFrameEOFCreateReturn,
		InsertEOFCreateOutcome: mockInsertEOFCreateOutcome,
	}

	ctx := &Context[any, Database]{}
	frame := &Frame{}
	memory := &SharedMemory{}
	callFrame := &CallFrame{}
	createFrame := &CreateFrame{}
	eofCreateFrame := &EOFCreateFrame{}
	frameResult := &FrameResult{}
	callInputs := &CallInputs{}
	createInputs := &CreateInputs{}
	eofCreateInputs := EOFCreateInputs{}

	// Test LastFrameReturn
	t.Run("LastFrameReturn Success", func(t *testing.T) {
		err := handler.LastFrameReturn(ctx, frameResult)
		assert.Empty(t, err.Message, "LastFrameReturn should not return an error")
	})

	// Test ExecuteFrame
	t.Run("ExecuteFrame Success", func(t *testing.T) {
		_, err := handler.ExecuteFrame(frame, memory, &InstructionTables[Context[any, Database]]{}, ctx)
		assert.Empty(t, err.Message, "ExecuteFrame should not return an error")
	})

	// Test Call
	t.Run("Call Success", func(t *testing.T) {
		_, err := handler.Call(ctx, callInputs)
		assert.Empty(t, err.Message, "Call should not return an error")
	})

	// Test CallReturn
	t.Run("CallReturn Success", func(t *testing.T) {
		_, err := handler.CallReturn(ctx, callFrame, InterpreterResult{})
		assert.Empty(t, err.Message, "CallReturn should not return an error")
	})

	// Test InsertCallOutcome
	t.Run("InsertCallOutcome Success", func(t *testing.T) {
		err := handler.InsertCallOutcome(ctx, frame, memory, CallOutcome{})
		assert.Empty(t, err.Message, "InsertCallOutcome should not return an error")
	})

	// Test Create
	t.Run("Create Success", func(t *testing.T) {
		_, err := handler.Create(ctx, createInputs)
		assert.Empty(t, err.Message, "Create should not return an error")
	})

	// Test CreateReturn
	t.Run("CreateReturn Success", func(t *testing.T) {
		_, err := handler.CreateReturn(ctx, createFrame, InterpreterResult{})
		assert.Empty(t, err.Message, "CreateReturn should not return an error")
	})

	// Test InsertCreateOutcome
	t.Run("InsertCreateOutcome Success", func(t *testing.T) {
		err := handler.InsertCreateOutcome(ctx, frame, CreateOutcome{})
		assert.Empty(t, err.Message, "InsertCreateOutcome should not return an error")
	})

	// Test EOFCreate
	t.Run("EOFCreate Success", func(t *testing.T) {
		_, err := handler.EOFCreate(ctx, eofCreateInputs)
		assert.Empty(t, err.Message, "EOFCreate should not return an error")
	})

	// Test EOFCreateReturn
	t.Run("EOFCreateReturn Success", func(t *testing.T) {
		_, err := handler.EOFCreateReturn(ctx, eofCreateFrame, InterpreterResult{})
		assert.Empty(t, err.Message, "EOFCreateReturn should not return an error")
	})

	// Test InsertEOFCreateOutcome
	t.Run("InsertEOFCreateOutcome Success", func(t *testing.T) {
		_, err := handler.InsertEOFCreateOutcome(ctx, callFrame, InterpreterResult{})
		assert.Empty(t, err.Message, "InsertEOFCreateOutcome should not return an error")
	})
}

// Test with errors
func TestExecutionHandlerWithError(t *testing.T) {
	handler := ExecutionHandler[any, Database]{
		LastFrameReturn: func(ctx *Context[any, Database], frameResult *FrameResult) EvmError {
			return EvmError{Message: "mock error"}
		},
		ExecuteFrame: func(frame *Frame, memory *SharedMemory, tables *InstructionTables[Context[any, Database]], ctx *Context[any, Database]) (InterpreterAction, EvmError) {
			return InterpreterAction{}, EvmError{Message: "mock execute error"}
		},
		Call: func(ctx *Context[any, Database], inputs *CallInputs) (FrameOrResult, EvmError) {
			return FrameOrResult{}, EvmError{Message: "mock call error"}
		},
		CallReturn: func(ctx *Context[any, Database], frame *CallFrame, result InterpreterResult) (CallOutcome, EvmError) {
			return CallOutcome{}, EvmError{Message: "mock call return error"}
		},
		InsertCallOutcome: func(ctx *Context[any, Database], frame *Frame, memory *SharedMemory, outcome CallOutcome) EvmError {
			return EvmError{Message: "mock insert call outcome error"}
		},
		Create: func(ctx *Context[any, Database], inputs *CreateInputs) (FrameOrResult, EvmError) {
			return FrameOrResult{}, EvmError{Message: "mock create error"}
		},
		CreateReturn: func(ctx *Context[any, Database], frame *CreateFrame, result InterpreterResult) (CreateOutcome, EvmError) {
			return CreateOutcome{}, EvmError{Message: "mock create return error"}
		},
		InsertCreateOutcome: func(ctx *Context[any, Database], frame *Frame, outcome CreateOutcome) EvmError {
			return EvmError{Message: "mock insert create outcome error"}
		},
		EOFCreate: func(ctx *Context[any, Database], inputs EOFCreateInputs) (FrameOrResult, EvmError) {
			return FrameOrResult{}, EvmError{Message: "mock EOF create error"}
		},
		EOFCreateReturn: func(ctx *Context[any, Database], frame *EOFCreateFrame, result InterpreterResult) (CreateOutcome, EvmError) {
			return CreateOutcome{}, EvmError{Message: "mock EOF create return error"}
		},
		InsertEOFCreateOutcome: func(ctx *Context[any, Database], frame *CallFrame, result InterpreterResult) (CallOutcome, EvmError) {
			return CallOutcome{}, EvmError{Message: "mock insert EOF create outcome error"}
		},
	}

	ctx := &Context[any, Database]{}
	frame := &Frame{}
	memory := &SharedMemory{}
	callFrame := &CallFrame{}
	createFrame := &CreateFrame{}
	eofCreateFrame := &EOFCreateFrame{}
	frameResult := &FrameResult{}
	callInputs := &CallInputs{}
	createInputs := &CreateInputs{}
	eofCreateInputs := EOFCreateInputs{}

	// Test LastFrameReturn with an error
	t.Run("LastFrameReturn Error", func(t *testing.T) {
		err := handler.LastFrameReturn(ctx, frameResult)
		assert.Error(t, err, "Expected error from LastFrameReturn")
		assert.Equal(t, "mock error", err.Message, "Error message should match expected")
	})

	// Test ExecuteFrame with an error
	t.Run("ExecuteFrame Error", func(t *testing.T) {
		_, err := handler.ExecuteFrame(frame, memory, &InstructionTables[Context[any, Database]]{}, ctx)
		assert.Error(t, err, "Expected error from ExecuteFrame")
		assert.Equal(t, "mock execute error", err.Message, "Error message should match expected")
	})

	// Test Call with an error
	t.Run("Call Error", func(t *testing.T) {
		_, err := handler.Call(ctx, callInputs)
		assert.Error(t, err, "Expected error from Call")
		assert.Equal(t, "mock call error", err.Message, "Error message should match expected")
	})

	// Test CallReturn with an error
	t.Run("CallReturn Error", func(t *testing.T) {
		_, err := handler.CallReturn(ctx, callFrame, InterpreterResult{})
		assert.Error(t, err, "Expected error from CallReturn")
		assert.Equal(t, "mock call return error", err.Message, "Error message should match expected")
	})

	// Test InsertCallOutcome with an error
	t.Run("InsertCallOutcome Error", func(t *testing.T) {
		err := handler.InsertCallOutcome(ctx, frame, memory, CallOutcome{})
		assert.Error(t, err, "Expected error from InsertCallOutcome")
		assert.Equal(t, "mock insert call outcome error", err.Message, "Error message should match expected")
	})

	// Test Create with an error
	t.Run("Create Error", func(t *testing.T) {
		_, err := handler.Create(ctx, createInputs)
		assert.Error(t, err, "Expected error from Create")
		assert.Equal(t, "mock create error", err.Message, "Error message should match expected")
	})

	// Test CreateReturn with an error
	t.Run("CreateReturn Error", func(t *testing.T) {
		_, err := handler.CreateReturn(ctx, createFrame, InterpreterResult{})
		assert.Error(t, err, "Expected error from CreateReturn")
		assert.Equal(t, "mock create return error", err.Message, "Error message should match expected")
	})

	// Test InsertCreateOutcome with an error
	t.Run("InsertCreateOutcome Error", func(t *testing.T) {
		err := handler.InsertCreateOutcome(ctx, frame, CreateOutcome{})
		assert.Error(t, err, "Expected error from InsertCreateOutcome")
		assert.Equal(t, "mock insert create outcome error", err.Message, "Error message should match expected")
	})

	// Test EOFCreate with an error
	t.Run("EOFCreate Error", func(t *testing.T) {
		_, err := handler.EOFCreate(ctx, eofCreateInputs)
		assert.Error(t, err, "Expected error from EOFCreate")
		assert.Equal(t, "mock EOF create error", err.Message, "Error message should match expected")
	})

	// Test EOFCreateReturn with an error
	t.Run("EOFCreateReturn Error", func(t *testing.T) {
		_, err := handler.EOFCreateReturn(ctx, eofCreateFrame, InterpreterResult{})
		assert.Error(t, err, "Expected error from EOFCreateReturn")
		assert.Equal(t, "mock EOF create return error", err.Message, "Error message should match expected")
	})

	// Test InsertEOFCreateOutcome with an error
	t.Run("InsertEOFCreateOutcome Error", func(t *testing.T) {
		_, err := handler.InsertEOFCreateOutcome(ctx, callFrame, InterpreterResult{})
		assert.Error(t, err, "Expected error from InsertEOFCreateOutcome")
		assert.Equal(t, "mock insert EOF create outcome error", err.Message, "Error message should match expected")
	})
}
