package evm

import (
	"testing"
    "fmt"
	"github.com/stretchr/testify/assert"
	"errors"
)

// Define MockEXT and MockDB as simple mocks of the EXT and DB types
type MockEXT struct{}
type MockDB2 Database
type mockDB struct{}

// Mock extension type for testing
type mockEXT struct{}
func TestPreverifyTransactionInner(t *testing.T) {
    tests := []struct {
        name            string
        envValidation   error
        gasValidation   struct {
            gas uint64
            err error
        }
        stateValidation error
        wantGas         uint64
        wantErr         bool
        expectedErrMsg  string
    }{
        {
            name:            "Success case",
            envValidation:   nil,
            gasValidation:   struct{gas uint64; err error}{gas: 21000, err: nil},
            stateValidation: nil,
            wantGas:         21000,
            wantErr:         false,
        },
        {
            name:            "Env validation fails",
            envValidation:   errors.New("invalid environment"),
            gasValidation:   struct{gas uint64; err error}{gas: 0, err: nil},
            stateValidation: nil,
            wantGas:         0,
            wantErr:         true,
            expectedErrMsg:  "invalid environment",
        },
        {
            name:            "Gas validation fails",
            envValidation:   nil,
            gasValidation:   struct{gas uint64; err error}{gas: 0, err: errors.New("insufficient gas")},
            stateValidation: nil,
            wantGas:         0,
            wantErr:         true,
            expectedErrMsg:  "insufficient gas",
        },
        {
            name:            "State validation fails",
            envValidation:   nil,
            gasValidation:   struct{gas uint64; err error}{gas: 21000, err: nil},
            stateValidation: errors.New("invalid state"),
            wantGas:         0,
            wantErr:         true,
            expectedErrMsg:  "invalid state",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create mock validation handlers
            mockValidation := ValidationHandler[mockEXT, MockDB2]{
                Env: func(env *Env) error {
                    return tt.envValidation
                },
                InitialTxGas: func(env *Env) (uint64, error) {
                    return tt.gasValidation.gas, tt.gasValidation.err
                },
                TxAgainstState: func(ctx *Context[mockEXT, MockDB2]) error {
                    return tt.stateValidation
                },
            }

            evmCtx := EvmContext[MockDB2]{
				Inner: InnerEvmContext[MockDB2]{
					Env: &Env{},
				},
            }

            evm := &Evm[mockEXT, MockDB2]{
                Context: Context[mockEXT, MockDB2]{
                    Evm: evmCtx,
                },
                Handler: Handler[Context[mockEXT, MockDB2], mockEXT, MockDB2]{
                    Validation: mockValidation,
                },
            }

            gotGas, gotErr := evm.preverifyTransactionInner()

            // Check if we got the expected gas value
            if gotGas != tt.wantGas {
                t.Errorf("preverifyTransactionInner() gotGas = %v, want %v", gotGas, tt.wantGas)
            }

            // Check if we got the expected error
            if tt.wantErr {
                if gotErr.Message != tt.expectedErrMsg {
                    t.Errorf("preverifyTransactionInner() error = %v, expected error = %v", gotErr.Message, tt.expectedErrMsg)
                }
            } else if gotErr.Message != "" {
                t.Errorf("preverifyTransactionInner() unexpected error = %v", gotErr.Message)
            }
        })
    }
}
func TestPreverifyTransactionInner_NilEnv(t *testing.T) {
    evm := &Evm[mockEXT, MockDB2]{
        Context: Context[mockEXT,MockDB2]{

		},
        Handler: Handler[Context[mockEXT, MockDB2], mockEXT, MockDB2]{
            Validation: ValidationHandler[mockEXT, MockDB2]{
                Env: func(env *Env) error {
                    if env == nil {
                        return errors.New("nil environment")
                    }
                    return nil
                },
                InitialTxGas: func(env *Env) (uint64, error) {
                    return 0, nil
                },
                TxAgainstState: func(ctx *Context[mockEXT, MockDB2]) error {
                    return nil
                },
            },
        },
    }

    gotGas, gotErr := evm.preverifyTransactionInner()
    if gotGas != 0 {
        t.Errorf("Expected 0 gas for nil environment, got %v", gotGas)
    }
    if gotErr.Message != "nil environment" {
        t.Errorf("Expected 'nil environment' error, got %v", gotErr.Message)
    }
}
// TestNewEvm verifies that a new Evm instance is created with the expected specID and context
func TestNewEvm(t *testing.T) {
	mockHandler := &Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{}
	mockContext := Context[MockEXT, MockDB2]{}

	mockHandler.Cfg = HandlerCfg{specID: 1}
	evm := NewEvm[MockEXT, MockDB2](mockContext, *mockHandler)

	assert.Equal(t, mockHandler.Cfg.specID, evm.specId(), "SpecId should match the handler config specID")
	assert.Equal(t, mockContext, evm.Context, "Context should be correctly set in Evm")
}

// TestSpecId verifies that the SpecId method returns the expected specID from the handler configuration
func TestSpecId(t *testing.T) {
	handler := Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{}
	handler.Cfg.specID = 42
	evm := Evm[MockEXT, MockDB2]{Handler: handler}

	assert.Equal(t, SpecId(42), evm.specId(), "SpecId should return the handler's config specID")
}

func Test__preverifyTransactionInner(t *testing.T) {
	config := HandlerCfg{
		specID: 1, // Adjust SpecID as needed; this is just a placeholder
	}

	// Initialize the handler with a function literal matching ValidateTxEnvAgainstState
	handler := Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{
		Cfg: config,
		Validation: ValidationHandler[MockEXT, MockDB2]{
			TxAgainstState: func(ctx *Context[MockEXT, MockDB2]) error {
				// Simulate failure
				return fmt.Errorf("Preverification failed")
			},
			InitialTxGas: func(env *Env) (uint64, error){
				return 0, fmt.Errorf("Preverification failed")
			},
			Env: func(env *Env) error{
				return fmt.Errorf("Preverification failed")
			},
		},
	}
	// Set up the environment variable for EvmContext
	var env *Env = &Env{} // Ensure env is initialized properly

	inner := InnerEvmContext[MockDB2]{
		Env: env,
	}

	context := Context[MockEXT, MockDB2]{
		Evm: EvmContext[MockDB2]{
			Inner: inner,
		},
		External: MockEXT{},
	}

	// Initialize Evm instance with the handler and context
	evm := Evm[MockEXT, MockDB2]{
		Context: context,
		Handler: handler,
	}

	_ , result:= evm.preverifyTransactionInner()
	assert.NotNil(t, result.Message, "Expecting an error due to preverification failure")
	assert.Equal(t, "Preverification failed", result.Message, "Error message should match expected preverification failure")
}



// Setup mock objects and return an instance of Evm with mocked Handler
func setupMockEvm() *Evm[MockEXT, MockDB2] {
	handlerCfg := HandlerCfg{
		specID: 1, // Adjust SpecID as needed; this is just a placeholder
	}
	mockContext := Context[MockEXT, MockDB2]{}

	mockHandler := &Handler[Context[MockEXT,MockDB2], MockEXT, MockDB2]{
		Cfg: handlerCfg,
		PreExecution: PreExecutionHandler[MockEXT, MockDB2]{
			LoadAccounts: func(ctx *Context[MockEXT, MockDB2]) error {
				return nil // No error for successful account load
			},
			DeductCaller: func(ctx *Context[MockEXT, MockDB2]) EVMResultGeneric[ResultAndState, DatabaseError] {
				return EVMResultGeneric[ResultAndState, DatabaseError]{} // No error for successful deduction
			},
			LoadPrecompiles: func() ContextPrecompiles[MockDB2] {
				return ContextPrecompiles[MockDB2]{} // Provide empty precompiles for testing
			},
		},
		Execution: ExecutionHandler[MockEXT, MockDB2]{
			Call: func(ctx *Context[MockEXT , MockDB2], inputs *CallInputs) (FrameOrResult, EvmError) {
				return FrameOrResult{}, EvmError{} // No error for successful call
			},
			Create: func(ctx *Context[MockEXT , MockDB2], inputs *CreateInputs) (FrameOrResult, EvmError) {
				return FrameOrResult{}, EvmError{} // No error for successful create
			},
			EOFCreate: func(ctx *Context[MockEXT, MockDB2], inputs EOFCreateInputs) (FrameOrResult, EvmError) {
				return FrameOrResult{}, EvmError{} // No error for successful EOF create
			},
		},
		PostExecution: PostExecutionHandler[MockEXT, MockDB2]{
			ReimburseCaller: func(ctx *Context[MockEXT , MockDB2], gas *Gas) EVMResultGeneric[struct{}, any] {
				return EVMResultGeneric[struct{}, any]{} // Successful reimbursement
			},
			RewardBeneficiary: func(ctx *Context[MockEXT,MockDB2], gas *Gas) EVMResultGeneric[struct{}, any] {
				return EVMResultGeneric[struct{}, any]{} // Successful reward
			},
			Output: func(ctx *Context[MockEXT,MockDB2], frameResult FrameResult) (ResultAndState, EvmError) {
				return ResultAndState{}, EvmError{} // No error for successful output
			},
		},
	}
	return &Evm[MockEXT, MockDB2]{Handler: *mockHandler, Context: *&mockContext}
}
func TestIntoContextWithHandlerCfg(t *testing.T) {
	// Setup mock handler and context
	mockHandler := Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{
		Cfg: HandlerCfg{specID: 123},
	}
	mockContext := Context[MockEXT, MockDB2]{}

	evm := Evm[MockEXT, MockDB2]{
		Context: mockContext,
		Handler: mockHandler,
	}

	// Call the method
	contextWithCfg := evm.IntoContextWithHandlerCfg()

	// Verify the returned struct
	assert.Equal(t, mockContext, contextWithCfg.Context, "Context should match the original")
	assert.Equal(t, mockHandler.Cfg, contextWithCfg.Cfg, "Handler config should match the original")
}



// Mock types and structures
type MockPrecompiles struct {
    Addresses map[Address]struct{}
}

type MockStaticRef struct {
    Addresses map[Address]struct{}
}

type MockPrecompilesCow struct {
    StaticRef *MockStaticRef
}

type MockContextPrecompiles struct {
    Inner MockPrecompilesCow
}

func TestTransactPreverifiedInner(t *testing.T) {
    tests := []struct {
        name            string
        initialGasSpend uint64
        setupMocks      func() (*Evm[mockEXT, MockDB2], error)
        wantErr         bool
        expectedErrMsg  string
    }{
        {
            name:            "Successful Call transaction",
            initialGasSpend: 21000,
            setupMocks: func() (*Evm[mockEXT, MockDB2], error) {
                mockPreExec := PreExecutionHandler[mockEXT, MockDB2]{
                    LoadAccounts: func(ctx *Context[mockEXT, MockDB2]) error {
                        return nil
                    },
                    LoadPrecompiles: func() ContextPrecompiles[MockDB2] {
                        return ContextPrecompiles[MockDB2]{
                            Inner: PrecompilesCow[MockDB2]{
                                StaticRef: &Precompiles{
                                    Addresses: make(map[Address]struct{}),
                                },
                            },
                        }
                    },
                    DeductCaller: func(ctx *Context[mockEXT, MockDB2]) EVMResultGeneric[ResultAndState, DatabaseError] {
                        return EVMResultGeneric[ResultAndState, DatabaseError]{
                            Value: ResultAndState{},
                            Err:   nil,
                        }
                    },
                }

                mockExec := ExecutionHandler[mockEXT, MockDB2]{
                    Call: func(ctx *Context[mockEXT, MockDB2], inputs *CallInputs) (FrameOrResult, EvmError) {
                        return FrameOrResult{
                            Type: Result_FrameOrResult,
                            Result: FrameResult{
                                ResultType: Call,
                                Call:      &CallOutcome{},
                            },
                        }, EvmError{}
                    },
                    LastFrameReturn: func(ctx *Context[mockEXT, MockDB2], result *FrameResult) EvmError {
                        return EvmError{}
                    },
                }

                mockPostExec := PostExecutionHandler[mockEXT, MockDB2]{
                    ReimburseCaller: func(ctx *Context[mockEXT, MockDB2], gas uint64) EVMResultGeneric {
                    },
                    RewardBeneficiary: func(ctx *Context[mockEXT, MockDB2], gas uint64) {
                    },
                    Output: func(ctx *Context[mockEXT, MockDB2], result FrameResult) (ResultAndState, error) {
                        return ResultAndState{}, nil
                    },
                }

                // Create mock EVM instance
                evm := &Evm[mockEXT, MockDB2]{
                    Context: Context[mockEXT, MockDB2]{
                        Evm: EvmContext[MockDB2]{
                            Inner: struct {
                                Env            *Env
                                JournaledState struct {
                                    WarmPreloadedAddresses map[Address]struct{}
                                }
                            }{
                                Env: &Env{
                                    Tx: TxEnv{
                                        TransactTo: TransactTo{
                                            Type:    Call2,
                                            Address: &Address{},
                                        },
                                        GasLimit: 100000,
                                    },
                                },
                                JournaledState: struct {
                                    WarmPreloadedAddresses map[Address]struct{}
                                }{
                                    WarmPreloadedAddresses: make(map[Address]struct{}),
                                },
                            },
                        },
                    },
                    Handler: Handler[Context[mockEXT, MockDB2], mockEXT, MockDB2]{
                        PreExecution:  mockPreExec,
                        Execution:     mockExec,
                        PostExecution: mockPostExec,
                        Cfg:           Config{specID: SpecID{version: 1}},
                    },
                }

                return evm, nil
            },
            wantErr: false,
        },
        {
            name:            "LoadAccounts fails",
            initialGasSpend: 21000,
            setupMocks: func() (*Evm[mockEXT, MockDB2], error) {
                mockPreExec := PreExecutionHandler[mockEXT, MockDB2]{
                    LoadAccounts: func(ctx *Context[mockEXT, MockDB2]) error {
                        return errors.New("failed to load accounts")
                    },
                }

                evm := &Evm[mockEXT, MockDB2]{
                    Handler: Handler[Context[mockEXT, MockDB2], mockEXT, MockDB2]{
                        PreExecution: mockPreExec,
                    },
                }

                return evm, nil
            },
            wantErr:        true,
            expectedErrMsg: "failed to load accounts",
        },
        {
            name:            "DeductCaller fails",
            initialGasSpend: 21000,
            setupMocks: func() (*Evm[mockEXT, MockDB2], error) {
                mockPreExec := PreExecutionHandler[mockEXT, MockDB2]{
                    LoadAccounts: func(ctx *Context[mockEXT, MockDB2]) error {
                        return nil
                    },
                    LoadPrecompiles: func() ContextPrecompiles[MockDB2] {
                        return ContextPrecompiles[MockDB2]{
                            Inner: PrecompilesCow[MockDB2]{
                                StaticRef: &Precompiles{
                                    Addresses: make(map[Address]struct{}),
                                },
                            },
                        }
                    },
                    DeductCaller: func(ctx *Context[mockEXT, MockDB2]) EVMResultGeneric[ResultAndState, DatabaseError] {
                        return EVMResultGeneric[ResultAndState, DatabaseError]{
                            Err: errors.New("insufficient funds"),
                        }
                    },
                }

                evm := &Evm[mockEXT, MockDB2]{
                    Handler: Handler[Context[mockEXT, MockDB2], mockEXT, MockDB2]{
                        PreExecution: mockPreExec,
                    },
                }

                return evm, nil
            },
            wantErr:        true,
            expectedErrMsg: "insufficient funds",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            evm, err := tt.setupMocks()
            if err != nil {
                t.Fatalf("Failed to setup test: %v", err)
            }

            result := evm.TransactPreverifiedInner(tt.initialGasSpend)

            if tt.wantErr {
                if result.Err == nil {
                    t.Errorf("TransactPreverifiedInner() error = nil, expected error with message: %v", tt.expectedErrMsg)
                } else if result.Err.Error() != tt.expectedErrMsg {
                    t.Errorf("TransactPreverifiedInner() error = %v, expected error: %v", result.Err, tt.expectedErrMsg)
                }
            } else if result.Err != nil {
                t.Errorf("TransactPreverifiedInner() unexpected error = %v", result.Err)
            }
        })
    }
}
func TestTransactPreverifiedInner_Success(t *testing.T) {
    config := HandlerCfg{
        specID: 1, // Adjust SpecID as needed; this is just a placeholder
    }

    // Initialize the handler with a function literal matching ValidateTxEnvAgainstState
    handler := Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{
        Cfg: config,
        Validation: ValidationHandler[MockEXT, MockDB2]{
            TxAgainstState: func(ctx *Context[MockEXT, MockDB2]) error {
                // Simulate failure
                return fmt.Errorf("Preverification failed")
            },
            InitialTxGas: func(env *Env) (uint64, error){
                return 0, fmt.Errorf("Preverification failed")
            },
            Env: func(env *Env) error{
                return fmt.Errorf("Preverification failed")
            },
        },
    }

    // Set up the environment variable for EvmContext
    var env *Env = &Env{
		Tx: TxEnv{
			TransactTo: TransactTo{
				Type: Call2,
				Address: &Address{}, // Provide a valid address if necessary.
			},
			GasLimit: 1000000, // Set an appropriate gas limit.
		},
	}// Ensure env is initialized properly

    inner := InnerEvmContext[MockDB2]{
        Env: env,
    }

    context := Context[MockEXT, MockDB2]{
        Evm: EvmContext[MockDB2]{
            Inner: inner,
        },
        External: MockEXT{},
    }

    // Initialize Evm instance with the handler and context
    evm := Evm[MockEXT, MockDB2]{
        Context: context,
        Handler: handler,
    }
    result := evm.TransactPreverifiedInner(1)
    assert.NoError(t, result.Err, "Expected no error for successful transaction")
}
/*
func TestTransactPreverifiedInner_Success(t *testing.T) {
	evm:=setupSuccessfulMockEvm()
	result := evm.TransactPreverifiedInner(1)
	assert.NoError(t, result.Err, "Expected no error for successful transaction")
}*/

func setupSuccessfulMockEvm() *Evm[MockEXT, MockDB2] {
	handlerCfg := HandlerCfg{specID: 1}
	mockContext := Context[MockEXT, MockDB2]{
		Evm: EvmContext[MockDB2]{
			Inner: InnerEvmContext[MockDB2]{
				Env: &Env{
					Tx: TxEnv{
						TransactTo: TxKind{Type: Call2},
						GasLimit: 1000000,
					},
				},
			},
		},
	}
	mockHandler := &Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{
		Cfg: handlerCfg,
		Validation: ValidationHandler[MockEXT, MockDB2]{
			Env: func(env *Env) error { return nil },
			InitialTxGas: func(env *Env) (uint64, error) { return 0, nil },
			TxAgainstState: func(ctx *Context[MockEXT, MockDB2]) error { return nil },
		},
		PreExecution: PreExecutionHandler[MockEXT, MockDB2]{
			LoadAccounts: func(ctx *Context[MockEXT, MockDB2]) error { return nil },
			DeductCaller: func(ctx *Context[MockEXT, MockDB2]) EVMResultGeneric[ResultAndState, DatabaseError] { 
				return EVMResultGeneric[ResultAndState, DatabaseError]{} 
			},
			LoadPrecompiles: func() ContextPrecompiles[MockDB2] { return ContextPrecompiles[MockDB2]{} },
		},
		Execution: ExecutionHandler[MockEXT, MockDB2]{
			Call: func(ctx *Context[MockEXT, MockDB2], inputs *CallInputs) (FrameOrResult, EvmError) {
				return FrameOrResult{
					Type: Result_FrameOrResult,
					Result: FrameResult{},
				}, EvmError{}
			},

			LastFrameReturn: func(ctx *Context[MockEXT, MockDB2], result *FrameResult) EvmError {return EvmError{}},
		},
		PostExecution: PostExecutionHandler[MockEXT, MockDB2]{
			ReimburseCaller: func(ctx *Context[MockEXT, MockDB2], gas *Gas) EVMResultGeneric[struct{}, any] { 
				return EVMResultGeneric[struct{}, any]{} 
			},
			RewardBeneficiary: func(ctx *Context[MockEXT, MockDB2], gas *Gas) EVMResultGeneric[struct{}, any] { 
				return EVMResultGeneric[struct{}, any]{} 
			},
			Output: func(ctx *Context[MockEXT, MockDB2], frameResult FrameResult) (ResultAndState, EvmError) { 
				return ResultAndState{}, EvmError{} 
			},
			End: func(ctx *Context[MockEXT, MockDB2], result EVMResultGeneric[ResultAndState, EvmError]) (ResultAndState, EvmError) {
				return ResultAndState{}, EvmError{}
			},
		},
	}

	return &Evm[MockEXT, MockDB2]{
		Context: mockContext,
		Handler: *mockHandler,
	}
}



func TestTransactPreverifiedInner_LoadAccountsError(t *testing.T) {
	e := setupMockEvm()
	e.Handler.PreExecution.LoadAccounts = func(ctx *Context[MockEXT, MockDB2]) error {
		return fmt.Errorf("LoadAccounts error")
	}
	result := e.TransactPreverifiedInner(1000)
	assert.Error(t, result.Err, "Expected error due to LoadAccounts failure")
	assert.Contains(t, result.Err.Error(), "LoadAccounts error")
}

func TestTransactPreverifiedInner_DeductCallerError(t *testing.T) {
	e := setupMockEvm()
	e.Handler.PreExecution.DeductCaller = func(ctx *Context[MockEXT,MockDB2]) EVMResultGeneric[ResultAndState, DatabaseError] {
		return EVMResultGeneric[ResultAndState, DatabaseError]{Err: fmt.Errorf("DeductCaller error")}
	}
	result := e.TransactPreverifiedInner(1000)
	assert.Error(t, result.Err, "Expected error due to DeductCaller failure")
	assert.Contains(t, result.Err.Error(), "DeductCaller error")
}

func TestTransactPreverifiedInner_CallError(t *testing.T) {
	e := setupMockEvm()
	e.Handler.Execution.Call = func(ctx *Context[MockEXT,MockDB2], inputs *CallInputs) (FrameOrResult, EvmError) {
		return FrameOrResult{}, EvmError{Message: "Call error"}
	}
	result := e.TransactPreverifiedInner(1000)
	assert.Error(t, result.Err, "Expected error due to Call failure")
	assert.Contains(t, result.Err.Error(), "Call error")
}

func TestTransactPreverifiedInner_OutputError(t *testing.T) {
	e := setupMockEvm()
	e.Handler.PostExecution.Output = func(ctx *Context[MockEXT,MockDB2], frameResult FrameResult) (ResultAndState, EvmError) {
		return ResultAndState{}, EvmError{Message: "Output error"}
	}
	result := e.TransactPreverifiedInner(1000)
	assert.Error(t, result.Err, "Expected error due to Output failure")
	assert.Contains(t, result.Err.Error(), "Output error")
}


// TestTransact_PreVerifyFailure simulates a failure in preverifyTransactionInner and expects an error result
func TestTransact_PreVerifyFailure(t *testing.T) {
	config := HandlerCfg{
		specID: 1, // Adjust SpecID as needed; this is just a placeholder
	}

	// Initialize the handler with a function literal matching ValidateTxEnvAgainstState
	handler := Handler[Context[MockEXT, MockDB2], MockEXT, MockDB2]{
		Cfg: config,
		Validation: ValidationHandler[MockEXT, MockDB2]{
			TxAgainstState: func(ctx *Context[MockEXT, MockDB2]) error {
				// Simulate failure
				return fmt.Errorf("Preverification failed")
			},
			InitialTxGas: func(env *Env) (uint64, error){
				return 0, fmt.Errorf("Preverification failed")
			},
			Env: func(env *Env) error{
				return fmt.Errorf("Preverification failed")
			},
		},
	}

	// Set up the environment variable for EvmContext
	var env *Env = &Env{} // Ensure env is initialized properly

	inner := InnerEvmContext[MockDB2]{
		Env: env,
	}

	context := Context[MockEXT, MockDB2]{
		Evm: EvmContext[MockDB2]{
			Inner: inner,
		},
		External: MockEXT{},
	}

	// Initialize Evm instance with the handler and context
	evm := Evm[MockEXT, MockDB2]{
		Context: context,
		Handler: handler,
	}

	// Ensure the Evm is not nil and initialized correctly
	assert.NotNil(t, evm, "Evm instance should not be nil")

	// Simulate a failure scenario by making the mock implementation return an error
	result := evm.Transact()

	// Verify that the result contains the expected error
	assert.NotNil(t, result.Err, "Expecting an error due to preverification failure")
	assert.Equal(t, "Preverification failed", result.Err.Error(), "Error message should match expected preverification failure")
}
