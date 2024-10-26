package evm

type LoadPrecompilesHandle[DB Database] func() ContextPrecompiles
type LoadAccountsHandle[EXT any, DB Database] func(ctx *Context) error
type DeductCallerHandle[EXT any, DB Database] func(ctx *Context)

type PreExecutionHandler[EXT any, DB Database] struct {
	LoadPrecompiles LoadPrecompilesHandle[DB]
	LoadAccounts    LoadAccountsHandle[EXT, DB]
	DeductCaller    DeductCallerHandle[EXT, DB]
}

func (p *PreExecutionHandler[EXT, DB]) LoadPrecompilesFunction() ContextPrecompiles {
	return p.LoadPrecompiles()
}