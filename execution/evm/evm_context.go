package evm
func (e *EvmContext) SetPrecompiles(precompiles ContextPrecompiles) {
	for i, address := range precompiles.Inner.StaticRef.Addresses {
		e.Inner.JournaledState.WarmPreloadedAddresses[i] = address
	}
	e.Precompiles = precompiles
}

type EvmContext[DB Database] struct {
	Inner InnerEvmContext [DB]
	Precompiles ContextPrecompiles
}
func NewEvmContext[DB Database](db DB) EvmContext[DB] {
	return EvmContext[DB]{
		Inner: NewInnerEvmContext(db),
		Precompiles: DefaultContextPrecompiles(),
	}
}
/*
func (c EvmContext[DB])WithDB(db Database) EvmContext[DB]{
	return EvmContext[DB]{
		Inner: c.Inner.WithDB(db),
		Precompiles: DefaultContextPrecompiles(),
	}
}*/
func (ctx EvmContext[DB]) WithDB[ODB Database](db ODB) EvmContext[ODB] {
    return EvmContext[ODB]{
        inner:       ctx.inner.WithDB(db),
        precompiles: NewContextPrecompiles(),
    }
}