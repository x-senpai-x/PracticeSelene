package evm
import (
	"encoding/json"
	"fmt"
	"sync"
)
type Context struct {
	Evm EvmContext
	External interface{}
}
func DefaultContext() Context {
	return Context{
		Evm: NewEvmContext(),
		External: nil,
	}
}
type EvmContext struct {
	Inner InnerEvmContext
	Precompiles ContextPrecompiles
}
func NewEvmContext() EvmContext {
	return EvmContext{
		Inner: NewInnerEvmContext(db),
		Precompiles: DefaultContextPrecompiles(),
	}
}
type InnerEvmContext struct {
    Env                    *Env
    JournaledState        JournaledState
    DB                     Database
    Error                  error
    ValidAuthorizations    []Address
    L1BlockInfo           *L1BlockInfo // For optimism feature
}
//To be reviewed
func NewInnerEvmContext(db Database) InnerEvmContext {
	return InnerEvmContext{
		Env: NewEnv(),
		JournaledState: nil,//to be changed
		DB: db,
		Error: nil,
		ValidAuthorizations: nil,
		L1BlockInfo: nil,
	}
}
type ContextPrecompiles struct {
	Inner PrecompilesCow
}
func NewContextPrecompiles() *ContextPrecompiles {
	return &ContextPrecompiles{
		Inner: 
	}
}

type PrecompilesCow struct {
	isStatic bool
	StaticRef *Precompiles
	Owned     map[Address]ContextPrecompile
}
func NewPrecompilesCow () PrecompilesCow{
	return PrecompilesCow{
		Owned: make(map[Address]ContextPrecompile),
	}

}
type Precompiles struct {
	Inner map[Address]Precompile
	Addresses map[Address]struct{}
}
type Precompile struct {
	precompileType string // "Standard", "Env", "Stateful", or "StatefulMut"
	standard       StandardPrecompileFn
	env           EnvPrecompileFn
	stateful      *StatefulPrecompileArc
	statefulMut   *StatefulPrecompileBox
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
//Doubt
type StatefulPrecompileArc struct {
	sync.RWMutex
	impl StatefulPrecompile
}
// StatefulPrecompileBox is a mutable reference to a StatefulPrecompileMut
type StatefulPrecompileBox struct {
	impl StatefulPrecompileMut
}
type ContextPrecompile struct {
	precompileType string // "Ordinary", "ContextStateful", or "ContextStatefulMut"
	ordinary       *Precompile
	contextStateful      *ContextStatefulPrecompileArc
	contextStatefulMut   *ContextStatefulPrecompileBox
}
// ContextStatefulPrecompileArc is a thread-safe reference to a ContextStatefulPrecompile
type ContextStatefulPrecompileArc struct {
	sync.RWMutex
	impl ContextStatefulPrecompile
}
// ContextStatefulPrecompileBox is a mutable reference to a ContextStatefulPrecompileMut
type ContextStatefulPrecompileBox struct {
	impl ContextStatefulPrecompileMut       
}
type ContextStatefulPrecompile interface {
	Call(bytes *Bytes, gasLimit uint64, evmCtx *InnerEvmContext) PrecompileResult
}
// ContextStatefulPrecompileMut interface for mutable stateful precompiles with context
type ContextStatefulPrecompileMut interface {
	CallMut(bytes *Bytes, gasLimit uint64, evmCtx *InnerEvmContext) PrecompileResult
	Clone() ContextStatefulPrecompileMut
}
type PrecompileResult struct {
	output *PrecompileOutput
	err    PrecompileError//Doubt
}
type PrecompileOutput struct {
	GasUsed uint64
	Bytes   []byte
}
type PrecompileError struct {
	errorType string
	message   string
}
const (
	ErrorOutOfGas                  = "OutOfGas"
	ErrorBlake2WrongLength         = "Blake2WrongLength"
	ErrorBlake2WrongFinalFlag      = "Blake2WrongFinalIndicatorFlag"
	ErrorModexpExpOverflow         = "ModexpExpOverflow"
	ErrorModexpBaseOverflow        = "ModexpBaseOverflow"
	ErrorModexpModOverflow         = "ModexpModOverflow"
	ErrorBn128FieldPointNotMember  = "Bn128FieldPointNotAMember"
	ErrorBn128AffineGFailedCreate  = "Bn128AffineGFailedToCreate"
	ErrorBn128PairLength           = "Bn128PairLength"
	ErrorBlobInvalidInputLength    = "BlobInvalidInputLength"
	ErrorBlobMismatchedVersion     = "BlobMismatchedVersion"
	ErrorBlobVerifyKzgProofFailed  = "BlobVerifyKzgProofFailed"
	ErrorOther                     = "Other"
)

// Error implementation for PrecompileError
func (e PrecompileError) Error() string {
	if e.message != "" {
		return e.message
	}
	return e.errorType
}



