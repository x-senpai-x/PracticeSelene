package evm
type Evm struct{
	Context Context
	Handler Handler
}
func NewEvm(context Context , handler Handler) Evm {
	context.Evm.Inner.JournaledState.SetSpecId(handler.Cfg.specID)
	return Evm{
		Context: context,
		Handler: handler,
	}
}
func (evm Evm) IntoContextWithHandlerCfg() ContextWithHandlerCfg {
	return ContextWithHandlerCfg{
		Context: evm.Context,
		Cfg: evm.Handler.Cfg,
	}
}