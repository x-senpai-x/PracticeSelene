package evm
type Context struct {
	Evm EvmContext
	External interface{}
}
type EvmContext struct {
	Inner InnerEvmContext
	Precompiles ContextPrecompiles
}
type InnerEvmContext struct {
    Env                    *Env
    JournaledState        JournaledState
    DB                     Database
    Error                  error
    ValidAuthorizations    []Address
    L1BlockInfo           *L1BlockInfo // For optimism feature
}
type Env struct{
    cfg CfgEnv
    block BlockEnv
    tx TxEnv
}
type CfgEnv struct {
	ChainID                       uint64         `json:"chain_id"`
	KzgSettings                   *EnvKzgSettings `json:"kzg_settings,omitempty"`
	PerfAnalyseCreatedBytecodes    AnalysisKind   `json:"perf_analyse_created_bytecodes"`
	LimitContractCodeSize         *uint64        `json:"limit_contract_code_size,omitempty"`
	MemoryLimit                   uint64         `json:"memory_limit,omitempty"` // Consider using a pointer if optional
	DisableBalanceCheck           bool           `json:"disable_balance_check,omitempty"`
	DisableBlockGasLimit          bool           `json:"disable_block_gas_limit,omitempty"`
	DisableEIP3607                bool           `json:"disable_eip3607,omitempty"`
	DisableGasRefund               bool           `json:"disable_gas_refund,omitempty"`
	DisableBaseFee                bool           `json:"disable_base_fee,omitempty"`
	DisableBeneficiaryReward      bool           `json:"disable_beneficiary_reward,omitempty"`
}
type BlockEnv struct {
	Number                  U256                       `json:"number"`
	Coinbase                Address                    `json:"coinbase"`
	Timestamp               U256                       `json:"timestamp"`
	GasLimit                U256                       `json:"gas_limit"`
	BaseFee                 U256                       `json:"basefee"`
	Difficulty              U256                       `json:"difficulty"`
	Prevrandao              *B256                      `json:"prevrandao,omitempty"`
	BlobExcessGasAndPrice   *BlobExcessGasAndPrice    `json:"blob_excess_gas_and_price,omitempty"`
}
type BlobExcessGasAndPrice struct {
	ExcessGas   uint64 `json:"excess_gas"`
	BlobGasPrice uint64 `json:"blob_gas_price"`
}

// EnvKzgSettings defines the KZG settings used in the environment.
type EnvKzgSettings struct {
	Mode   string                `json:"mode"`
	Custom *KzgSettings          `json:"custom,omitempty"`
}

// KzgSettings represents custom KZG settings.
type KzgSettings struct {
	// Define fields for KzgSettings based on your requirements.
}

// AnalysisKind represents the type of analysis for created bytecodes.
type AnalysisKind int

const (
	Raw AnalysisKind = iota
	Analyse
)

// MarshalJSON customizes the JSON marshalling for AnalysisKind.
func (ak AnalysisKind) MarshalJSON() ([]byte, error) {
	switch ak {
	case Raw:
		return json.Marshal("raw")
	case Analyse:
		return json.Marshal("analyse")
	default:
		return json.Marshal("unknown")
	}
}

// UnmarshalJSON customizes the JSON unmarshalling for AnalysisKind.
func (ak *AnalysisKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch s {
	case "raw":
		*ak = Raw
	case "analyse":
		*ak = Analyse
	default:
		return fmt.Errorf("unknown AnalysisKind: %s", s)
	}
	return nil
}