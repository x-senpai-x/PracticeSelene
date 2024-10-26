package evm

type EvmContext struct {
	Inner InnerEvmContext
	Precompiles ContextPrecompiles
}
func NewEvmContext() EvmContext {
	return EvmContext{
		Inner: NewInnerEvmContext(),
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

func (e *EvmContext) SetPrecompiles(precompiles ContextPrecompiles) {
	for i, address := range precompiles.Inner.StaticRef.Addresses {
		e.Inner.JournaledState.WarmPreloadedAddresses[i] = address
	}
	e.Precompiles = precompiles
}