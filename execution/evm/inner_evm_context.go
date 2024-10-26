package evm
type InnerEvmContext[DB Database] struct {
    Env                    *Env
    JournaledState        JournaledState
    DB                     DB
    Error                  error
    ValidAuthorizations    []Address
    L1BlockInfo           *L1BlockInfo // For optimism feature : needs to be implemented 
}
func NewInnerEvmContext[DB Database](db DB) InnerEvmContext[DB] {
	return InnerEvmContext[DB]{
		Env: &Env{},
		JournaledState: nil,//to be changed
		DB: db,
		Error: nil,
		ValidAuthorizations: nil,
		L1BlockInfo: nil,
	}
}
/*
func (inner InnerEvmContext[DB])WithDB[ODB Database](newdb ODB ) InnerEvmContext[ODB]{
	return InnerEvmContext[ODB]{
		Env: inner.Env,
		JournaledState: inner.JournaledState,
		DB: db,
		Error: inner.Error,
		ValidAuthorizations: nil,
		L1BlockInfo: inner.L1BlockInfo,
	}
}*/
func (ctx InnerEvmContext[DB]) WithDB[ODB Database](newDb ODB) InnerEvmContext[ODB] {
    return InnerEvmContext[ODB]{
        env:                 ctx.env,
        journaledState:     ctx.JournaledState,
        db:                 newDb,
        error:              ctx.Error
        validAuthorizations: make(map[string]bool), // equivalent to Default::default() or nil // doubt
        l1BlockInfo:        ctx.l1BlockInfo,
    }
}