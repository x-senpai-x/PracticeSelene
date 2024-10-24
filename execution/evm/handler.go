package evm
import "github.com/BlocSoc-iitr/selene/common"
func NewHandler(cfg HandlerCfg){
	if cfg.isOptimism{
		Handler{

		}
	}
}
type HandlerCfg struct{
	specID specs.SpecId
	isOptimism bool
}
func NewHandlerCfg(specId SpecId, isOptimismEnabled bool) HandlerCfg {
    var isOptimism bool
    // Traditional if-else to set the isOptimism flag
    if isOptimismEnabled {
        isOptimism = true
    } else {
        isOptimism = false
    }
    return HandlerCfg{
        specID:     specId,
        isOptimism: isOptimism,
    }
}
type Handler struct{
	Cfg HandlerCfg
	InstructionTable InstructionTables
	Registers []HandleRegisters
	Validation ValidationHandler
	PreExecution PreExecutionHandler
	PostExecution PostExecutionHandler
	Execution ExecutionHandler
}