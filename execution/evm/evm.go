package evm

import (
	// "math/big"

	// "github.com/ethereum/go-ethereum/common"
)

type Evm[EXT interface{}, DB Database] struct {
	Context Context[EXT, DB]
	Handler Handler[Context[EXT, DB], EXT, DB]
}

func NewEvm(context Context , handler Handler) Evm {
	context.Evm.Inner.JournaledState.SetSpecId(handler.Cfg.specID)
	return Evm{
		Context: context,
		Handler: handler,
	}
}
/*
func (evm Evm) IntoContextWithHandlerCfg[EXT interface{}, DB Database]() ContextWithHandlerCfg[EXT, DB] {
    return ContextWithHandlerCfg[EXT, DB]{
        Context: evm.Context,
        Cfg:     evm.Handler.Cfg,
    }
}*/
func IntoContextWithHandlerCfg[EXT interface{}, DB Database](evm Evm) ContextWithHandlerCfg[EXT, DB] {
    return ContextWithHandlerCfg[EXT, DB]{
        Context: evm.Context,
        Cfg:     evm.Handler.Cfg,
    }
}
type EvmResult struct{}
// type EvmError struct{}

var EOF_MAGIC_BYTES = []byte{0xef, 0x00}

func (e *Evm) Transact() EvmResult {
	initialGasSpend, err := e.preverifyTransactionInner()
	if err != nil {
		e.clear()
		// return e
	}
	output, err := e.TransactPreverifiedInner(initialGasSpend)
	output = e.Handler.PostExecution.End(output)
	e.clear()
	return output, nil
}

func (e *Evm) preverifyTransactionInner() (uint64, EvmError) {
	err := e.Handler.Validation.Env(e.Context.Evm.Inner.Env) //?
	initialGasSpend, err := e.Handler.Validation.InitialTxGas(e.Context.Evm.Inner.Env)
	err = e.Handler.Validation.TxAgainstState(e.Context)
	return initialGasSpend, EvmError{}
}

func (e *Evm) clear() {
	e.Handler.PostExecution.Clear(e.Context)
}

func (e *Evm) TransactPreverifiedInner(initialGasSpend uint64) (EvmResult, DatabaseError) {
	specId := e.Handler.Cfg.specID
	ctx := &e.Context
	preExec := e.Handler.PreExecution

	err := preExec.LoadAccounts(ctx)

	precompiles := preExec.LoadPrecompilesFunction()
	ctx.Evm.SetPrecompiles(precompiles)

	err = preExec.DeductCaller(ctx)
	gasLimit := ctx.Evm.Inner.Env.tx.gasLimit - initialGasSpend

	exec := e.Handler.Execution
	firstFrameOrResult := func() FrameOrResult {
		if ctx.Evm.Inner.Env.Tx.TransactTo.Call != nil {
			return exec.Call(ctx, (&CallInputs{}).NewBoxed(ctx.Evm.Inner.Env.tx, gasLimit))
		} else {
			if specId.IsEnabledIn(PRAGUE_EOF) && ctx.Evm.Inner.Env.tx.data.get(2) == EOF_MAGIC_BYTES {
				return exec.EofCreate(ctx, EOFCreateInputs{}.NewTx(ctx.Evm.Inner.Env.tx, gasLimit))
			} else {
				return exec.Create(ctx, (&CreateInputs{}).NewBoxed(ctx.Evm.Inner.Env.tx, gasLimit))
			}
		}
	}()

	var result FrameResult
	if firstFrameOrResult == Frame {
		result, err = e.RunTheLoop(Frame.FirstFrame)
	} else {
		result = firstFrameOrResult
	}

	ctx = &e.Context
	err = e.Handler.Execution.LastFrameReturn(ctx, &result)

	postExec := e.Handler.PostExecution
	err = postExec.ReimburseCaller(ctx, result.Gas)
	err = postExec.RewardBeneficiary(ctx, result.Gas)

	return postExec.Output(ctx, result)
}

func (e *Evm) RunTheLoop(firstFrame Frame) (FrameResult, DatabaseError) {
	callStack := make([]Frame, 1025)
	callStack = append(callStack, firstFrame)
	
}


func (s SpecId) IsEnabledIn(spec SpecId) bool {
	return s.Enabled(s, spec)
}

func (s SpecId) Enabled(a SpecId, b SpecId) bool {
	return uint8(a) >= uint8(b)
}

// String method to provide a string representation for each CallScheme variant.
func (cs CallScheme) String() string {
	switch cs {
	case Call:
		return "Call"
	case CallCode:
		return "CallCode"
	case DelegateCall:
		return "DelegateCall"
	case StaticCall:
		return "StaticCall"
	case ExtCall:
		return "ExtCall"
	case ExtStaticCall:
		return "ExtStaticCall"
	case ExtDelegateCall:
		return "ExtDelegateCall"
	default:
		return "Unknown"
	}
}

func New(txEnv *TxEnv, gasLimit uint64) *CallInputs {
	// Check if the transaction kind is Call and extract the target address.
	if txEnv.TransactTo != Call {
		return nil
	}

	// Create and return the CallInputs instance.
	return &CallInputs{
		Input:              txEnv.Data,
		GasLimit:           gasLimit,
		TargetAddress:      txEnv.TargetAddress,
		BytecodeAddress:    txEnv.TargetAddress, // Set to target_address as in Rust code.
		Caller:             txEnv.Caller,
		Value:              CallValue{Type: Transfer, Value: txEnv.Value},
		Scheme:             Call, // Assuming CallScheme.Call is represented by Call.
		IsStatic:           false,
		IsEof:              false,
		ReturnMemoryOffset: Range{Start: 0, End: 0},
	}
}

func (c *CallInputs) NewBoxed(txEnv *TxEnv, gasLimit uint64) *CallInputs {
	return New(txEnv, gasLimit) // Returns a pointer or nil, similar to Option<Box<Self>> in Rust.
}

func (e EOFCreateInputs) NewTx(tx *TxEnv, gasLimit uint64) EOFCreateInputs {
	return EOFCreateInputs{
		Caller:   tx.Caller,
		Value:    tx.Value,
		GasLimit: gasLimit,
		Kind: EOFCreateKind{
			Kind:     Tx,
			InitData: tx.Data,
		},
	}
}

func (c *CreateInputs) New(txEnv *TxEnv, gasLimit uint64) *CreateInputs {
	if txEnv.TransactTo != Create {
		return nil
	}
	return &CreateInputs{
		Caller: txEnv.Caller,
		Scheme: CreateScheme{
			Kind: Create,
		},
		Value: txEnv.Value,
		InitCode: txEnv.Data,
		GasLimit: gasLimit,
	}
}

func (c *CreateInputs) NewBoxed(txEnv *TxEnv, gasLimit uint64) *CreateInputs {
	return c.New(txEnv, gasLimit)
}