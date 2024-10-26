package evm

import (
	"crypto/ecdsa"
	// "crypto/elliptic"
	"encoding/json"
	"fmt"
	"math/big"
)

// Signature represents an ECDSA signature consisting of V, R, and S values
type Signature struct {
    V uint8    // Recovery ID
    R *big.Int // R component of signature
    S *big.Int // S component of signature
}

// NewSignature creates a new ECDSA signature from V, R, and S values
func NewSignature(v uint8, r, s *big.Int) *Signature {
    return &Signature{
        V: v,
        R: new(big.Int).Set(r),
        S: new(big.Int).Set(s),
    }
}

// ToRawSignature converts the signature to R || S format
func (s *Signature) ToRawSignature() []byte {
    rBytes := s.R.Bytes()
    sBytes := s.S.Bytes()
    
    // Ensure each component is 32 bytes
    signature := make([]byte, 64)
    copy(signature[32-len(rBytes):32], rBytes)
    copy(signature[64-len(sBytes):], sBytes)
    
    return signature
}

// FromRawSignature creates a Signature from a raw 64-byte R || S format
func FromRawSignature(data []byte, v uint8) (*Signature, error) {
    if len(data) != 64 {
        return nil, fmt.Errorf("invalid signature length: got %d, want 64", len(data))
    }

    r := new(big.Int).SetBytes(data[:32])
    s := new(big.Int).SetBytes(data[32:])

    return &Signature{
        V: v,
        R: r,
        S: s,
    }, nil
}

// Verify verifies the signature against a message hash and public key
func (s *Signature) Verify(pubKey *ecdsa.PublicKey, hash []byte) bool {
    // Check if r and s are in the valid range
    if s.R.Sign() <= 0 || s.S.Sign() <= 0 {
        return false
    }
    if s.R.Cmp(pubKey.Params().N) >= 0 || s.S.Cmp(pubKey.Params().N) >= 0 {
        return false
    }

    // Verify the signature
    return ecdsa.Verify(pubKey, hash, s.R, s.S)
}
type Env struct{
    Cfg CfgEnv
    Block BlockEnv
    Tx TxEnv
}
func NewEnv() *Env {
    return &Env{} // Returns a pointer to a zero-initialized struct
}
type TxEnv struct {
	Caller Address
	Gas_limit uint64
	Gas_price U256
	Transact_to TxKind
	Value U256
	Data Bytes
	Nonce *uint64
	Chain_id *uint64
	Access_list []AccessListItem
	Gas_priority_fee *U256
	Blob_hashes []B256
	Max_fee_per_blob_gas *U256
	Authorization_list *AuthorizationList
	Optimism OptimismFields//flag on optimism
}
type AuthorizationListType int

const (
	Signed AuthorizationListType = iota
	Recovered
)
type AuthorizationList struct {
	Type              AuthorizationListType
	SignedAuthorizations   []SignedAuthorization
	RecoveredAuthorizations []RecoveredAuthorization
}
type ChainID uint64
type OptionalNonce struct {
	Nonce *uint64 // nil if no nonce is provided
}
type Authorization struct {
	ChainID ChainID    // The chain ID of the authorization
	Address Address    // The address of the authorization
	Nonce   OptionalNonce // The nonce for the authorization
}
type SignedAuthorization struct {
	Inner     Authorization // Embedded authorization data
	Signature Signature     // Signature associated with authorization
}

type RecoveredAuthorization struct {
	Inner     Authorization // Embedded authorization data
	Authority *Address      // Optional authority address
}

type TransactTo = TxKind;

type TxKind struct {
	Type    TxKindType
	Address *Address // Address is nil for Create type
}
type TxKindType int

const (
	Create TxKindType = iota
	Call2
)
type AccessListItem struct {
	Address Address
	StorageKeys []B256
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