package evm
type Handler [H Host, EXT any, DB Database]struct{
	Cfg HandlerCfg
	InstructionTable InstructionTables[H]
	Registers []HandleRegisters[H, EXT, DB]
	Validation ValidationHandler[EXT, DB]
	PreExecution PreExecutionHandler[EXT, DB]
	PostExecution PostExecutionHandler[EXT, DB]
	Execution ExecutionHandler[EXT,DB]
}

type PostExecutionHandler[EXT any, DB Database] struct {
    ReimburseCaller      ReimburseCallerHandle[EXT, DB]
    RewardBeneficiary    RewardBeneficiaryHandle[EXT, DB]
    Output               OutputHandle[EXT, DB]
    End                  EndHandle[EXT, DB]
    Clear                ClearHandle[EXT, DB]
}

// ReimburseCallerHandle type definition.
type ReimburseCallerHandle[EXT any, DB Database] func(ctx *Context[EXT, DB], gas *Gas) EVMResultGeneric[struct{}, any]

// RewardBeneficiaryHandle type definition (same as ReimburseCallerHandle).
type RewardBeneficiaryHandle[EXT any, DB Database]  ReimburseCallerHandle[EXT, DB]

// OutputHandle type definition.
type OutputHandle[EXT any, DB Database] func(ctx *Context[EXT, DB], frameResult FrameResult) (ResultAndState, EvmError)

// EndHandle type definition.
type EndHandle[EXT any, DB Database] func(ctx *Context[EXT, DB], result EVMResultGeneric[ResultAndState, EvmError]) (ResultAndState, EvmError)

// ClearHandle type definition.
type ClearHandle[EXT any, DB Database] func(ctx *Context[EXT, DB])

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
func NewHandler[H Host, EXT any, DB Database](cfg HandlerCfg) Handler[H, EXT, DB] {
    return createHandlerWithConfig[H, EXT, DB](cfg)
}
func createHandlerWithConfig[H Host, EXT any, DB Database](cfg HandlerCfg) Handler[H, EXT, DB] {
    spec := getSpecForID[H, EXT, DB](cfg.specID)
    
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
type EvmHandler[H Host, EXT any, DB Database] Handler[H, EXT, DB]

// HandleRegister defines a function type that accepts a pointer to EvmHandler.
type HandleRegister[H Host, EXT any, DB Database] func(handler *EvmHandler[H, EXT, DB])

// HandleRegisterBox defines a function type as a boxed register.
type HandleRegisterBox[H Host, EXT any, DB Database] func(handler *EvmHandler[H, EXT, DB])

// HandleRegisters is an interface that defines the `Register` method.
type HandleRegisters[H Host, EXT any, DB Database] interface {
	Register(handler *EvmHandler[H, EXT, DB])
}

// PlainRegister struct implements HandleRegisters with a plain function register.
type PlainRegister[H Host, EXT any, DB Database] struct {
	RegisterFn HandleRegister[H, EXT, DB]
}

// BoxRegister struct implements HandleRegisters with a boxed function register.
type BoxRegister[H Host, EXT any, DB Database] struct {
	RegisterFn HandleRegisterBox[H, EXT, DB]
}

// Register method for PlainRegister, to satisfy HandleRegisters interface.
func (p PlainRegister[H, EXT, DB]) Register(handler *EvmHandler[H, EXT, DB]) {
	p.RegisterFn(handler)
}

// Register method for BoxRegister, to satisfy HandleRegisters interface.
func (b BoxRegister[H, EXT, DB]) Register(handler *EvmHandler[H, EXT, DB]) {
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