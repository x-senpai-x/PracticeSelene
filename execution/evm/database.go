package evm

// DatabaseError represents possible database errors
type DatabaseError struct {
	msg string
	err error
}

func (e *DatabaseError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.msg
}

// Database interface defines required methods for blockchain database interaction
type Database interface {
	// Basic retrieves basic account information
	Basic(address Address) (*AccountInfo, error)

	// CodeByHash retrieves bytecode using its hash
	CodeByHash(codeHash B256) (Bytecode, error)

	// Storage retrieves storage value at given address and index
	Storage(address Address, index U256) (*U256, error)

	// BlockHash retrieves the hash of a block by its number
	BlockHash(number uint64) (B256, error)
}
//SAMPLE IMPLEMENTATION
/*
type MemoryDatabase struct {
	accounts map[Address]*AccountInfo
	storage  map[Address]map[string]*U256    // string is hex representation of Uint index
	code     map[B256]Bytecode
	blocks   map[uint64]B256
	mu       sync.RWMutex
}

func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		accounts: make(map[Address]*AccountInfo),
		storage:  make(map[Address]map[string]*U256),
		code:     make(map[B256]Bytecode),
		blocks:   make(map[uint64]B256),
	}
}

// Implement Database interface for MemoryDatabase
func (db *MemoryDatabase) Basic(address Address) (*AccountInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if acc, ok := db.accounts[address]; ok {
		return acc, nil
	}
	return nil, nil
}

func (db *MemoryDatabase) CodeByHash(codeHash B256) (Bytecode, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if code, ok := db.code[codeHash]; ok {
		return code, nil
	}
	return Bytecode{}, &DatabaseError{msg: "code not found"}
}

func (db *MemoryDatabase) Storage(address Address, index *U256) (*U256, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if storage, ok := db.storage[address]; ok {
		if value, ok := storage[index.data.Text(16)]; ok {
			return value, nil
		}
	}
	return *U256, nil // Return zero value if not found
}

func (db *MemoryDatabase) BlockHash(number uint64) (B256, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if hash, ok := db.blocks[number]; ok {
		return hash, nil
	}
	return B256{}, &DatabaseError{msg: "block not found"}
}
*/