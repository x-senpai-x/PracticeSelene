package evm
import(
	"os"
	"github.com/BlocSoc-iitr/selene/common"
	Common "github.com/ethereum/go-ethereum/common"
	"math/big"
)
type BlockTag = common.BlockTag
type U256 = *big.Int
type B256 = Common.Hash
type Address = Common.Address
type EvmBuilder struct {
	context Context
	handler Handler
	phantom struct{} // equivalent of PhantomData in Rust
}
func DefaultEvmBuilder(isOptimismEnabled bool) EvmBuilder {
    var handlerCfg HandlerCfg
    if isOptimismEnabled {
        handlerCfg = NewHandlerCfg(LATEST, true) // Set is_optimism to true
    } else {
        handlerCfg = NewHandlerCfg(LATEST, false)
    }
    return EvmBuilder{
        context: nil, // Replace with actual context initialization
        handler: 
        phantom: nil, 
    }
}

func DefaultContext() Context {
	return Context{
		Evm: NewEvmContext(),
		External: nil,
	}
}
func NewEvmContext() EvmContext {
	return EvmContext{
		Inner: NewInnerEvmContext(db),
		Precompiles: DefaultContextPrecompiles(),
	}
}
func NewInnerEvmContext(db Database) InnerEvmContext {
	return InnerEvmContext{
		Env: *Env,
		JournaledState: nil,
		DB: db,
		Error: nil,
		ValidAuthorizations: nil,
		L1BlockInfo: nil,
	}
}
func NewEnv() *Env {
    return new(Env) // Returns a pointer to a zero-initialized struct
}
func NewJournalState(spec specs.SpecId,warmPreloadedAddresses map[common.Address]struct{}) JournaledState {
	return JournaledState{
		State: nil,
		TransientStorage: nil,
		Journal: [][]JournalEntry{},
		Spec: spec,
	}
}

type JournaledState struct {
	State EvmState
	TransientStorage TransientStorage
	Logs []Log 
	Depth uint
	Journal [][]JournalEntry
	Spec specs.SpecId
	WarmPreloadedAddresses map[common.Address]struct{}
}
type EvmState map [common.Address]Account
type TransientStorage map[Key]U256
//Tuple implementation in golang:
type Key struct {
	Account common.Address
	Slot U256
}
type Log[T any] struct {
	// The address which emitted this log.
	Address Address `json:"address"`
	// The log data.
	Data T `json:"data"`
}
// LogData represents the data structure of a log entry.
type LogData struct {
	// The indexed topic list.
	Topics []B256 `json:"topics"`
	// The plain data.
	Data Bytes `json:"data"`
}
type EvmStorage map[U256]EvmStorageSlot
type EvmStorageSlot struct {
	OriginalValue U256
	PresentValue U256
	IsCold bool
}
type JournalEntryType uint8
const (
	AccountWarmedType JournalEntryType = iota
	AccountDestroyedType
	AccountTouchedType
	BalanceTransferType
	NonceChangeType
	AccountCreatedType
	StorageChangedType
	StorageWarmedType
	TransientStorageChangeType
	CodeChangeType
)
//Creating a struct for JournalEntry containing all possible fields that might be needed
type JournalEntry struct {
	Type         JournalEntryType
	Address      Address
	Target       Address  // Used for AccountDestroyed
	WasDestroyed bool    // Used for AccountDestroyed
	HadBalance   U256 // Used for AccountDestroyed
	Balance      U256 // Used for BalanceTransfer
	From         Address  // Used for BalanceTransfer
	To           Address  // Used for BalanceTransfer
	Key          U256// Used for Storage operations
	HadValue     U256 // Used for Storage operations
}
func NewAccountWarmedEntry(address Address) *JournalEntry {
	return &JournalEntry{
		Type:    AccountWarmedType,
		Address: address,
	}
}

// NewAccountDestroyedEntry creates a new journal entry for destroying an account
func NewAccountDestroyedEntry(address, target Address, wasDestroyed bool, hadBalance U256) *JournalEntry {
	return &JournalEntry{
		Type:         AccountDestroyedType,
		Address:      address,
		Target:       target,
		WasDestroyed: wasDestroyed,
		HadBalance:   new(big.Int).Set(hadBalance),//to avoid mutating the original value (had balance not written directly) 
	}
}

// NewAccountTouchedEntry creates a new journal entry for touching an account
func NewAccountTouchedEntry(address Address) *JournalEntry {
	return &JournalEntry{
		Type:    AccountTouchedType,
		Address: address,
	}
}

// NewBalanceTransferEntry creates a new journal entry for balance transfer
func NewBalanceTransferEntry(from, to Address, balance U256) *JournalEntry {
	return &JournalEntry{
		Type:    BalanceTransferType,
		From:    from,
		To:      to,
		Balance: new(big.Int).Set(balance),
	}
}

// NewNonceChangeEntry creates a new journal entry for nonce change
func NewNonceChangeEntry(address Address) *JournalEntry {
	return &JournalEntry{
		Type:    NonceChangeType,
		Address: address,
	}
}

// NewAccountCreatedEntry creates a new journal entry for account creation
func NewAccountCreatedEntry(address Address) *JournalEntry {
	return &JournalEntry{
		Type:    AccountCreatedType,
		Address: address,
	}
}

// NewStorageChangedEntry creates a new journal entry for storage change
func NewStorageChangedEntry(address Address, key, hadValue U256) *JournalEntry {
	return &JournalEntry{
		Type:     StorageChangedType,
		Address:  address,
		Key:      new(big.Int).Set(key),
		HadValue: new(big.Int).Set(hadValue),
	}
}

// NewStorageWarmedEntry creates a new journal entry for storage warming
func NewStorageWarmedEntry(address Address, key U256) *JournalEntry {
	return &JournalEntry{
		Type:    StorageWarmedType,
		Address: address,
		Key:     new(big.Int).Set(key),
	}
}

// NewTransientStorageChangeEntry creates a new journal entry for transient storage change
func NewTransientStorageChangeEntry(address Address, key, hadValue U256) *JournalEntry {
	return &JournalEntry{
		Type:     TransientStorageChangeType,
		Address:  address,
		Key:      new(big.Int).Set(key),
		HadValue: new(big.Int).Set(hadValue),
	}
}

// NewCodeChangeEntry creates a new journal entry for code change
func NewCodeChangeEntry(address Address) *JournalEntry {
	return &JournalEntry{
		Type:    CodeChangeType,
		Address: address,
	}
}

// MarshalJSON implements the json.Marshaler interface
func (j JournalEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type         JournalEntryType `json:"type"`
		Address      Address          `json:"address"`
		Target       Address          `json:"target,omitempty"`
		WasDestroyed bool            `json:"was_destroyed,omitempty"`
		HadBalance   U256         `json:"had_balance,omitempty"`
		Balance      U256         `json:"balance,omitempty"`
		From         Address          `json:"from,omitempty"`
		To           Address          `json:"to,omitempty"`
		Key          U256         `json:"key,omitempty"`
		HadValue     U256         `json:"had_value,omitempty"`
	}{
		Type:         j.Type,
		Address:      j.Address,
		Target:       j.Target,
		WasDestroyed: j.WasDestroyed,
		HadBalance:   j.HadBalance,
		Balance:      j.Balance,
		From:         j.From,
		To:           j.To,
		Key:          j.Key,
		HadValue:     j.HadValue,
	})
}
type Account struct {
	Info AccountInfo
	Storage EvmStorage
	Status AccountStatus
}
type AccountInfo struct {
	Balance U256
	Nonce uint64
	CodeHash B256
	Code *Bytecode
}
type BytecodeKind int
const (
    LegacyRawKind BytecodeKind = iota
    LegacyAnalyzedKind
    EofKind
)

type AccountStatus uint8

// Define the constants using iota for bit shifting.
const (
	Loaded              AccountStatus = 0b00000000 // Account is loaded but not interacted with
	Created             AccountStatus = 0b00000001 // Account is newly created, no DB fetch required
	SelfDestructed      AccountStatus = 0b00000010 // Account is marked for self-destruction
	Touched             AccountStatus = 0b00000100 // Account is touched and will be saved to DB
	LoadedAsNotExisting AccountStatus = 0b0001000 // Pre-spurious state; account loaded but marked as non-existing
	Cold                AccountStatus = 0b0010000 // Account is marked as cold (not frequently accessed)
)
//BytecodeKind serves as discriminator for Bytecode to identify which variant of enum is being used
//1 , 2 or 3 in this case
type Bytecode struct{
	Kind BytecodeKind
	LegacyRaw      []byte                   // For LegacyRaw variant
    LegacyAnalyzed *LegacyAnalyzedBytecode  // For LegacyAnalyzed variant
    Eof            *Eof 				   // For Eof variant
}
//one field missing?
type LegacyAnalyzedBytecode struct {
	Bytecode    []byte
	OriginalLen uint64
	JumpTable   JumpTable
}
// JumpTable equivalent in Go, using a sync.Once pointer to simulate Arc and BitVec<u8>.
type JumpTable struct {
	BitVector *Bitvector       // Simulating BitVec<u8> as []byte
	once      sync.Once    // Lazy initialization if needed
}
//Doubt in JumpTable implementation

type Bitvector struct {
	bits []uint8
	size int // Total number of bits represented
}
//Doubt in bitvector implementation
//Code actually to be implemnted as ByteCode

//Or implementing as bool
/*
func NewEvmBuilder() *EvmBuilder {
	handlerCfg := NewHandlerCfg(SpecId(LATEST))
	// Enable optimism if the feature is available
	if os.Getenv("OPTIMISM_ENABLED") == "true" {
		handlerCfg.isOptimism = true
	} else {
		handlerCfg.isOptimism = false
	}
	return &EvmBuilder{
		context: Context{},
		handler: EvmBuilderHandler(handlerCfg),
	}
}
*/
/*
Enabling run time flexibility
# Run with optimism disabled (default)
go run main.go

# Run with optimism enabled
OPTIMISM_ENABLED=true go run main.go
*/
/*
func NewEvmBuilder() *EvmBuilder {
	handlerCfg := NewHandlerCfg(SpecId(LATEST))
	// Enable optimism if the feature is available
	if os.Getenv("OPTIMISM_ENABLED") == "true" {
		handlerCfg.isOptimism = true
	} else {
		handlerCfg.isOptimism = false
	}
	return &EvmBuilder{
		context: Context{},
		handler: EvmBuilderHandler(handlerCfg),
	}
}
*/



// SpecId is the enumeration of various Ethereum hard fork specifications.


// PhantomData is a Go struct that simulates the behavior of Rust's PhantomData.
type PhantomData[T any] struct{}
