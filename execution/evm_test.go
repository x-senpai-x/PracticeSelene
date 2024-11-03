package execution

import (
	// "encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	//"github.com/stretchr/testify/mock"
	"fmt"
	"sync"
	"time"

	"github.com/BlocSoc-iitr/selene/common"
	Gevm "github.com/BlocSoc-iitr/selene/execution/evm"
	Common "github.com/ethereum/go-ethereum/common" //geth common imported as Common
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)
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
func CreateTestState() *State {
	blockChan := make(chan *common.Block)
	finalizedBlockChan := make(chan *common.Block)

	state := NewState(5, blockChan, finalizedBlockChan)

	// Create test blocks
	block1 := &common.Block{
		Number: 1,
		Hash:   [32]byte{0x1},
		Transactions: common.Transactions{
			Hashes: [][32]byte{{0x11}, {0x12}},
		},
	}

	block2 := &common.Block{
		Number: 2,
		Hash:   [32]byte{0x2},
		Transactions: common.Transactions{
			Hashes: [][32]byte{{0x21}, {0x22}},
			Full: []common.Transaction{
				{
					Hash:         Common.Hash([32]byte{0x21}),
					GasPrice:     hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
					Gas:          hexutil.Uint64(5),
					MaxFeePerGas: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
				},
				{Hash: Common.Hash([32]byte{0x22})},
			},
		},
	}

	// Push blocks through channel
	go func() {
		blockChan <- block1
		blockChan <- block2
		close(blockChan)
	}()

	// Wait for blocks to be processed
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for len(state.blocks) < 2 {
			// wait for blocks to be processed
		}
	}()
	wg.Wait()

	return state
}
func TestUpdateStateBasic(t *testing.T) {
	// Setup
	evmState := CreateNewEvmState()
	evmState.mu = sync.RWMutex{}
	var address common.Address = [20]byte(Common.Hex2Bytes("95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"))
	evmState.Access = &StateAccess{
		Basic: &address,
	}
	evmState.Block = common.BlockTag{Finalized: true}
	err := evmState.UpdateState()
	assert.NoError(t, err, "Expected no error on UpdateState for Basic access")
	accountInfo, exists := evmState.Basic[address]
	assert.True(t, exists, "Expected account to be added to Basic map in EvmState")
	assert.Equal(t, uint64(1), accountInfo.Nonce, "Expected correct Nonce in account info")
	assert.Equal(t, ConvertU256(big.NewInt(1000)), accountInfo.Balance, "Expected correct Balance in account info")
}
func TestPrefetchState(t *testing.T) {
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
				From: (*Address)(Common.Hex2Bytes("710bDa329b2a6224E4B44833DE30F38E7f81d564")),
				To:   (*Address)(Common.Hex2Bytes("b8901acB165ed027E32754E0FFe830802919727f")),
			},
			expectedError: nil,
			validateState: func(t *testing.T, state *EvmState) {
				assert.NotNil(t, state.Basic, "Basic state should not be nil")
				assert.NotNil(t, state.Storage, "Storage state should not be nil")
				// Add more specific state validations
			},
		},
		// {
		// 	name: "RPC error",
		// 	setupState: func(t *testing.T) *EvmState {
		// 		executionClient := CreateNewExecutionClient()

		// 		state := &EvmState{
		// 			Basic:     make(map[Address]Gevm.AccountInfo),
		// 			Storage:   make(map[Address]map[U256]U256),
		// 			Block:     BlockTag{Finalized: true},
		// 			Execution: executionClient,
		// 			mu:        sync.RWMutex{},
		// 		}

		// 		return state
		// 	},
		// 	opts: &CallOpts{
		// 		From: (*Address)(Common.Hex2Bytes("710bDa329b2a6224E4B44833DE30F38E7f81d564")),
		// 		To:   (*Address)(Common.Hex2Bytes("b8901acB165ed027E32754E0FFe830802919727f")),
		// 	},
		// 	expectedError: fmt.Errorf("create access list: RPC error"),
		// 	validateState: func(t *testing.T, state *EvmState) {
		// 		assert.Empty(t, state.Basic, "Basic state should be empty")
		// 		assert.Empty(t, state.Storage, "Storage state should be empty")
		// 	},
		// },
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

func TestGetEnv(t *testing.T) {
	exec := CreateNewExecutionClient()
	block := common.Block{
		Number: 1,
		Timestamp: 5,
		Difficulty: uint256.Int{1},
		Miner: common.Address{0x11},
	}
	exec.state.PushBlock(&block)
	evm := NewEvm(exec, 1, common.BlockTag{})
	
	opts := CallOpts{
		From: &common.Address{0x11},
		To: &common.Address{0x12},
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

/*

// Mock RPC client for testing
type MockRPCClient struct {
    accessList []AccessListItem
    block      *common.Block
    account    Account
    err        error
}
func (m *MockRPCClient) New(rpc *string) (ExecutionRpc, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m, nil
}

func (m *MockRPCClient) GetProof(address *common.Address, slots *[]common.Hash, block uint64) (EIP1186ProofResponse, error) {
    if m.err != nil {
        return EIP1186ProofResponse{}, m.err
    }
    return m.proof, nil
}

func (m *MockRPCClient) GetCode(address *common.Address, block uint64) ([]byte, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.code, nil
}

func (m *MockRPCClient) SendRawTransaction(bytes *[]byte) (common.Hash, error) {
    if m.err != nil {
        return common.Hash{}, m.err
    }
    return m.txHash, nil
}

func (m *MockRPCClient) GetTransactionReceipt(txHash *common.Hash) (types.Receipt, error) {
    if m.err != nil {
        return types.Receipt{}, m.err
    }
    return m.receipt, nil
}

func (m *MockRPCClient) GetTransaction(txHash *common.Hash) (seleneCommon.Transaction, error) {
    if m.err != nil {
        return seleneCommon.Transaction{}, m.err
    }
    return m.transaction, nil
}

func (m *MockRPCClient) GetLogs(filter *ethereum.FilterQuery) ([]types.Log, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.logs, nil
}

func (m *MockRPCClient) GetFilterChanges(filterID *uint256.Int) ([]types.Log, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.filterChanges, nil
}

func (m *MockRPCClient) UninstallFilter(filterID *uint256.Int) (bool, error) {
    if m.err != nil {
        return false, m.err
    }
    return true, nil
}

func (m *MockRPCClient) GetNewFilter(filter *ethereum.FilterQuery) (uint256.Int, error) {
    if m.err != nil {
        return uint256.Int{}, m.err
    }
    return m.filterID, nil
}

func (m *MockRPCClient) GetNewBlockFilter() (uint256.Int, error) {
    if m.err != nil {
        return uint256.Int{}, m.err
    }
    return m.blockFilterID, nil
}

func (m *MockRPCClient) GetNewPendingTransactionFilter() (uint256.Int, error) {
    if m.err != nil {
        return uint256.Int{}, m.err
    }
    return m.pendingFilterID, nil
}

func (m *MockRPCClient) ChainId() (uint64, error) {
    if m.err != nil {
        return 0, m.err
    }
    return m.chainID, nil
}

func (m *MockRPCClient) GetFeeHistory(blockCount uint64, lastBlock uint64, rewardPercentiles *[]float64) (FeeHistory, error) {
    if m.err != nil {
        return FeeHistory{}, m.err
    }
    return m.feeHistory, nil
}

func (m *MockRPCClient) CreateAccessList(opts CallOpts, block BlockTag) ([]AccessListItem, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.accessList, nil
}

func (m *MockRPCClient) GetBlock(tag BlockTag, full bool) (*common.Block, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.block, nil
}

func (m *MockRPCClient) GetAccount(address *Address, storageKeys *[]Common.Hash, block *BlockTag) (Account, error) {
    if m.err != nil {
        return Account{}, m.err
    }
    return m.account, nil
}


func CreateTestExecutionClient() (*ExecutionClient, *MockRPCClient) {
    state := CreateTestState()
	mockRPC:=HttpRpc{

		url: "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt",
		provider: nil,
	}
    /*
    // Create mock RPC client with test data
    mockRPC := &MockRPCClient{
        accessList: []AccessListItem{
            {
                Address: Address{Addr: [20]byte{1}},
                StorageKeys: []Common.Hash{},
            },
        },
        block: &common.Block{
            Miner: Address{Addr: [20]byte{2}},
            // Add other block fields as needed
        },
        account: Account{
            Balance:   big.NewInt(100),
            Nonce:    1,
            Code:     []byte{1, 2, 3},
            CodeHash: [32]byte{1},
            Slots:    []Slot{{Key: Common.Hash{1}, Value: big.NewInt(200)}},
        },
    }
*/
/*
    // Create execution client with mock RPC
    executionClient := &ExecutionClient{
        Rpc:   mockRPC,
		state: state,
    }

    return executionClient, mockRPC
}
*/
/*



/*
func TestUpdateStateStorage(t *testing.T) {
	// Setup
	mockClient := &MockExecutionClient{}
	evmState := CreateNewEvmState(mockClient, BlockTag{Number: 1234})
	evmState.mu = sync.RWMutex{}
	address := Address{0x4, 0x5, 0x6}
	slot := U256FromBig(big.NewInt(1))

	// Set the Access to update Storage
	evmState.Access = &StateAccess{
		Storage: &struct {
			Address Address
			Slot    U256
		}{
			Address: address,
			Slot:    slot,
		},
	}

	// Test UpdateState
	err := evmState.UpdateState()
	assert.NoError(t, err, "Expected no error on UpdateState for Storage access")
	storage, exists := evmState.Storage[address]
	assert.True(t, exists, "Expected storage to be added to Storage map in EvmState")
	value, found := storage[slot]
	assert.True(t, found, "Expected slot to be added to address Storage map")
	assert.Equal(t, U256FromBig(big.NewInt(42)), value, "Expected correct value in slot")
}



*/
/*
type ExecutionInterface interface {
	GetAccount(address *Address, slots []Common.Hash, tag BlockTag) (Account, error)
	GetBlock(tag BlockTag, fullTx bool) (common.Block, error)
}
*/
// MockExecutionClient mocks the ExecutionClient
/*type MockExecutionClient struct {
	mock.Mock
	*ExecutionClient // Embed the actual ExecutionClient
}*/
/*
func (m *MockExecutionClient) GetAccount(address *Address, slots []Common.Hash, tag BlockTag) (Account, error) {
	args := m.Called(address, slots, tag)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockExecutionClient) GetBlock(tag BlockTag, fullTx bool) (common.Block, error) {
	args := m.Called(tag, fullTx)
	return args.Get(0).(common.Block), args.Error(1)
}*/
/*
func NewMockExecutionClient() *MockExecutionClient {
	return &MockExecutionClient{
		ExecutionClient: &ExecutionClient{}, // Initialize with empty ExecutionClient
	}
}
func TestNewProofDB(t *testing.T) {
	mockExec := NewMockExecutionClient()
	tag := BlockTag{Number: 1234}

	db, err := NewProofDB(tag, mockExec.ExecutionClient)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, db.State)
	assert.Equal(t, tag, db.State.Block)
}

func TestEvmState_GetBasic(t *testing.T) {
	mockExec := NewMockExecutionClient()
	state := NewEvmState(mockExec.ExecutionClient, BlockTag{Number: 1234})

	addr := Address{Addr: [20]byte{1, 2, 3, 4}}

	// Test different bytecode variants
	testCases := []struct {
		name     string
		bytecode Gevm.Bytecode
	}{
		{
			name: "LegacyRaw bytecode",
			bytecode: Gevm.Bytecode{
				Kind:      Gevm.LegacyRawKind,
				LegacyRaw: []byte{1, 2, 3},
			},
		},
		{
			name: "LegacyAnalyzed bytecode",
			bytecode: Gevm.Bytecode{
				Kind: Gevm.LegacyAnalyzedKind,
				LegacyAnalyzed: &Gevm.LegacyAnalyzedBytecode{
					Bytecode:    []byte{1, 2, 3},
					OriginalLen: 3,
					JumpTable:   Gevm.JumpTable{},
				},
			},
		},
		{
			name: "Empty bytecode",
			bytecode: Gevm.Bytecode{
				Kind:      Gevm.LegacyRawKind,
				LegacyRaw: []byte{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedAccount := Gevm.AccountInfo{
				Balance:  big.NewInt(100),
				Nonce:    1,
				CodeHash: B256{1},
				Code:     &tc.bytecode,
			}

			state.mu.Lock()
			state.Basic[addr] = expectedAccount
			state.mu.Unlock()

			account, err := state.GetBasic(addr)
			assert.NoError(t, err)
			assert.Equal(t, expectedAccount, account)
            fmt.Printf("Balance: %v, Code: %v, CodeHash: %x, Nonce: %d, Code.Kind: %v, Code.LegacyAnalyzed: %v, Code.LegacyAnalyzed.OriginalLen: %v\n",
                expectedAccount.Balance,
                expectedAccount.Code,
                expectedAccount.CodeHash,
                expectedAccount.Nonce,
                expectedAccount.Code.Kind,
                expectedAccount.Code.LegacyAnalyzed,
                expectedAccount.Code.LegacyAnalyzed.OriginalLen)})
	}

	// Test cache miss
	missAddr := Address{Addr: [20]byte{5, 6, 7, 8}}
	_, err := state.GetBasic(missAddr)
	assert.Error(t, err)
	assert.Equal(t, "state missing", err.Error())

	state.mu.RLock()
	assert.NotNil(t, state.Access)
	assert.Equal(t, &missAddr, state.Access.Basic)
	state.mu.RUnlock()
}
func TestEvmState_GetStorage(t *testing.T) {
	mockExec := NewMockExecutionClient()
	state := NewEvmState(mockExec.ExecutionClient, BlockTag{Number: 1234})

	addr := Address{Addr: [20]byte{1, 2, 3, 4}}
	slot := big.NewInt(1)
	expectedValue := big.NewInt(100)

	// Test cache hit
	state.mu.Lock()
	storage := make(map[U256]U256)
	storage[slot] = expectedValue
	state.Storage[addr] = storage
	state.mu.Unlock()

	value, err := state.GetStorage(addr, slot)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)

	// Test cache miss
	missSlot := big.NewInt(2)
	value, err = state.GetStorage(addr, missSlot)
	assert.Error(t, err)
	assert.Equal(t, "state missing", err.Error())

	state.mu.RLock()
	assert.NotNil(t, state.Access)
	assert.NotNil(t, state.Access.Storage)
	assert.Equal(t, addr, state.Access.Storage.Address)
	assert.Equal(t, missSlot, state.Access.Storage.Slot)
	state.mu.RUnlock()
}

func TestEvmState_GetBlockHash(t *testing.T) {
	mockExec := NewMockExecutionClient()
	state := NewEvmState(mockExec.ExecutionClient, BlockTag{Number: 1234})

	blockNum := uint64(1233)
	expectedHash := B256{1, 2, 3}

	// Test cache hit
	state.mu.Lock()
	state.BlockHash[blockNum] = expectedHash
	state.mu.Unlock()

	hash, err := state.GetBlockHash(blockNum)
	assert.NoError(t, err)
	assert.Equal(t, expectedHash, hash)

	// Test cache miss
	missBlockNum := uint64(1232)
	hash, err = state.GetBlockHash(missBlockNum)
	assert.Error(t, err)
	assert.Equal(t, "state missing", err.Error())

	state.mu.RLock()
	assert.NotNil(t, state.Access)
	assert.Equal(t, &missBlockNum, state.Access.BlockHash)
	state.mu.RUnlock()
}

func TestEvmState_UpdateState(t *testing.T) {
	mockExec := NewMockExecutionClient()
	state := NewEvmState(mockExec.ExecutionClient, BlockTag{Number: 1234})

	// Test basic account update
	addr := Address{Addr: [20]byte{1, 2, 3, 4}}
	expectedAccount := Account{
		Balance:  big.NewInt(100),
		Nonce:    1,
		CodeHash: Common.Hash{1},
		Code:     []byte{1, 2, 3},
	}

	mockExec.On("GetAccount", &addr, []Common.Hash{}, BlockTag{Number: 1234}).Return(expectedAccount, nil)

	state.mu.Lock()
	state.Access = &StateAccess{Basic: &addr}
	state.mu.Unlock()

	err := state.UpdateState()
	assert.NoError(t, err)

	state.mu.RLock()
	account, exists := state.Basic[addr]
	state.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, expectedAccount.Balance, account.Balance)
	assert.Equal(t, expectedAccount.Nonce, account.Nonce)
}

func TestEvmState_PrefetchState(t *testing.T) {
	mockExec := NewMockExecutionClient()
	state := NewEvmState(mockExec.ExecutionClient, BlockTag{Number: 1234})

	fromAddr := Address{Addr: [20]byte{1}}
	toAddr := Address{Addr: [20]byte{2}}

	opts := &CallOpts{
		From: &fromAddr,
		To:   &toAddr,
	}

	accessList := []AccessListItem{
		{
			Address:     fromAddr,
			StorageKeys: []Common.Hash{{1}},
		},
		{
			Address:     toAddr,
			StorageKeys: []Common.Hash{{2}},
		},
	}

	mockBlock := common.Block{
		Number: 1234,
		Miner:  Address{Addr: [20]byte{3}},
	}

	mockExec.On("GetBlock", state.Block, false).Return(mockBlock, nil)
	mockExec.On("Rpc.CreateAccessList", *opts, state.Block).Return(accessList, nil)

	// Mock GetAccount calls for each address
	for _, item := range accessList {
		mockExec.On("GetAccount", &item.Address, item.StorageKeys, state.Block).Return(
			Account{
				Balance:  big.NewInt(100),
				Nonce:    1,
				CodeHash: Common.Hash{1},
				Code:     []byte{1, 2, 3},
				Slots:    make(map[Common.Hash]*big.Int),
			},
			nil,
		)
	}

	err := state.PrefetchState(opts)
	assert.NoError(t, err)

	// Verify that accounts were cached
	state.mu.RLock()
	defer state.mu.RUnlock()

	for _, item := range accessList {
		_, exists := state.Basic[item.Address]
		assert.True(t, exists)
	}
}

func TestIsPrecompile(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		want    bool
	}{
		{
			name:    "Zero address",
			address: Address{},
			want:    false,
		},
		{
			name: "Precompile address",
			address: Address{
				Addr: [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
			want: true,
		},
		{
			name: "Normal address",
			address: Address{
				Addr: [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrecompile(tt.address)
			assert.Equal(t, tt.want, got)
		})
	}
}*/
/*package execution

import (
	"math/big"
	"sync"
	"testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	Gevm "github.com/BlocSoc-iitr/selene/execution/evm"
	seleneCommon "github.com/BlocSoc-iitr/selene/common"


)

// MockExecutionClient is a mock of ExecutionClient
type MockExecutionClient struct {
	mock.Mock
}

func (m *MockExecutionClient) GetAccount(address *Address, slots []common.Hash, block BlockTag) (Account, error) {
	args := m.Called(address, slots, block)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockExecutionClient) GetBlock(tag BlockTag, full bool) (seleneCommon.Block, error) {
	args := m.Called(tag, full)
	return args.Get(0).(seleneCommon.Block), args.Error(1)
}

// Helper function to create test addresses
func createTestAddress(num byte) Address {
	addr := Address{}
	addr.Addr[19] = num
	return addr
}

func TestProofDB_Basic(t *testing.T) {
	tests := []struct {
		name          string
		address       Address
		mockSetup     func(*MockExecutionClient)
		expectedInfo  Gevm.AccountInfo
		expectedError bool
	}{
		{
			name:    "successful fetch of basic account info",
			address: createTestAddress(1),
			mockSetup: func(m *MockExecutionClient) {
				m.On("GetAccount",
					mock.MatchedBy(func(addr *Address) bool {
						return addr.Addr[19] == 1
					}),
					mock.AnythingOfType("[]common.Hash"),
					mock.AnythingOfType("BlockTag")).
					Return(Account{
						Balance:   big.NewInt(100),
						Nonce:     1,
						Code:      []byte{1, 2, 3},
						CodeHash:  common.Hash{1},
						Slots:     make(map[common.Hash]*big.Int),
					}, nil).Once()
			},
			expectedInfo: Gevm.AccountInfo{
				Balance:  big.NewInt(100),
				Nonce:    1,
				CodeHash: B256FromSlice(common.Hash{1}.Bytes()),
				Code:    &Gevm.Bytecode{Kind: Gevm.LegacyRawKind, LegacyRaw: []byte{1, 2, 3}},
			},
			expectedError: false,
		},
		{
			name:    "precompile address",
			address: Address{Addr: common.HexToAddress("0x0000000000000000000000000000000000000001")},
			mockSetup: func(m *MockExecutionClient) {
				// No mock setup needed as precompile shouldn't call GetAccount
			},
			expectedInfo:  Gevm.AccountInfo{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockExecutionClient)
			tt.mockSetup(mockClient)

			db, err := NewProofDB(BlockTag{Number: 1}, mockClient)
			assert.NoError(t, err)

			info, err := db.Basic(tt.address)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if !isPrecompile(tt.address) {
					assert.Equal(t, tt.expectedInfo.Balance.String(), info.Balance.String())
					assert.Equal(t, tt.expectedInfo.Nonce, info.Nonce)
					assert.Equal(t, tt.expectedInfo.CodeHash, info.CodeHash)
					assert.Equal(t, tt.expectedInfo.Code, info.Code)
				}
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestProofDB_Storage(t *testing.T) {
	tests := []struct {
		name          string
		address       Address
		slot          *big.Int
		mockSetup     func(*MockExecutionClient)
		expectedValue *big.Int
		expectedError bool
	}{
		{
			name:    "successful storage fetch",
			address: createTestAddress(1),
			slot:    big.NewInt(1),
			mockSetup: func(m *MockExecutionClient) {
				slotHash := common.BigToHash(big.NewInt(1))
				m.On("GetAccount",
					mock.MatchedBy(func(addr *Address) bool {
						return addr.Addr[19] == 1
					}),
					mock.MatchedBy(func(slots []common.Hash) bool {
						return len(slots) == 1 && slots[0] == slotHash
					}),
					mock.AnythingOfType("BlockTag")).
					Return(Account{
						Slots: map[common.Hash]*big.Int{
							slotHash: (big.NewInt(100)),
						},
					}, nil).Once()
			},
			expectedValue: big.NewInt(100),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockExecutionClient)
			tt.mockSetup(mockClient)

			db, err := NewProofDB(BlockTag{Number: 1}, mockClient)
			assert.NoError(t, err)

			// First call should return "state missing"
			value, err := db.Storage(tt.address, tt.slot)
			assert.Error(t, err)
			assert.Equal(t, "state missing", err.Error())

			// Update state
			err = db.State.UpdateState()
			assert.NoError(t, err)

			// Second call should return the value
			value, err = db.Storage(tt.address, tt.slot)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestProofDB_BlockHash(t *testing.T) {
	tests := []struct {
		name          string
		blockNum      uint64
		mockSetup     func(*MockExecutionClient)
		expectedHash  B256
		expectedError bool
	}{
		{
			name:     "successful block hash fetch",
			blockNum: 1,
			mockSetup: func(m *MockExecutionClient) {
				m.On("GetBlock",
					BlockTag{Number: 1},
					false).
					Return(seleneCommon.Block{
						Hash: common.Hash{1},
					}, nil).Once()
			},
			expectedHash:  B256FromSlice(common.Hash{1}.Bytes()),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockExecutionClient)
			tt.mockSetup(mockClient)

			db, err := NewProofDB(BlockTag{Number: 1}, mockClient)
			assert.NoError(t, err)

			// First call should return "state missing"
			hash, err := db.BlockHash(tt.blockNum)
			assert.Error(t, err)
			assert.Equal(t, "state missing", err.Error())

			// Update state
			err = db.State.UpdateState()
			assert.NoError(t, err)

			// Second call should return the hash
			hash, err = db.BlockHash(tt.blockNum)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHash, hash)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestProofDB_PrefetchState(t *testing.T) {
	mockClient := new(MockExecutionClient)

	// Setup test data
	testAddr1 := createTestAddress(1)
	testAddr2 := createTestAddress(2)

	opts := &CallOpts{
		From: &testAddr1,
		To:   &testAddr2,
	}

	// Mock the access list creation
	mockClient.On("CreateAccessList", mock.Anything, mock.Anything).
		Return([]AccessListItem{
			{
				Address:     testAddr1,
				StorageKeys: []common.Hash{common.Hash{1}},
			},
		}, nil)

	// Mock getting block for coinbase
	mockClient.On("GetBlock", mock.Anything, false).
		Return(seleneCommon.Block{
			Miner: createTestAddress(3),
		}, nil)

	// Mock getting account information
	mockClient.On("GetAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(Account{
			Balance:  big.NewInt(100),
			Nonce:    1,
			Code:     []byte{1, 2, 3},
			CodeHash: common.Hash{1},
			Slots: map[common.Hash]*big.Int{
				common.Hash{1}: common.Hash{2},
			},
		}, nil)

	db, err := NewProofDB(BlockTag{Number: 1}, mockClient)
	assert.NoError(t, err)

	err = db.State.PrefetchState(opts)
	assert.NoError(t, err)

	// Verify the state was prefetched correctly
	info, err := db.Basic(testAddr1)
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(100), info.Balance)

	mockClient.AssertExpectations(t)
}

func TestProofDB_ConcurrentAccess(t *testing.T) {
	mockClient := new(MockExecutionClient)

	// Setup mock responses
	mockClient.On("GetAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(Account{
			Balance:  big.NewInt(100),
			Nonce:    1,
			CodeHash: common.Hash{1},
		}, nil)

	db, err := NewProofDB(BlockTag{Number: 1}, mockClient)
	assert.NoError(t, err)

	// Run concurrent accesses
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			addr := createTestAddress(byte(i))
			_, err := db.Basic(addr)
			assert.NoError(t, err)

			err = db.State.UpdateState()
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestEvmState_NeedsUpdate(t *testing.T) {
	mockClient := new(MockExecutionClient)
	state := NewEvmState(mockClient, BlockTag{Number: 1})

	assert.False(t, state.NeedsUpdate(), "Fresh state should not need update")

	// Trigger need for update
	addr := createTestAddress(1)
	_, err := state.GetBasic(addr)
	assert.Error(t, err)
	assert.True(t, state.NeedsUpdate(), "State should need update after GetBasic call")

	// Mock the update
	mockClient.On("GetAccount", mock.Anything, mock.Anything, mock.Anything).
		Return(Account{}, nil)

	err = state.UpdateState()
	assert.NoError(t, err)
	assert.False(t, state.NeedsUpdate(), "State should not need update after UpdateState")
}

func TestIsPrecompile(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		want    bool
	}{
		{
			name:    "zero address",
			address: Address{},
			want:    false,
		},
		{
			name: "precompile address 1",
			address: Address{Addr: common.HexToAddress("0x0000000000000000000000000000000000000001")},
			want: true,
		},
		{
			name: "precompile address 9",
			address: Address{Addr: common.HexToAddress("0x0000000000000000000000000000000000000009")},
			want: true,
		},
		{
			name: "normal contract address",
			address: Address{Addr: common.HexToAddress("0x1234567890123456789012345678901234567890")},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrecompile(tt.address)
			assert.Equal(t, tt.want, got)
		})
	}
}*/

func CreateNewState() *State {
    blockChan := make(chan *common.Block)
    finalizedBlockChan := make(chan *common.Block)

    state := NewState(5, blockChan, finalizedBlockChan)

    // Simulate blocks to push
    block1 := &common.Block{
        Number: 1,
        Hash:   [32]byte{0x1},
        Transactions: common.Transactions{
            Hashes: [][32]byte{{0x11}, {0x12}},
        },
    }
    block2 := &common.Block{
        Number: 2,
        Hash:   [32]byte{0x2},
        Transactions: common.Transactions{
            Hashes: [][32]byte{{0x21}, {0x22}},
            Full: []common.Transaction{
                {
                    Hash: Common.Hash([32]byte{0x21}),
                    GasPrice: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
                    Gas: hexutil.Uint64(5),
                    MaxFeePerGas: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
            },
                {Hash: Common.Hash([32]byte{0x22})},
            },
        },
    }

    block3 := &common.Block{
        Number: 1000000,
        Hash:   Common.HexToHash("0x8e38b4dbf6b11fcc3b9dee84fb7986e29ca0a02cecd8977c161ff7333329681e"),
        Transactions: common.Transactions{
            Hashes: [][32]byte{
                Common.HexToHash("0xe9e91f1ee4b56c0df2e9f06c2b8c27c6076195a88a7b8537ba8313d80e6f124e"),
                Common.HexToHash("0xea1093d492a1dcb1bef708f771a99a96ff05dcab81ca76c31940300177fcf49f"),
                },
            Full: []common.Transaction{
                {
                    Hash: Common.HexToHash("0xe9e91f1ee4b56c0df2e9f06c2b8c27c6076195a88a7b8537ba8313d80e6f124e"),
                    GasPrice: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
                    Gas: hexutil.Uint64(60),
                    MaxFeePerGas: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
            },
            {
                Hash: Common.HexToHash("0xea1093d492a1dcb1bef708f771a99a96ff05dcab81ca76c31940300177fcf49f"),
                GasPrice: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
                Gas: hexutil.Uint64(60),
                MaxFeePerGas: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
        },
            },
        },
    }

    // Push blocks through channel
    go func() {
        blockChan <- block1
        blockChan <- block2
        blockChan <- block3
        close(blockChan)
    }()

    // Allow goroutine to process the blocks
    wg := sync.WaitGroup{}
    wg.Add(1)
    go func() {
        defer wg.Done()
        for len(state.blocks) < 3 {
            // wait for blocks to be processed
        }
    }()

    wg.Wait()

    // Simulate finalized block
    finalizedBlock := &common.Block{
        Number: 2,
        Hash:   [32]byte{0x2},
        Transactions: common.Transactions{
            Hashes: [][32]byte{{0x21}, {0x22}},
            Full: []common.Transaction{
                {
                    Hash: Common.Hash([32]byte{0x21}),
                    GasPrice: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
                    Gas: hexutil.Uint64(5),
                    MaxFeePerGas: hexutil.Big(*hexutil.MustDecodeBig("0x12345")),
            },
                {Hash: Common.Hash([32]byte{0x22})},
            },
        },
    }
    go func() {
        finalizedBlockChan <- finalizedBlock
        close(finalizedBlockChan)
    }()

    // Wait for finalized block to be processed
    wg.Add(1)
    go func() {
        defer wg.Done()
        for state.finalizedBlock == nil {
            // wait for finalized block to be processed
        }
    }()
    wg.Wait()

    return state
}
func CreateNewExecutionClient() *ExecutionClient {
	rpc := "https://eth-mainnet.g.alchemy.com/v2/j28GcevSYukh-GvSeBOYcwHOfIggF1Gt"
    state := CreateNewState()
	var executionClient *ExecutionClient
	executionClient, _ = executionClient.New(rpc, state)

	return executionClient
}