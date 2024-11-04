package evm

import (
	"sync"
)

func (e *EvmContext[DB]) SetPrecompiles(precompiles ContextPrecompiles[DB]) {
	// Check if StaticRef is nil before accessing its Addresses
	if precompiles.Inner.StaticRef == nil || len(precompiles.Inner.StaticRef.Addresses) == 0 {
		precompiles.Inner.StaticRef = &Precompiles{}
		precompiles.Inner.StaticRef.Addresses = make(map[Address]struct{})
	}

	for address := range precompiles.Inner.StaticRef.Addresses {
		e.Inner.JournaledState.WarmPreloadedAddresses[address] = struct{}{}
	}
	e.Precompiles = precompiles
}

type ContextPrecompiles[DB Database] struct {
	Inner PrecompilesCow[DB]
}

func DefaultContextPrecompiles[DB Database]() ContextPrecompiles[DB] {
	return ContextPrecompiles[DB]{
		Inner: NewPrecompilesCow[DB](),
	}
}
func NewPrecompilesCow[DB Database]() PrecompilesCow[DB] {
	return PrecompilesCow[DB]{
		Owned: make(map[Address]ContextPrecompile[DB]),
	}
}

type PrecompilesCow[DB Database] struct {
	IsStatic  bool
	StaticRef *Precompiles
	Owned     map[Address]ContextPrecompile[DB]
}
type Precompiles struct {
	Inner     map[Address]Precompile
	Addresses map[Address]struct{}
}
type Precompile struct {
	PrecompileType string // "Standard", "Env", "Stateful", or "StatefulMut"
	Standard       StandardPrecompileFn
	Env            EnvPrecompileFn
	Stateful       *StatefulPrecompileArc
	StatefulMut    *StatefulPrecompileBox
}
type StandardPrecompileFn func(input *Bytes, gasLimit uint64) PrecompileResult
type EnvPrecompileFn func(input *Bytes, gasLimit uint64, env *Env) PrecompileResult
type StatefulPrecompile interface {
	Call(bytes *Bytes, gasLimit uint64, env *Env) PrecompileResult
}
type StatefulPrecompileMut interface {
	CallMut(bytes *Bytes, gasLimit uint64, env *Env) PrecompileResult
	Clone() StatefulPrecompileMut
}

// Doubt
type StatefulPrecompileArc struct {
	sync.RWMutex
	Impl StatefulPrecompile
}

// StatefulPrecompileBox is a mutable reference to a StatefulPrecompileMut
type StatefulPrecompileBox struct {
	Impl StatefulPrecompileMut
}
type ContextPrecompile[DB Database] struct {
	PrecompileType     string // "Ordinary", "ContextStateful", or "ContextStatefulMut"
	Ordinary           *Precompile
	ContextStateful    *ContextStatefulPrecompileArc[DB]
	ContextStatefulMut *ContextStatefulPrecompileBox[DB]
}

// ContextStatefulPrecompileArc is a thread-safe reference to a ContextStatefulPrecompile
type ContextStatefulPrecompileArc[DB Database] struct {
	sync.RWMutex
	Impl ContextStatefulPrecompile[DB]
}

// ContextStatefulPrecompileBox is a mutable reference to a ContextStatefulPrecompileMut
type ContextStatefulPrecompileBox[DB Database] struct {
	Impl ContextStatefulPrecompileMut[DB]
}
type ContextStatefulPrecompile[DB Database] interface {
	Call(bytes *Bytes, gasLimit uint64, evmCtx *InnerEvmContext[DB]) PrecompileResult
}

// ContextStatefulPrecompileMut interface for mutable stateful precompiles with context
type ContextStatefulPrecompileMut[DB Database] interface {
	CallMut(bytes *Bytes, gasLimit uint64, evmCtx *InnerEvmContext[DB]) PrecompileResult
	Clone() ContextStatefulPrecompileMut[DB]
}
type PrecompileResult struct {
	Output *PrecompileOutput
	Err    PrecompileErrorStruct //Doubt
}
type PrecompileOutput struct {
	GasUsed uint64
	Bytes   []byte
}
type PrecompileErrorStruct struct {
	ErrorType string
	Message   string
}

const (
	ErrorOutOfGas                 = "OutOfGas"
	ErrorBlake2WrongLength        = "Blake2WrongLength"
	ErrorBlake2WrongFinalFlag     = "Blake2WrongFinalIndicatorFlag"
	ErrorModexpExpOverflow        = "ModexpExpOverflow"
	ErrorModexpBaseOverflow       = "ModexpBaseOverflow"
	ErrorModexpModOverflow        = "ModexpModOverflow"
	ErrorBn128FieldPointNotMember = "Bn128FieldPointNotAMember"
	ErrorBn128AffineGFailedCreate = "Bn128AffineGFailedToCreate"
	ErrorBn128PairLength          = "Bn128PairLength"
	ErrorBlobInvalidInputLength   = "BlobInvalidInputLength"
	ErrorBlobMismatchedVersion    = "BlobMismatchedVersion"
	ErrorBlobVerifyKzgProofFailed = "BlobVerifyKzgProofFailed"
	ErrorOther                    = "Other"
)

// Error implementation for PrecompileError
func (e PrecompileErrorStruct) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.ErrorType
}