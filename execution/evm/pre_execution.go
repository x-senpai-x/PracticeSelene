package evm

type LoadPrecompilesHandle[DB Database] func() ContextPrecompiles[DB]
type LoadAccountsHandle[EXT any, DB Database] func(ctx *Context[EXT, DB]) error
type DeductCallerHandle[EXT any, DB Database] func(ctx *Context[EXT, DB])

type PreExecutionHandler[EXT any, DB Database] struct {
	LoadPrecompiles LoadPrecompilesHandle[DB]
	LoadAccounts    LoadAccountsHandle[EXT, DB]
	DeductCaller    DeductCallerHandle[EXT, DB]
}

func (p *PreExecutionHandler[EXT, DB]) LoadPrecompilesFunction() ContextPrecompiles[DB] {
	return p.LoadPrecompiles()
}