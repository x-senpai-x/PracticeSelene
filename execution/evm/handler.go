package evm
type Handler struct{
	Cfg HandlerCfg
	InstructionTable InstructionTables
	Registers []HandleRegisters
	Validation ValidationHandler
	PreExecution PreExecutionHandler
	PostExecution PostExecutionHandler
	Execution ExecutionHandler
}
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
