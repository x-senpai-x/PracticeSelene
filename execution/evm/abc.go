// +build optimism_default_handler,!negate_optimism_default_handler

package evm

type SpecId int

const LATEST SpecId = 0 // Equivalent to SpecId::LATEST in Rust

type HandlerCfg struct {
	SpecId     SpecId
	IsOptimism bool
}

// NewHandlerCfg creates a new HandlerCfg instance with feature flags
func NewHandlerCfg(specId SpecId) *HandlerCfg {
	isOptimism := false

	// Check if build tags enable Optimism by default
	// For `optimism_default_handler` with no `negate_optimism_default_handler` tag, set IsOptimism to true
	// Otherwise, defaults to false
	if IsOptimismEnabled() {
		isOptimism = true
	}

	return &HandlerCfg{
		SpecId:     specId,
		IsOptimism: isOptimism,
	}
}

// IsOptimismEnabled checks if the feature flag for Optimism is enabled
// Build tag-based conditional compilation
// +build !negate_optimism_default_handler
func IsOptimismEnabled() bool {
	return true
}

// Fallback implementation without specific feature flags
// +build !optimism_default_handler
func IsOptimismEnabled() bool {
	return false
}

type EvmBuilder struct {
	Context   Context
	Handler   *HandlerCfg
	Phantom   interface{}
}

// NewEvmBuilder initializes a default EvmBuilder instance
func NewEvmBuilder() *EvmBuilder {
	handlerCfg := NewHandlerCfg(LATEST)

	return &EvmBuilder{
		Context: Context{},
		Handler: handlerCfg,
		Phantom: nil, // No need for PhantomData equivalent in Go
	}
}

type Context struct{}