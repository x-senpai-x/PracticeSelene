package evm
type Handler [EXT any, DB Database]struct{
	Cfg HandlerCfg
	InstructionTable InstructionTables
	Registers []HandleRegisters[EXT, DB]
	Validation ValidationHandler[EXT, DB]
	PreExecution PreExecutionHandler[EXT, DB]
	PostExecution PostExecutionHandler[EXT, DB]
	Execution ExecutionHandler[EXT,DB]
}

type PostExecutionHandler[EXT any, DB any] struct {
    ReimburseCaller      ReimburseCallerHandle[EXT, DB]
    RewardBeneficiary    RewardBeneficiaryHandle[EXT, DB]
    Output               OutputHandle[EXT, DB]
    End                  EndHandle[EXT, DB]
    Clear                ClearHandle[EXT, DB]
}

// ReimburseCallerHandle type definition.
type ReimburseCallerHandle[EXT any, DB any] func(ctx *Context[EXT, DB], gas *Gas) EVMResultGeneric[struct{}, any]

// RewardBeneficiaryHandle type definition (same as ReimburseCallerHandle).
type RewardBeneficiaryHandle[EXT any, DB any]  ReimburseCallerHandle[EXT, DB]

// OutputHandle type definition.
type OutputHandle[EXT any, DB any] func(ctx *Context[EXT, DB], frameResult FrameResult) (ResultAndState, EVMError[any])

// EndHandle type definition.
type EndHandle[EXT any, DB any] func(ctx *Context[EXT, DB], result EVMResultGeneric[ResultAndState, EVMError[any]]) (ResultAndState, EVMError[any])

// ClearHandle type definition.
type ClearHandle[EXT any, DB any] func(ctx *Context[EXT, DB])

type LoadPrecompilesHandle[DB Database] func() ContextPrecompiles[DB]
type LoadAccountsHandle[EXT any, DB Database] func(ctx *Context[EXT, DB]) error
type DeductCallerHandle[EXT any, DB Database] func(ctx *Context[EXT, DB]) EVMResultGeneric[any ,  any]
type EVMResultGeneric[T any, DBError any] struct {
    Value T
    Err   error
}
type PreExecutionHandler[EXT any, DB Database] struct {
    LoadPrecompiles LoadPrecompilesHandle[DB]
    LoadAccounts    LoadAccountsHandle[EXT, DB]
    DeductCaller    DeductCallerHandle[EXT, DB]
}
type ValidationHandler[EXT any, DB Database] struct {
    InitialTxGas   ValidateInitialTxGasHandle[DB]
    TxAgainstState ValidateTxEnvAgainstState[EXT, DB]
    Env            ValidateEnvHandle[DB]
}
type ValidateEnvHandle[DB Database] func(env *Env) error
type ValidateTxEnvAgainstState[EXT any, DB Database] func(ctx *Context[EXT, DB]) error
type ValidateInitialTxGasHandle[DB Database] func(env *Env) (uint64, error)
func NewHandler(cfg HandlerCfg) Handler {
    return createHandlerWithConfig(cfg)
}
func createHandlerWithConfig(cfg HandlerCfg) Handler {
    spec := getSpecForID(cfg.specID)
    
    if cfg.isOptimism {
        return createOptimismHandler(spec)
    }
    return createMainnetHandler(spec)
}
type HandlerCfg struct{
	specID SpecId
	isOptimism bool
}
func NewHandlerCfg(specID SpecId) HandlerCfg {
    return HandlerCfg{
        specID:     specID,
        isOptimism: getDefaultOptimismSetting(),
    }
}
// EvmHandler is a type alias for Handler with specific type parameters.
type EvmHandler[EXT any, DB Database] Handler[EXT, DB]

// HandleRegister defines a function type that accepts a pointer to EvmHandler.
type HandleRegister[EXT any, DB Database] func(handler *EvmHandler[EXT, DB])

// HandleRegisterBox defines a function type as a boxed register.
type HandleRegisterBox[EXT any, DB Database] func(handler *EvmHandler[EXT, DB])

// HandleRegisters is an interface that defines the `Register` method.
type HandleRegisters[EXT any, DB Database] interface {
	Register(handler *EvmHandler[EXT, DB])
}

// PlainRegister struct implements HandleRegisters with a plain function register.
type PlainRegister[EXT any, DB Database] struct {
	RegisterFn HandleRegister[EXT, DB]
}

// BoxRegister struct implements HandleRegisters with a boxed function register.
type BoxRegister[EXT any, DB Database] struct {
	RegisterFn HandleRegisterBox[EXT, DB]
}

// Register method for PlainRegister, to satisfy HandleRegisters interface.
func (p PlainRegister[EXT, DB]) Register(handler *EvmHandler[EXT, DB]) {
	p.RegisterFn(handler)
}

// Register method for BoxRegister, to satisfy HandleRegisters interface.
func (b BoxRegister[EXT, DB]) Register(handler *EvmHandler[EXT, DB]) {
	b.RegisterFn(handler)
}
/*
type HandleRegister[EXT any, DB Database] func(handler *EvmHandler[EXT, DB])
type EvmHandler[EXT any, DB Database] Handler[EXT, DB]
type HandleRegisterBox[EXT any, DB Database] interface {
    Call(handler *EvmHandler[EXT, DB])
}
type HandleRegisters[EXT any, DB Database] struct {
    Plain HandleRegister[EXT, DB]      // Plain function register
    Box   HandleRegisterBox[EXT, DB]   // Boxed function register
}*/