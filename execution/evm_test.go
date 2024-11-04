package execution

import (
	"fmt"
	"log"
	"math/big"
	"sync"
	"testing"
	"time"
	"encoding/hex"
	"github.com/BlocSoc-iitr/selene/common"
	Gevm "github.com/BlocSoc-iitr/selene/execution/evm"
	Common "github.com/ethereum/go-ethereum/common" //geth common imported as Common
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)
func TestNewEvm(t *testing.T) {
	// Arrange
	executionClient := CreateNewExecutionClientWith() // Helper function to initialize ExecutionClient
	chainID := uint64(1)                              // Example chain ID
	tag := BlockTag{Latest: true}                   // Example BlockTag

	// Act
	evm := NewEvm(executionClient, chainID, tag)

	// Assert
	assert.NotNil(t, evm)
	assert.Equal(t, executionClient, evm.execution, "Expected execution client to be correctly set in Evm struct")
	assert.Equal(t, chainID, evm.chainID, "Expected chain ID to be correctly set in Evm struct")
	assert.Equal(t, tag, evm.tag, "Expected BlockTag to be correctly set in Evm struct")
}
func TestEvmExecutionClient(t *testing.T) {
	// Arrange
	executionClient := CreateNewExecutionClientWith() // Setup ExecutionClient with mock state and blocks
	evm := NewEvm(executionClient, 1, BlockTag{Number: 1})

	// Act
	state := executionClient.state // Hypothetical method to retrieve state in ExecutionClient

	// Assert
	assert.NotNil(t, state, "State should be retrievable from ExecutionClient")
	assert.NotNil(t, evm.execution, "Execution client within Evm should be initialized")
}
func TestGetEnv(t *testing.T) {
	exec := CreateNewExecutionClientWith()
	block := common.Block{
		Number: 1,
		Timestamp: 5,
		Difficulty: uint256.Int{1},
		Miner: common.Address{0x11},
	}
	exec.state.PushBlock(&block)
	evm := NewEvm(exec, 1, common.BlockTag{})
	
	opts := CallOpts{
		From: &Address{0x11},
		To: &Address{0x12},
		Gas: big.NewInt(1),
		GasPrice: big.NewInt(2),
		Value: big.NewInt(3),
		Data: []byte{0x13},
	}
	env := evm.getEnv(&opts, common.BlockTag{Number: 1})
	assert.Equal(t, *opts.From, env.Tx.Caller, "Not Equal")
	assert.Equal(t, opts.Value, env.Tx.Value, "Not Equal")
	assert.Equal(t, opts.Gas.Uint64(), env.Tx.GasLimit, "Not Equal")
	assert.Equal(t, opts.GasPrice, env.Tx.GasPrice, "Not Equal")
	assert.Equal(t, big.NewInt(int64(block.Number)), env.Block.Number, "Not Equal")
	assert.Equal(t, block.Miner, env.Block.Coinbase, "Not Equal")
	assert.Equal(t, big.NewInt(int64(block.Timestamp)), env.Block.Timestamp, "Not Equal")
	assert.Equal(t, block.Difficulty.ToBig(), env.Block.Difficulty, "Not Equal")
	assert.Equal(t, evm.chainID, env.Cfg.ChainID, "Not Equal")
}
func TestNewStateWithMultipleBlocks(t *testing.T) {
	// Create channels for blocks and finalized blocks
	blockChan := make(chan *common.Block)
	finalizedBlockChan := make(chan *common.Block)
	// Set up history length and initialize State
	historyLength := uint64(100)
	state := NewState(historyLength, blockChan, finalizedBlockChan)

	// Create mock addresses
	address1 := Common.Address{0x01}
	address2 := Common.Address{0x02}

	// Create mock transactions
	txHash1 := [32]byte{0x01}
	txHash2 := [32]byte{0x02}
	value1 := hexutil.Big(*uint256.NewInt(100).ToBig())
	gasPrice1 := hexutil.Big(*uint256.NewInt(1).ToBig())
	value2 := hexutil.Big(*uint256.NewInt(200).ToBig())
	gasPrice2 := hexutil.Big(*uint256.NewInt(2).ToBig())
	transaction1 := common.Transaction{
		Hash:             txHash1,
		Nonce:            1,
		From:             address1,
		To:               &address2,
		Value:            value1,
		GasPrice:         gasPrice1,
		TransactionIndex: 0,
	}
	transaction2 := common.Transaction{
		Hash:             txHash2,
		Nonce:            2,
		From:             address2,
		To:               &address1,
		Value:            value2,
		GasPrice:         gasPrice2,
		TransactionIndex: 1,
	}

	// Create a mock block with transactions
	block1 := &common.Block{
		Number:    1,
		Hash:      [32]byte{0x1},
		GasLimit:  1000000,
		GasUsed:   800000,
		Timestamp: 1234567890,
		Transactions: common.Transactions{
			Hashes: [][32]byte{txHash1, txHash2},
			Full:   []common.Transaction{transaction1, transaction2},
		},
	}

	// Create a second mock block with different transactions
	txHash3 := [32]byte{0x03}
	txHash4 := [32]byte{0x04}
	value3 := hexutil.Big(*uint256.NewInt(300).ToBig())
	gasPrice3 := hexutil.Big(*uint256.NewInt(3).ToBig())
	value4 := hexutil.Big(*uint256.NewInt(400).ToBig())
	gasPrice4 := hexutil.Big(*uint256.NewInt(4).ToBig())
	transaction3 := common.Transaction{
		Hash:             txHash3,
		Nonce:            3,
		From:             address1,
		To:               &address2,
		Value:            value3,
		GasPrice:         gasPrice3,
		TransactionIndex: 0,
	}
	transaction4 := common.Transaction{
		Hash:             txHash4,
		Nonce:            4,
		From:             address2,
		To:               &address1,
		Value:            value4,
		GasPrice:         gasPrice4,
		TransactionIndex: 1,
	}

	block2 := &common.Block{
		Number:    2,
		Hash:      [32]byte{0x2},
		GasLimit:  1000000,
		GasUsed:   850000,
		Timestamp: 1234567891,
		Transactions: common.Transactions{
			Hashes: [][32]byte{txHash3, txHash4},
			Full:   []common.Transaction{transaction3, transaction4},
		},
	}

	// Send blocks to blockChan and verify their addition
	go func() {
		blockChan <- block1
		blockChan <- block2
	}()
	time.Sleep(10 * time.Millisecond) // Allow time for processing

	state.mu.RLock()
	_, exists1 := state.blocks[block1.Number]
	_, exists2 := state.blocks[block2.Number]
	state.mu.RUnlock()
	assert.True(t, exists1, "block1 should be added to state.blocks")
	assert.True(t, exists2, "block2 should be added to state.blocks")

	// Verify transaction location mappings for each transaction in both blocks
	state.mu.RLock()
	txLoc1, txExists1 := state.txs[txHash1]
	txLoc2, txExists2 := state.txs[txHash2]
	txLoc3, txExists3 := state.txs[txHash3]
	txLoc4, txExists4 := state.txs[txHash4]
	state.mu.RUnlock()

	assert.True(t, txExists1, "transaction 1 should exist in state.txs")
	assert.True(t, txExists2, "transaction 2 should exist in state.txs")
	assert.True(t, txExists3, "transaction 3 should exist in state.txs")
	assert.True(t, txExists4, "transaction 4 should exist in state.txs")

	assert.Equal(t, TransactionLocation{Block: block1.Number, Index: 0}, txLoc1, "transaction 1 location should match")
	assert.Equal(t, TransactionLocation{Block: block1.Number, Index: 1}, txLoc2, "transaction 2 location should match")
	assert.Equal(t, TransactionLocation{Block: block2.Number, Index: 0}, txLoc3, "transaction 3 location should match")
	assert.Equal(t, TransactionLocation{Block: block2.Number, Index: 1}, txLoc4, "transaction 4 location should match")

	// Create a finalized block and verify it's updated in the state
	finalizedBlock := &common.Block{
		Number:    2,
		Hash:      [32]byte{0x2},
		GasLimit:  1000000,
		GasUsed:   850000,
		Timestamp: 1234567891,
	}
	go func() { finalizedBlockChan <- finalizedBlock }()
	time.Sleep(10 * time.Millisecond) // Allow time for processing

	state.mu.RLock()
	assert.Equal(t, finalizedBlock, state.finalizedBlock, "finalized block should be updated")
	state.mu.RUnlock()
}
func TestNewProofDB(t *testing.T) {
	// Setup
	tag := BlockTag{Number: 1}
	executionClient := CreateNewExecutionClientWith()
	proofDB, err := NewProofDB(tag, executionClient)
	assert.NoError(t, err, "Expected no error when creating NewProofDB")
	assert.NotNil(t, proofDB, "Expected non-nil ProofDB instance")
	assert.NotNil(t, proofDB.State, "Expected non-nil EvmState in ProofDB")
	assert.Equal(t, tag, proofDB.State.Block, "Expected BlockTag to be set correctly in EvmState")

	// Verify EvmState fields initialization
	assert.NotNil(t, proofDB.State.Basic, "Expected Basic map to be initialized in EvmState")
	assert.NotNil(t, proofDB.State.BlockHash, "Expected BlockHash map to be initialized in EvmState")
	assert.NotNil(t, proofDB.State.Storage, "Expected Storage map to be initialized in EvmState")
}
func TestNewEvmState(t *testing.T) {
	executionClient := CreateNewExecutionClientWith()
	tag := BlockTag{Number: 1}
	evmState := NewEvmState(executionClient, tag)
	assert.NotNil(t, evmState, "Expected non-nil EvmState instance")
	assert.Equal(t, tag, evmState.Block, "Expected BlockTag to be set correctly in EvmState")

	// Verify EvmState fields initialization
	assert.NotNil(t, evmState.Basic, "Expected Basic map to be initialized in EvmState")
	assert.NotNil(t, evmState.BlockHash, "Expected BlockHash map to be initialized in EvmState")
	assert.NotNil(t, evmState.Storage, "Expected Storage map to be initialized in EvmState")
	assert.Equal(t, executionClient, evmState.Execution, "Expected ExecutionClient to be set correctly in EvmState")
}
func CreateNewProofDB() *ProofDB {
	tag := BlockTag{Number: 1}
	executionClient := CreateNewExecutionClientWith() //Creates executionClient
	proofDB, err := NewProofDB(tag, executionClient)
	if err != nil {
		fmt.Println("Error in creating NewProofDB")
	}
	return proofDB
}
func CreateStateWithMultipleBlocks() (*State, *common.Block, *common.Block) {
	// Create channels for blocks and finalized blocks
	blockChan := make(chan *common.Block)
	finalizedBlockChan := make(chan *common.Block)

	// Set up history length and initialize State
	historyLength := uint64(100)
	state := NewState(historyLength, blockChan, finalizedBlockChan)

	// Create mock addresses
	address1 := Common.Address{0x01}
	address2 := Common.Address{0x02}

	// Create mock transactions
	txHash1 := [32]byte{0x01}
	txHash2 := [32]byte{0x02}
	value1 := hexutil.Big(*uint256.NewInt(100).ToBig())
	gasPrice1 := hexutil.Big(*uint256.NewInt(1).ToBig())
	value2 := hexutil.Big(*uint256.NewInt(200).ToBig())
	gasPrice2 := hexutil.Big(*uint256.NewInt(2).ToBig())
	transaction1 := common.Transaction{
		Hash:             txHash1,
		Nonce:            1,
		From:             address1,
		To:               &address2,
		Value:            value1,
		GasPrice:         gasPrice1,
		TransactionIndex: 0,
	}
	transaction2 := common.Transaction{
		Hash:             txHash2,
		Nonce:            2,
		From:             address2,
		To:               &address1,
		Value:            value2,
		GasPrice:         gasPrice2,
		TransactionIndex: 1,
	}
 
	// Create a mock block with transactions
	block1 := &common.Block{
		Number:    1,
		Hash:      [32]byte{0x1},
		GasLimit:  1000000,
		GasUsed:   800000,
		Timestamp: 1234567890,
		Transactions: common.Transactions{
			Hashes: [][32]byte{txHash1, txHash2},
			Full:   []common.Transaction{transaction1, transaction2},
		},
	}

	// Create a second mock block with different transactions
	txHash3 := [32]byte{0x03}
	txHash4 := [32]byte{0x04}
	value3 := hexutil.Big(*uint256.NewInt(300).ToBig())
	gasPrice3 := hexutil.Big(*uint256.NewInt(3).ToBig())
	value4 := hexutil.Big(*uint256.NewInt(400).ToBig())
	gasPrice4 := hexutil.Big(*uint256.NewInt(4).ToBig())
	transaction3 := common.Transaction{
		Hash:             txHash3,
		Nonce:            3,
		From:             address1,
		To:               &address2,
		Value:            value3,
		GasPrice:         gasPrice3,
		TransactionIndex: 0,
	}
	transaction4 := common.Transaction{
		Hash:             txHash4,
		Nonce:            4,
		From:             address2,
		To:               &address1,
		Value:            value4,
		GasPrice:         gasPrice4,
		TransactionIndex: 1,
	}

	block2 := &common.Block{
		Number:    2,
		Hash:      [32]byte{0x2},
		GasLimit:  1000000,
		GasUsed:   850000,
		Timestamp: 1234567891,
		Transactions: common.Transactions{
			Hashes: [][32]byte{txHash3, txHash4},
			Full:   []common.Transaction{transaction3, transaction4},
		},
	}
	// Send blocks to blockChan
	go func() {
		blockChan <- block1
		blockChan <- block2
	}()
	time.Sleep(100 * time.Millisecond) // Allow time for processing
	finalizedBlock := &common.Block{
		Number:    2,
		Hash:      [32]byte{0x2},
		GasLimit:  1000000,
		GasUsed:   850000,
		Timestamp: 1234567891,
	}
	go func() { finalizedBlockChan <- finalizedBlock }()
	time.Sleep(100 * time.Millisecond)

	return state, block1, block2
}
func CreateNewExecutionClientWith() *ExecutionClient {
	//rpc := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
	rpc := "https://eth-mainnet.g.alchemy.com/v2/6KA6UTwKL2hmb9AOorypuZX805DIl9KB/getNFTs?owner=0xF039fbEfBA314ecF4Bf0C32bBe85f620C8C460D2"
	state, _, _ := CreateStateWithMultipleBlocks()
	var executionClient *ExecutionClient
	executionClient, _ = executionClient.New(rpc, state)

	return executionClient
}
func CreateNewEvmState() *EvmState {
	tag := BlockTag{Number: 1}
	executionClient := CreateNewExecutionClientWith()
	evmState := NewEvmState(executionClient, tag)
	return evmState
}
func TestUpdateStateBlockHash(t *testing.T) {
	// Setup
	evmState := CreateNewEvmState()
	evmState.mu = sync.RWMutex{}
	blockNumber := uint64(2)
	blockTag := BlockTag{Number: blockNumber}
	block, _:= evmState.Execution.GetBlock(blockTag, false)
	t.Logf("Retrieved Block: %+v", block)

	// Set the Access to update BlockHash
	evmState.Access = &StateAccess{
		BlockHash: &blockNumber,
	}
	// Test UpdateState
	err := evmState.UpdateState()
	assert.NoError(t, err, "Expected no error on UpdateState for BlockHash access")
	hash, exists := evmState.BlockHash[blockNumber]
	assert.True(t, exists, "Expected block hash to be added to BlockHash map in EvmState")
	expectedHashArray := [32]byte{0x2}
	expectedHash := B256FromSlice(expectedHashArray[:])
	assert.Equal(t, expectedHash, hash, "Expected correct block hash in BlockHash map")
}

func TestUpdateStateInvalidAccessType(t *testing.T) {
	// Setup
	evmState := CreateNewEvmState()
	evmState.mu = sync.RWMutex{}
	// Set the Access to nil (no access type set)
	evmState.Access = nil
	// Test UpdateState
	err := evmState.UpdateState()
	assert.NoError(t, err, "Expected no error on UpdateState with nil Access")
}
//Tests remaining for TestUpdateStateBasic,TestUpdateStateStorage 

// Unable to check concurrency
func TestNeedsUpdate(t *testing.T) {
	// Isolate each test by creating a new evmState
	evmState := CreateNewEvmState()

	// 1. Test with Access as nil (no update needed)
	assert.False(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return false when Access is nil")

	// 2. Set Access to a valid StateAccess
	evmState.Access = &StateAccess{}
	assert.True(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return true when Access is set")

	// 3. Test with Access set to a nil pointer (still considered as update needed)
	evmState.Access = &StateAccess{Basic: nil}
	assert.True(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return true when Access is set with nil Basic")

	// 4. Set Access back to nil and test again
	evmState.Access = nil
	assert.False(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return false when Access is nil again")
	/*
	   // 5. Concurrent access test
	   var wg sync.WaitGroup
	   wg.Add(2)

	   go func() {
	       defer wg.Done()
	       evmState.Access = &StateAccess{Basic: new(Address)} // Simulate an update in a goroutine
	   }()

	   go func() {
	       defer wg.Done()
	       // Check NeedsUpdate concurrently
	       needsUpdate := evmState.NeedsUpdate()
	       assert.True(t, needsUpdate, "Expected NeedsUpdate to return true during concurrent access")
	   }()

	   wg.Wait()
	*/
	// 6. Reset Access again
	evmState.Access = nil
	assert.False(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return false when Access is reset to nil")

	// 7. Multiple updates
	evmState.Access = &StateAccess{Basic: new(Address)}
	assert.True(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return true with Basic set")

	evmState.Access = &StateAccess{Storage: &struct {
		Address Address
		Slot    U256
	}{Address: Address{0x45, 0x65}, Slot: U256(big.NewInt(2))}}
	assert.True(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return true with Storage set")

	evmState.Access = &StateAccess{BlockHash: new(uint64)}
	assert.True(t, evmState.NeedsUpdate(), "Expected NeedsUpdate to return true with BlockHash set")
}

func TestGetBasicExists(t *testing.T) {
	evmState := CreateNewEvmState()
	address := Address{0x12, 0x30}
	accountInfo := Gevm.AccountInfo{Balance: big.NewInt(100)}

	// Add the address to Basic map for testing
	evmState.Basic[address] = accountInfo

	// Test GetBasic with an existing address
	result, err := evmState.GetBasic(address)
	assert.NoError(t, err, "Expected no error when getting existing account")
	assert.Equal(t, accountInfo, result, "Expected returned account info to match")
}
func TestGetBasicMissing(t *testing.T) {
	evmState := CreateNewEvmState()
	address := Address{0x12, 0x30}

	// Test GetBasic with a missing address
	result, err := evmState.GetBasic(address)
	assert.Error(t, err, "Expected an error when getting a missing account")
	assert.Equal(t, Gevm.AccountInfo{}, result, "Expected returned account info to be zero value")
	assert.NotNil(t, evmState.Access, "Expected Access to be set when account is missing")
	assert.Equal(t, address, *evmState.Access.Basic, "Expected Access.Basic to match the requested address")
}
func TestGetStorageExists(t *testing.T) {
	evmState := CreateNewEvmState()
	address := Address{0x12, 0x30}
	log.Println("Address:", address)
	slot := U256(big.NewInt(1))
	value := U256(big.NewInt(200))
	// Add storage for the address
	if evmState.Storage[address] == nil {
		evmState.Storage[address] = make(map[U256]U256)
	}
	evmState.Storage[address][slot] = value

	// Test GetStorage with an existing slot
	result, err := evmState.GetStorage(address, slot)
	assert.NoError(t, err, "Expected no error when getting existing storage")
	assert.Equal(t, value, result, "Expected returned storage value to match")
}
func TestGetStorageMissing(t *testing.T) {
	evmState := CreateNewEvmState()
	address := Address{0x12, 0x30}
	slot := U256(big.NewInt(1))

	// Test GetStorage with a missing slot
	result, err := evmState.GetStorage(address, slot)
	assert.Error(t, err, "Expected an error when getting missing storage")
	assert.Equal(t, &big.Int{}, result, "Expected returned storage value to be zero value")
	assert.NotNil(t, evmState.Access, "Expected Access to be set when storage is missing")
	assert.Equal(t, address, evmState.Access.Storage.Address, "Expected Access.Storage.Address to match the requested address")
	assert.Equal(t, slot, evmState.Access.Storage.Slot, "Expected Access.Storage.Slot to match the requested slot")
}
func TestGetBlockHashExists(t *testing.T) {
	evmState := CreateNewEvmState()
	block := uint64(1)
	expectedHash := B256{1}

	// Add block hash to BlockHash map
	evmState.BlockHash[block] = expectedHash

	// Test GetBlockHash with an existing block
	result, err := evmState.GetBlockHash(block)
	assert.NoError(t, err, "Expected no error when getting existing block hash")
	assert.Equal(t, expectedHash, result, "Expected returned block hash to match")
}

func TestGetBlockHashMissing(t *testing.T) {
	evmState := CreateNewEvmState()
	block := uint64(1)

	// Test GetBlockHash with a missing block
	result, err := evmState.GetBlockHash(block)
	assert.Error(t, err, "Expected an error when getting missing block hash")
	assert.Equal(t, B256{}, result, "Expected returned block hash to be zero value")
	assert.NotNil(t, evmState.Access, "Expected Access to be set when block hash is missing")
	assert.Equal(t, block, *evmState.Access.BlockHash, "Expected Access.BlockHash to match the requested block")
}
func TestUpdateState(t *testing.T) {
    // Replace with a known good address for testing

	addressBytes := parseHexAddress("0x710bDa329b2a6224E4B44833DE30F38E7f81d564")
	address := common.Address(addressBytes)

    tests := []struct {
        name          string
        setupState    func(*testing.T) *EvmState
        expectedError string
        validateState func(*testing.T, *EvmState)
    }{
        {
            name: "Successful update from basic account",
            setupState: func(t *testing.T) *EvmState {
                executionClient := CreateNewExecutionClient()
                state := &EvmState{
                    Basic:     make(map[Address]Gevm.AccountInfo),
                    Storage:   make(map[Address]map[U256]U256),
                    Block:     BlockTag{Finalized: true},
                    Execution: executionClient,
                    mu:        sync.RWMutex{},
                }

                // Mock Access for a known address
                state.Access = &StateAccess{
                    Basic: &address,
                }

                return state
            },
            expectedError: "", // Expect no error
            validateState: func(t *testing.T, state *EvmState) {
                accountInfo, exists := state.Basic[address]
                assert.True(t, exists, "Basic state should have entry for address")
                assert.NotNil(t, accountInfo.Code, "Account code should not be nil")
            },
        },
        {
            name: "Failed update due to invalid account",
            setupState: func(t *testing.T) *EvmState {
                executionClient := CreateNewExecutionClient()
                state := &EvmState{
                    Basic:     make(map[Address]Gevm.AccountInfo),
                    Storage:   make(map[Address]map[U256]U256),
                    Block:     BlockTag{Finalized: true},
                    Execution: executionClient,
                    mu:        sync.RWMutex{},
                }

                // Using an address known not to exist
                invalidAddress := common.Address(parseHexAddress("0xb8901acB165ed027E32754E0FFe830802919727f"))
                state.Access = &StateAccess{
                    Basic: &invalidAddress,
                }

                return state
            },
            expectedError: "invalid account proof", // Adjust based on actual expected error
            validateState: func(t *testing.T, state *EvmState) {
                _, exists := state.Basic[address]
                assert.False(t, exists, "Basic state should not have entry for address after error")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            state := tt.setupState(t)

            // Execute
            err := state.UpdateState()

            // Verify error
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
            }

            // Validate state
            tt.validateState(t, state)
        })
    }
}


func TestPrefetchState(t *testing.T) {
	addressBytes := parseHexAddress("0x710bDa329b2a6224E4B44833DE30F38E7f81d564")
	addressFrom := common.Address(addressBytes)
	addressBytes2:=parseHexAddress("0xb8901acB165ed027E32754E0FFe830802919727f")
	addressTo:=common.Address(addressBytes2)
	tests := []struct {
		name          string
		setupState    func(*testing.T) *EvmState
		opts          *CallOpts
		expectedError error
		validateState func(*testing.T, *EvmState)
	}{
		{
			name: "Successful prefetch",
			setupState: func(t *testing.T) *EvmState {
				executionClient := CreateNewExecutionClient()

				state := &EvmState{
					Basic:     make(map[Address]Gevm.AccountInfo),
					Storage:   make(map[Address]map[U256]U256),
					Block:     BlockTag{Finalized: true},
					Execution: executionClient,
					mu:        sync.RWMutex{},
				}

				return state
			},
			opts: &CallOpts{
				From: (&addressFrom),

				To:   &addressTo,
				// Assuming CallOpts struct has string fields for addresses

			},
			expectedError: nil,
			validateState: func(t *testing.T, state *EvmState) {
				assert.NotNil(t, state.Basic, "Basic state should not be nil")
				assert.NotNil(t, state.Storage, "Storage state should not be nil")
				// Add more specific state validations
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			state := tt.setupState(t)

			// Execute
			err := state.PrefetchState(tt.opts)

			// Verify error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			// Validate state
			tt.validateState(t, state)
		})
	}
}
func parseHexAddress(hexStr string) [20]byte {
	var addr [20]byte
	bytes, _ := hex.DecodeString(hexStr[2:])
	copy(addr[:], bytes)
	return addr
}
