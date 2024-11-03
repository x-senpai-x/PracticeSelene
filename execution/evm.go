package execution

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strconv"
	"sync"

	seleneCommon "github.com/BlocSoc-iitr/selene/common"
	Gevm "github.com/BlocSoc-iitr/selene/execution/evm"
	"github.com/BlocSoc-iitr/selene/execution/logging"
	"github.com/ethereum/go-ethereum/common" //geth seleneCommon imported as common
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"

	// "bytes"
	// "fmt"
	// "math/big"

	// "errors"

	// seleneCommon "github.com/BlocSoc-iitr/selene/common"
	"github.com/BlocSoc-iitr/selene/utils"
	"github.com/ethereum/go-ethereum"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"

	"github.com/BlocSoc-iitr/selene/consensus/consensus_core"
	consensusTypes "github.com/BlocSoc-iitr/selene/consensus/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

type BlockTag = seleneCommon.BlockTag
type U256 = *big.Int
type B256 = common.Hash
type Address = seleneCommon.Address
type Evm struct {
	execution *ExecutionClient
	chainID   uint64
	tag       BlockTag
}

func NewEvm(execution *ExecutionClient, chainID uint64, tag BlockTag) *Evm {
	return &Evm{
		execution: execution,
		chainID:   chainID,
		tag:       tag,
	}
}
func (e *Evm) Call(opts *CallOpts) ([]byte, error) {
	tx, err := e.callInner(opts) // Call the already implemented call_inner method
	if err != nil {
		return nil, err // Return early if there's an error
	}

	switch tx.Type {
	case "Success":
		// Extract output data and return it as a byte slice
		return tx.Output.Data, nil
	case "Revert":
		return nil, &EvmError{
			Kind:    "Revert",
			Details: tx.Output.Data, // Assuming tx.Output.Data holds the revert reason
		}
	case "Halt":
		return nil, &EvmError{
			Kind:    "Revert",
			Details: nil, // No details for Halt
		}
	default:
		return nil, errors.New("unknown execution result type")
	}
}

/*
	func (e *Evm) Call(opts *CallOpts) ([]byte, error) {
	    tx, err := e.callInner(opts)
	    if err != nil {
	        return nil, err
	    }
	    switch tx.Result.Type {
	    case "Success":
	        return tx.Result.Output.Data, nil
	    case "Revert":
	        return nil, evm.EvmError{
	            Message: "Revert",
	            Data:    tx.Result.Output.Data,
	        }
	    case "Halt":
	        return nil, evm.EvmError{
	            Message: "Revert",
	        }
	    default:
	        return nil, evm.EvmError{
	            Message: "Unknown execution result type",
	        }
	    }
	}
*/
func (e *Evm) EstimateGas(opts *CallOpts) (uint64, error) {
	tx, err := e.callInner(opts) // Call the already implemented call_inner method
	if err != nil {
		return 0, err // Return early if there's an error
	}

	switch tx.Type {
	case "Success":
		return tx.GasUsed, nil // Return the gas used on success
	case "Revert":
		return tx.GasUsed, nil // Return the gas used on revert
	case "Halt":
		return tx.GasUsed, nil // Return the gas used on halt
	default:
		return 0, fmt.Errorf("unknown execution result type: %s", tx.Type) // Handle unknown types
	}
}

type SimpleExternal struct{}

// Implement ExternalType interface
func (e *SimpleExternal) ExternalMethod() {}
func (e *Evm) callInner(opts *CallOpts) (*Gevm.ExecutionResult, error) {
	db, err := NewProofDB(e.tag, e.execution)
	if err != nil {
		return nil, err
	}
	if err := db.State.PrefetchState(opts); err != nil {
		return nil, err
	}
	env := e.getEnv(opts, e.tag)
	var defaultBuilder = Gevm.NewDefaultEvmBuilder[SimpleExternal]()
	evmBuilder := Gevm.WithNewDB(defaultBuilder, db).WithEnv(&env)
	evm := evmBuilder.Build()

	//ctx:=evm.IntoContextWithHandlerCfg()
	// ctx is a ContextWithHandlerCfg instance created from evm using IntoContextWithHandlerCfg.
	ctx := evm.IntoContextWithHandlerCfg()
	var txRes *Gevm.ExecutionResult
	//ctx:=evm.IntoContextWithHandlerCfg[interface{}, ProofDB](evm)
	for {
		db := ctx.Context.Evm.Inner.DB
		if db.State.NeedsUpdate() {
			if err := db.State.UpdateState(); err != nil {
				return nil, err
			}
		}
		//evm := Gevm.NewDefaultEvmBuilder[SimpleExternal]().WithContextWithHandlerCfg(ctx).Build()
		defaultBuilder := Gevm.NewDefaultEvmBuilder[SimpleExternal]()
		evm := Gevm.WithContextWithHandlerCfg(defaultBuilder, ctx).Build()
		res := evm.Transact()
		ctx = evm.IntoContextWithHandlerCfg()
		if res.Err != nil {
			txRes = &res.Value.Result
			break
		}
	}
	if txRes == nil {
		return nil, &EvmError{Kind: "evm error"}
	}
	return txRes, nil
}
func (e *Evm) getEnv(opts *CallOpts, tag BlockTag) Gevm.Env {
	env := Gevm.NewEnv() //needs to be implemented
	//env.Tx.Transact_to=evm.TransactTo:Call
	env.Tx.Caller = *opts.From //is conversion required?
	env.Tx.Value = opts.Value
	env.Tx.Data = opts.Data
	env.Tx.GasLimit = opts.Gas.Uint64()
	env.Tx.GasPrice = opts.GasPrice
	block, err := e.execution.GetBlock(tag, false)
	if err != nil {
		log.Printf("Error getting block: %v", err)
		return Gevm.Env{}
	}
	env.Block.Number = new(big.Int).SetUint64(block.Number)
	env.Block.Coinbase = block.Miner
	env.Block.Timestamp = new(big.Int).SetUint64(block.Timestamp)
	env.Block.Difficulty = block.Difficulty.ToBig()
	env.Cfg.ChainID = e.chainID
	return *env
}

type ProofDB struct {
	State *EvmState
}

func NewProofDB(tag BlockTag, execution *ExecutionClient) (*ProofDB, error) {
	state := NewEvmState(execution, tag)
	return &ProofDB{
		State: state,
	}, nil
}

type StateAccess struct {
	Basic     *Address
	BlockHash *uint64
	Storage   *struct {
		Address Address
		Slot    U256
	}
}
type EvmState struct {
	Basic     map[Address]Gevm.AccountInfo
	BlockHash map[uint64]B256
	Storage   map[Address]map[U256]U256
	Block     BlockTag
	Access    *StateAccess
	Execution *ExecutionClient
	mu        sync.RWMutex
}

func NewEvmState(execution *ExecutionClient, block BlockTag) *EvmState {
	return &EvmState{
		Basic:     make(map[Address]Gevm.AccountInfo),
		BlockHash: make(map[uint64]B256),
		Storage:   make(map[Address]map[U256]U256),
		Block:     block,
		Access:    nil,
		Execution: execution,
	}
}
func (e *EvmState) UpdateState() error {
	e.mu.Lock()
	if e.Access == nil {
		e.mu.Unlock()
		return nil
	}
	access := e.Access
	e.Access = nil
	e.mu.Unlock()

	switch {
	case access.Basic != nil:
		account, err := e.Execution.GetAccount(access.Basic, &[]common.Hash{}, e.Block)
		if err != nil {
			return err
		}
		e.mu.Lock()
		bytecode := NewRawBytecode(account.Code)
		codeHash := B256FromSlice(account.CodeHash[:])
		balance := ConvertU256(account.Balance)
		accountInfo := Gevm.AccountInfo{
			Balance:  balance,
			Nonce:    account.Nonce,
			CodeHash: codeHash,
			Code:     &bytecode,
		}
		e.Basic[*access.Basic] = accountInfo
		e.mu.Unlock()
	case access.Storage != nil:
		slotHash := common.BytesToHash(access.Storage.Slot.Bytes())
		slots := []common.Hash{slotHash}
		account, err := e.Execution.GetAccount(&access.Storage.Address, &slots, e.Block)
		if err != nil {
			return err
		}
		e.mu.Lock()
		storage, ok := e.Storage[access.Storage.Address]
		if !ok {
			storage = make(map[U256]U256)
			e.Storage[access.Storage.Address] = storage
		}

		var slotValue *big.Int
		found := false
		for _, slot := range account.Slots {
			if slot.Key == slotHash {
				slotValue = slot.Value
				found = true
				break
			}
		}
		if !found {
			e.mu.Unlock()
			return fmt.Errorf("storage slot %v not found in account", slotHash)
		}

		value := U256FromBigEndian(slotValue.Bytes())
		storage[access.Storage.Slot] = value
		e.mu.Unlock()

	case access.BlockHash != nil:
		block, err := e.Execution.GetBlock(BlockTag{Number: *access.BlockHash}, false)
		if err != nil {
			return err
		}

		e.mu.Lock()
		hash := B256FromSlice(block.Hash[:])
		e.BlockHash[*access.BlockHash] = hash
		e.mu.Unlock()

	default:
		return errors.New("invalid access type")
	}
	return nil
}

func (e *EvmState) NeedsUpdate() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Access != nil
}
func (e *EvmState) GetBasic(address Address) (Gevm.AccountInfo, error) {
	e.mu.RLock()
	account, exists := e.Basic[address]
	e.mu.RUnlock()

	if exists {
		return account, nil
	}

	e.mu.Lock()
	e.Access = &StateAccess{Basic: &address}
	e.mu.Unlock()

	return Gevm.AccountInfo{}, fmt.Errorf("state missing")
}

func (e *EvmState) GetStorage(address Address, slot U256) (U256, error) {
	// Lock for reading
	e.mu.RLock()
	storage, exists := e.Storage[address]
	if exists {
		if value, exists := storage[slot]; exists {
			e.mu.RUnlock()
			return value, nil
		}
	}
	e.mu.RUnlock()

	// If we need to update state, we need a write lock
	e.mu.Lock()
	// Set the access pattern for state update
	e.Access = &StateAccess{
		Storage: &struct {
			Address Address
			Slot    U256
		}{
			Address: address,
			Slot:    slot,
		},
	}
	e.mu.Unlock()

	// Return zero value and error to indicate state needs to be updated
	return &big.Int{}, errors.New("state missing")
}
func (e *EvmState) GetBlockHash(block uint64) (B256, error) {
	e.mu.RLock()
	hash, exists := e.BlockHash[block]
	e.mu.RUnlock()

	if exists {
		return hash, nil
	}

	e.mu.Lock()
	e.Access = &StateAccess{BlockHash: &block}
	e.mu.Unlock()

	return B256{}, fmt.Errorf("state missing")
}

// PrefetchState prefetches state data
func (e *EvmState) PrefetchState(opts *CallOpts) error {
	list, err := e.Execution.Rpc.CreateAccessList(*opts, e.Block)
	if err != nil {
		return fmt.Errorf("create access list: %w", err)
	}

	// Create access entries
	fromAccessEntry := AccessListItem{
		Address:     [20]byte(*opts.From),
		StorageKeys: make([]common.Hash, 0),
	}
	toAccessEntry := AccessListItem{
		Address:     [20]byte(*opts.To),
		StorageKeys: make([]common.Hash, 0),
	}

	block, err := e.Execution.GetBlock(e.Block, false)
	if err != nil {
		return fmt.Errorf("get block: %w", err)
	}

	producerAccessEntry := AccessListItem{
		Address:     [20]byte(block.Miner),
		StorageKeys: make([]common.Hash, 0),
	}

	// Use a map for O(1) lookup of addresses
	listAddresses := make(map[Address]struct{})
	for _, item := range list.AccessList {
		listAddresses[[20]byte(item.Address)] = struct{}{}
	}

	// Add missing entries
	if _, exists := listAddresses[[20]byte(fromAccessEntry.Address)]; !exists {
		list.AccessList = append(list.AccessList, fromAccessEntry)
	}
	if _, exists := listAddresses[[20]byte(toAccessEntry.Address)]; !exists {
		list.AccessList = append(list.AccessList, toAccessEntry)
	}
	if _, exists := listAddresses[[20]byte(producerAccessEntry.Address)]; !exists {
		list.AccessList = append(list.AccessList, producerAccessEntry)
	}

	// Process accounts in parallel with bounded concurrency
	type accountResult struct {
		address Address
		account Account
		err     error
	}

	batchSize := PARALLEL_QUERY_BATCH_SIZE
	resultChan := make(chan accountResult, len(list.AccessList))
	semaphore := make(chan struct{}, batchSize)

	var wg sync.WaitGroup
	for _, item := range list.AccessList {
		wg.Add(1)
		go func(item AccessListItem) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			account, err := e.Execution.GetAccount((*seleneCommon.Address)(item.Address[:]), &item.StorageKeys, e.Block)
			resultChan <- accountResult{
				address: [20]byte(item.Address),
				account: account,
				err:     err,
			}
		}(item)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results and update state
	e.mu.Lock()
	defer e.mu.Unlock()

	for result := range resultChan {
		if result.err != nil {
			continue
		}

		account := result.account
		address := result.address

		// Update basic account info
		bytecode := NewRawBytecode(account.Code)
		codeHash := B256FromSlice(account.CodeHash[:])
		balance := ConvertU256(account.Balance)
		info := Gevm.NewAccountInfo(balance, account.Nonce, codeHash, bytecode)
		e.Basic[address] = info

		// Update storage
		storage := e.Storage[address]
		if storage == nil {
			storage = make(map[U256]U256)
			e.Storage[address] = storage
		}

		for _, slot := range account.Slots {
			slotHash := U256FromBytes(slot.Key.Bytes())
			valueU256 := ConvertU256(slot.Value)
			storage[slotHash] = valueU256
		}
	}

	return nil
}
func U256FromBytes(b []byte) U256 {
	return new(big.Int).SetBytes(b)
}
func B256FromSlice(slice []byte) common.Hash {
	return common.BytesToHash(slice)
}
func ConvertU256(value *big.Int) *big.Int {
	valueSlice := make([]byte, 32)
	value.FillBytes(valueSlice)
	result := new(big.Int).SetBytes(valueSlice)
	return result
}
func NewRawBytecode(raw []byte) Gevm.Bytecode {
	return Gevm.Bytecode{
		Kind:      Gevm.LegacyRawKind,
		LegacyRaw: raw,
	}
}
func U256FromBigEndian(b []byte) *big.Int {
	if len(b) != 32 {
		return nil // or handle the error appropriately
	}
	return new(big.Int).SetBytes(b)
}
func NewBytecodeRaw(code []byte) hexutil.Bytes {
	return hexutil.Bytes(code)
}

func isPrecompile(address Address) bool {
	precompileAddress := common.BytesToAddress([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09})
	zeroAddress := seleneCommon.Address{}
	return bytes.Compare(address[:], precompileAddress[:]) <= 0 && bytes.Compare(address[:], zeroAddress[:]) > 0
}

// ProofDb implements the Gevm.DB interface
func (db *ProofDB) Basic(address Address) (Gevm.AccountInfo, error) {
	if isPrecompile(address) {
		return Gevm.AccountInfo{}, nil // Return a default AccountInfo
	}
	//logging.Trace("fetch basic evm state for address", zap.String("address", address.Hex()))
	logging.Trace("fetch basic evm state for address", zap.String("address", hex.EncodeToString(address[:])))
	return db.State.GetBasic(address)
}
func (db *ProofDB) BlockHash(number uint64) (B256, error) {
	logging.Trace("fetch block hash for block number", zap.Uint64("number", number))
	return db.State.GetBlockHash(number)
}
func (db *ProofDB) Storage(address Address, slot *big.Int) (*big.Int, error) {
	logging.Trace("fetch storage for address and slot",
		zap.String("address", hex.EncodeToString(address[:])),
		zap.String("slot", slot.String()))
	return db.State.GetStorage(address, slot)
}
func (db *ProofDB) CodeByHash(_ B256) (Gevm.Bytecode, error) {
	return Gevm.Bytecode{}, errors.New("should never be called")
}

/////////////////////////////////////////////////

////////////////////////////////////////////
/////////// execution.go ///////////////////
////////////////////////////////////////////

// package execution

// import (
// 	"bytes"
// 	"fmt"
// 	"math/big"

// 	"errors"

// 	seleneCommon "github.com/BlocSoc-iitr/selene/common"
// 	"github.com/BlocSoc-iitr/selene/utils"
// 	"github.com/ethereum/go-ethereum"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/crypto"
// 	"github.com/ethereum/go-ethereum/rlp"
// 	"github.com/holiman/uint256"
// )

const MAX_SUPPORTED_LOGS_NUMBER = 5
const KECCAK_EMPTY = "0x"

type ExecutionClient struct {
	Rpc   ExecutionRpc
	state *State
}

func (e *ExecutionClient) New(rpc string, state *State) (*ExecutionClient, error) {
	// This is updated as earlier, when ExecutionRpc.New() was called, it was giving an
	// invalid address or nil pointer dereference error as there wasn't a concrete type that implemented ExecutionRpc
	var r ExecutionRpc
	r, err := (&HttpRpc{}).New(&rpc)
	if err != nil {
		return nil, err
	}
	return &ExecutionClient{
		Rpc:   r,
		state: state,
	}, nil
}

// CheckRpc checks the chain ID against the expected value
func (e *ExecutionClient) CheckRpc(chainID uint64) error {
	resultChan := make(chan struct {
		id  uint64
		err error
	})
	go func() {
		rpcChainID, err := e.Rpc.ChainId()
		resultChan <- struct {
			id  uint64
			err error
		}{rpcChainID, err}
	}()
	result := <-resultChan
	if result.err != nil {
		return result.err
	}
	if result.id != chainID {
		return NewIncorrectRpcNetworkError()
	}
	return nil
}

// GetAccount retrieves the account information
func (e *ExecutionClient) GetAccount(address *seleneCommon.Address, slots *[]common.Hash, tag seleneCommon.BlockTag) (Account, error) { //Account from execution/types.go
	block := e.state.GetBlock(tag)
	// Error Handling
	proof, err := e.Rpc.GetProof(address, slots, block.Number)
	if err != nil {
		return Account{}, err
	}
	accountPath := crypto.Keccak256(address[:])
	accountEncoded, err := EncodeAccount(&proof)
	if err != nil {
		return Account{}, err
	}
	accountProofBytes := make([][]byte, len(proof.AccountProof))
	for i, hexByte := range proof.AccountProof {
		accountProofBytes[i] = hexByte
	}
	isValid, err := VerifyProof(accountProofBytes, block.StateRoot[:], accountPath, accountEncoded)
	if err != nil {
		return Account{}, err
	}
	if !isValid {
		return Account{}, NewInvalidAccountProofError(*address)
	}
	// modify
	slotMap := []Slot{}
	for _, storageProof := range proof.StorageProof {
		key, err := utils.Hex_str_to_bytes(storageProof.Key.Hex())
		if err != nil {
			return Account{}, err
		}
		value, err := rlp.EncodeToBytes(storageProof.Value)
		if err != nil {
			return Account{}, err
		}
		keyHash := crypto.Keccak256(key)
		proofBytes := make([][]byte, len(storageProof.Proof))
		for i, hexByte := range storageProof.Proof {
			proofBytes[i] = hexByte
		}
		isValid, err := VerifyProof(
			proofBytes,
			proof.StorageHash.Bytes(),
			keyHash,
			value,
		)
		if err != nil {
			return Account{}, err
		}
		if !isValid {
			return Account{}, fmt.Errorf("invalid storage proof for address: %v, key: %v", *address, storageProof.Key)
		}
		slotMap = append(slotMap, Slot{
			Key:   storageProof.Key,
			Value: storageProof.Value.ToBig(),
		})
	}
	var code []byte
	if bytes.Equal(proof.CodeHash.Bytes(), crypto.Keccak256([]byte(KECCAK_EMPTY))) {
		code = []byte{}
	} else {
		code, err := e.Rpc.GetCode(address, block.Number)
		if err != nil {
			return Account{}, err
		}
		codeHash := crypto.Keccak256(code)
		if !bytes.Equal(proof.CodeHash.Bytes(), codeHash) {
			return Account{}, fmt.Errorf("code hash mismatch for address: %v, expected: %v, got: %v",
				*address, common.BytesToHash(codeHash).String(), proof.CodeHash.String())
		}
	}
	account := Account{
		Balance:     proof.Balance.ToBig(),
		Nonce:       uint64(proof.Nonce),
		Code:        code,
		CodeHash:    proof.CodeHash,
		StorageHash: proof.StorageHash,
		Slots:       slotMap,
	}
	return account, nil
}
func (e *ExecutionClient) SendRawTransaction(bytes []byte) (common.Hash, error) {
	var txHash common.Hash
	var err error
	done := make(chan bool)
	go func() {
		txHash, err = e.Rpc.SendRawTransaction(&bytes)
		done <- true
	}()
	<-done
	return txHash, err
}

/*
	func (e *ExecutionClient) GetBlock(tag seleneCommon.BlockTag, full_tx bool) (seleneCommon.Block, error) {
		blockChan := make(chan seleneCommon.Block)
		errChan := make(chan error)
		go func() {
			block := e.state.GetBlock(tag)
			blockChan <- *block
		}()
		select {
		case block := <-blockChan:
			if !full_tx {
				block.Transactions = seleneCommon.Transactions{Hashes: block.Transactions.HashesFunc()}
			}
			return block, nil
		case err := <-errChan:
			return seleneCommon.Block{}, err
		}
	}
*/
func (e *ExecutionClient) GetBlock(tag seleneCommon.BlockTag, full_tx bool) (seleneCommon.Block, error) {
	blockChan := make(chan seleneCommon.Block)
	errChan := make(chan error)

	go func() {
		block := e.state.GetBlock(tag)
		if block == nil {
			errChan <- fmt.Errorf("block not found for tag: %v", tag)
			return
		}
		blockChan <- *block
	}()

	select {
	case block := <-blockChan:
		if !full_tx {
			block.Transactions = seleneCommon.Transactions{Hashes: block.Transactions.HashesFunc()}
		}
		return block, nil
	case err := <-errChan:
		return seleneCommon.Block{}, err
	}
}
func (e *ExecutionClient) GetBlockByHash(hash common.Hash, full_tx bool) (seleneCommon.Block, error) {
	blockChan := make(chan seleneCommon.Block)
	errChan := make(chan error)
	go func() {
		block := e.state.GetBlockByHash(hash)
		blockChan <- *block
	}()
	select {
	case block := <-blockChan:
		if !full_tx {
			block.Transactions = seleneCommon.Transactions{Hashes: block.Transactions.HashesFunc()}
		}
		return block, nil
	case err := <-errChan:
		return seleneCommon.Block{}, err
	}
}
func (e *ExecutionClient) GetTransactionByBlockHashAndIndex(blockHash common.Hash, index uint64) (seleneCommon.Transaction, error) {
	txChan := make(chan seleneCommon.Transaction)
	errChan := make(chan error)
	go func() {
		tx := e.state.GetTransactionByBlockAndIndex(blockHash, index)
		txChan <- *tx
	}()
	select {
	case tx := <-txChan:
		return tx, nil
	case err := <-errChan:
		return seleneCommon.Transaction{}, err
	}

}
func (e *ExecutionClient) GetTransactionReceipt(txHash common.Hash) (types.Receipt, error) {
	receiptChan := make(chan types.Receipt)
	errChan := make(chan error)
	// var receipt types.Receipt
	go func() {
		receipt, err := e.Rpc.GetTransactionReceipt(&txHash)
		if err != nil {
			errChan <- err
			return
		}
		receiptChan <- receipt
	}()
	select {
	case receipt := <-receiptChan:
		blocknumber := receipt.BlockNumber
		blockChan := make(chan seleneCommon.Block)
		errChan := make(chan error)
		go func() {
			block := e.state.GetBlock(seleneCommon.BlockTag{Number: blocknumber.Uint64()})
			blockChan <- *block
		}()
		select {
		case block := <-blockChan:
			txHashes := block.Transactions.Hashes
			receiptsChan := make(chan types.Receipt)
			receiptsErrChan := make(chan error)
			for _, hash := range txHashes {
				go func(hash common.Hash) {
					receipt, err := e.Rpc.GetTransactionReceipt(&hash)
					if err != nil {
						receiptsErrChan <- err
						return
					}
					receiptsChan <- receipt
				}(hash)
			}
			var receipts []types.Receipt
			for range txHashes {
				select {
				case receipt := <-receiptsChan:
					receipts = append(receipts, receipt)
				case err := <-receiptsErrChan:
					return types.Receipt{}, err
				}
			}
			var receiptsEncoded [][]byte
			for _, receipt := range receipts {
				encodedReceipt, err := encodeReceipt(&receipt)
				if err != nil {
					receiptsErrChan <- err
					return types.Receipt{}, err
				}
				receiptsEncoded = append(receiptsEncoded, encodedReceipt)
			}
			expectedReceiptRoot, err := CalculateReceiptRoot(receiptsEncoded)
			if err != nil {
				return types.Receipt{}, err
			}

			if [32]byte(expectedReceiptRoot.Bytes()) != block.ReceiptsRoot || !contains(receipts, receipt) {
				return types.Receipt{}, fmt.Errorf("receipt root mismatch: %s", txHash.String())
			}

			return receipt, nil

		case err := <-errChan:
			return types.Receipt{}, err
		}
	case err := <-errChan:
		return types.Receipt{}, err
	}
}
func (e *ExecutionClient) GetTransaction(hash common.Hash) (seleneCommon.Transaction, error) {
	txChan := make(chan seleneCommon.Transaction)
	errChan := make(chan error)
	go func() {
		tx := e.state.GetTransaction(hash)
		txChan <- *tx
	}()
	select {
	case tx := <-txChan:
		return tx, nil
	case err := <-errChan:
		return seleneCommon.Transaction{}, err
	}
}
func (e *ExecutionClient) GetLogs(filter ethereum.FilterQuery) ([]types.Log, error) {
	if filter.ToBlock == nil && filter.BlockHash == nil {
		block := e.state.LatestBlockNumber()
		filter.ToBlock = new(big.Int).SetUint64(*block)
		if filter.FromBlock == nil {
			filter.FromBlock = new(big.Int).SetUint64(*block)
		}
	}
	logsChan := make(chan []types.Log)
	errChan := make(chan error)
	go func() {
		logs, err := e.Rpc.GetLogs(&filter)
		if err != nil {
			errChan <- err
			return
		}
		logsChan <- logs
	}()
	select {
	case logs := <-logsChan:
		if len(logs) > MAX_SUPPORTED_LOGS_NUMBER {
			// The earlier error was not returning properly
			return nil, errors.New("logs exceed max supported logs number")
		}
		logPtrs := make([]*types.Log, len(logs))
		for i := range logs {
			logPtrs[i] = &logs[i]
		}
		if err := e.verifyLogs(logPtrs); err != nil {
			return nil, err
		}

		return logs, nil
	case err := <-errChan:
		return nil, err
	}
}
func (e *ExecutionClient) GetFilterChanges(filterID *uint256.Int) ([]types.Log, error) {
	logsChan := make(chan []types.Log)
	errChan := make(chan error)
	go func() {
		logs, err := e.Rpc.GetFilterChanges(filterID)
		if err != nil {
			errChan <- err
			return
		}
		logsChan <- logs
	}()
	select {
	case logs := <-logsChan:
		if len(logs) > MAX_SUPPORTED_LOGS_NUMBER {
			return nil, &ExecutionError{
				Kind:    "TooManyLogs",
				Details: fmt.Sprintf("Too many logs to prove: %d, max: %d", len(logs), MAX_SUPPORTED_LOGS_NUMBER),
			}
		}
		logPtrs := make([]*types.Log, len(logs))
		for i := range logs {
			logPtrs[i] = &logs[i]
		}
		if err := e.verifyLogs(logPtrs); err != nil {
			return nil, err
		}
		return logs, nil
	case err := <-errChan:
		return nil, err
	}
}
func (e *ExecutionClient) UninstallFilter(filterID *uint256.Int) (bool, error) {
	resultChan := make(chan struct {
		result bool
		err    error
	})
	go func() {
		result, err := e.Rpc.UninstallFilter(filterID)
		resultChan <- struct {
			result bool
			err    error
		}{result, err}
	}()
	result := <-resultChan
	return result.result, result.err
}
func (e *ExecutionClient) GetNewFilter(filter ethereum.FilterQuery) (uint256.Int, error) {
	if filter.ToBlock == nil && filter.BlockHash == nil {
		block := e.state.LatestBlockNumber()
		filter.ToBlock = new(big.Int).SetUint64(*block)
		if filter.FromBlock == nil {
			filter.FromBlock = new(big.Int).SetUint64(*block)
		}
	}
	filterIDChan := make(chan uint256.Int)
	errChan := make(chan error)
	go func() {
		filterID, err := e.Rpc.GetNewFilter(&filter)
		if err != nil {
			errChan <- err
			return
		}
		filterIDChan <- filterID
	}()
	select {
	case filterID := <-filterIDChan:
		return filterID, nil
	case err := <-errChan:
		return uint256.Int{}, err
	}
}
func (e *ExecutionClient) GetNewBlockFilter() (uint256.Int, error) {
	filterIDChan := make(chan uint256.Int)
	errChan := make(chan error)
	go func() {
		filterID, err := e.Rpc.GetNewBlockFilter()
		if err != nil {
			errChan <- err
			return
		}
		filterIDChan <- filterID
	}()
	select {
	case filterID := <-filterIDChan:
		return filterID, nil
	case err := <-errChan:
		return uint256.Int{}, err
	}
}
func (e *ExecutionClient) GetNewPendingTransactionFilter() (uint256.Int, error) {
	filterIDChan := make(chan uint256.Int)
	errChan := make(chan error)
	go func() {
		filterID, err := e.Rpc.GetNewPendingTransactionFilter()
		if err != nil {
			errChan <- err
			return
		}
		filterIDChan <- filterID
	}()
	select {
	case filterID := <-filterIDChan:
		return filterID, nil
	case err := <-errChan:
		return uint256.Int{}, err
	}
}
func (e *ExecutionClient) verifyLogs(logs []*types.Log) error {
	errChan := make(chan error, len(logs))
	for _, log := range logs {
		go func(log *types.Log) {
			receiptSubChan := make(chan *types.Receipt)
			go func() {
				receipt, err := e.Rpc.GetTransactionReceipt(&log.TxHash)
				if err != nil {
					errChan <- err
					return
				}
				receiptSubChan <- &receipt
			}()
			select {
			case receipt := <-receiptSubChan:
				receiptLogsEncoded := make([][]byte, len(receipt.Logs))
				for i, receiptLog := range receipt.Logs {
					receiptLogsEncoded[i] = receiptLog.Data
				}
				logEncoded := log.Data
				found := false
				for _, encoded := range receiptLogsEncoded {
					if string(encoded) == string(logEncoded) {
						found = true
						break
					}
				}
				if !found {
					errChan <- fmt.Errorf("missing log for transaction %s", log.TxHash.Hex())
					return
				}
			case err := <-errChan:
				errChan <- err
				return
			}
			errChan <- nil
		}(log)
	}
	for range logs {
		if err := <-errChan; err != nil {
			return err
		}
	}
	return nil
}
func encodeReceipt(receipt *types.Receipt) ([]byte, error) {
	var stream []interface{}
	stream = append(stream, receipt.Status, receipt.CumulativeGasUsed, receipt.Bloom, receipt.Logs)
	legacyReceiptEncoded, err := rlp.EncodeToBytes(stream)
	if err != nil {
		return nil, err
	}
	txType := &receipt.Type
	if *txType == 0 {
		return legacyReceiptEncoded, nil
	}
	txTypeBytes := []byte{*txType}
	return append(txTypeBytes, legacyReceiptEncoded...), nil
}

// need to confirm if TxHash is actually used as the key to calculate the receipt root or not
func CalculateReceiptRoot(receipts [][]byte) (common.Hash, error) {
	if len(receipts) == 0 {
		return common.Hash{}, errors.New("no receipts to calculate root")
	}

	var receiptHashes []common.Hash
	for _, receipt := range receipts {
		receiptHash, err := rlpHash(receipt)
		if err != nil {
			return common.Hash{}, err
		}
		receiptHashes = append(receiptHashes, receiptHash)
	}
	return calculateMerkleRoot(receiptHashes), nil
}
func rlpHash(obj interface{}) (common.Hash, error) {
	encoded, err := rlp.EncodeToBytes(obj)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(encoded), nil
}

// This function is updated as it was going in an infinite loop
func calculateMerkleRoot(hashes []common.Hash) common.Hash {
	switch len(hashes) {
	case 0:
		return common.Hash{} // Return empty hash for empty slice
	case 1:
		return hashes[0]
	default:
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		var newLevel []common.Hash
		for i := 0; i < len(hashes); i += 2 {
			combinedHash := crypto.Keccak256(append(hashes[i].Bytes(), hashes[i+1].Bytes()...))
			newLevel = append(newLevel, common.BytesToHash(combinedHash))
		}
		return calculateMerkleRoot(newLevel)
	}
}

// contains checks if a receipt is in the list of receipts
func contains(receipts []types.Receipt, receipt types.Receipt) bool {
	for _, r := range receipts {
		if r.TxHash == receipt.TxHash {
			return true
		}
	}
	return false
}

/////////////////////////////////////////
//////////// http_rpc.go ////////////////
/////////////////////////////////////////

// package execution

// import (
// 	"strconv"
// 	"encoding/hex"
// 	"math/big"

// 	seleneCommon "github.com/BlocSoc-iitr/selene/common"
// 	"github.com/ethereum/go-ethereum"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/common/hexutil"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/rpc"
// 	"github.com/holiman/uint256"
// )

type HttpRpc struct {
	url      string
	provider *rpc.Client
}

// I have made some changes to ExecutionRpc and to HttpRpc as HttpRpc was not satisfying
// ExecutionRpc interface before
func (h *HttpRpc) New(rpcUrl *string) (ExecutionRpc, error) {
	client, err := rpc.Dial(*rpcUrl)
	if err != nil {
		return nil, err
	}

	return &HttpRpc{
		url:      *rpcUrl,
		provider: client,
	}, nil
}

// All the changes that have been made in functions that fetch from rpc are because:
// The rpc expects request to be in form of hexadecimal strings. For example, if we want
// to send block number equal to 1405, it will interpret it as 0x1405, which is not it's hex representation
//
// Similarly, the values recieved from rpc should also be treated as hex strings
func (h *HttpRpc) GetProof(address *seleneCommon.Address, slots *[]common.Hash, block uint64) (EIP1186ProofResponse, error) {
	resultChan := make(chan struct {
		proof EIP1186ProofResponse
		err   error
	})
	// All arguments to rpc are expected to be in form of hex strings
	var slotHex []string
	if slots != nil {
		for _, slot := range *slots {
			slotHex = append(slotHex, slot.Hex())
		}
	}
	if len(*slots) == 0 {
		slotHex = []string{}
	}
	go func() {
		var proof EIP1186ProofResponse
		err := h.provider.Call(&proof, "eth_getProof", "0x"+hex.EncodeToString(address[:]), slotHex, toBlockNumArg(block))
		resultChan <- struct {
			proof EIP1186ProofResponse
			err   error
		}{proof, err}
		close(resultChan)
	}()
	result := <-resultChan
	if result.err != nil {
		return EIP1186ProofResponse{}, result.err
	}
	return result.proof, nil
}

// TODO: CreateAccessList is throwing an error
// There is a problem in unmarshaling the response into types.AccessList
func (h *HttpRpc) CreateAccessList(opts CallOpts, block seleneCommon.BlockTag) (AccessList, error) {
	resultChan := make(chan struct {
		accessList AccessList
		err        error
	})
	callOpts := struct {
		From     string        `json:"from"`
		To       string        `json:"to"`
		Gas      *hexutil.Big  `json:"gas,omitempty"`
		GasPrice *hexutil.Big  `json:"gasPrice,omitempty"`
		Value    *hexutil.Big  `json:"value,omitempty"`
		Data     string        `json:"data,omitempty"`
	}{
		From: func() string {
			if opts.From != nil {
				return "0x" + hex.EncodeToString(opts.From[:])
			}
			return ""
		}(),
		To: func() string {
			if opts.To != nil {
				return "0x" + hex.EncodeToString(opts.To[:])
			}
			return "" // or handle cases where `to` is missing if necessary
		}(),
		Gas: func() *hexutil.Big {
			if opts.Gas != nil {
				return (*hexutil.Big)(opts.Gas)
			}
			return nil
		}(),
		GasPrice: func() *hexutil.Big {
			if opts.GasPrice != nil {
				return (*hexutil.Big)(opts.GasPrice)
			}
			return nil
		}(),
		Value: func() *hexutil.Big {
			if opts.Value != nil {
				return (*hexutil.Big)(opts.Value)
			}
			return nil
		}(),
		Data: hex.EncodeToString(opts.Data),
	}	

	go func() {
		var accessList AccessList
		err := h.provider.Call(&accessList, "eth_createAccessList", callOpts, block.String())
		resultChan <- struct {
			accessList AccessList
			err        error
		}{accessList, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return AccessList{}, result.err
	}
	return result.accessList, nil
}

func (h *HttpRpc) GetCode(address *seleneCommon.Address, block uint64) ([]byte, error) {
	resultChan := make(chan struct {
		code hexutil.Bytes
		err  error
	})

	go func() {
		var code hexutil.Bytes
		err := h.provider.Call(&code, "eth_getCode", "0x"+hex.EncodeToString(address[:]), toBlockNumArg(block))
		resultChan <- struct {
			code hexutil.Bytes
			err  error
		}{code, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return nil, result.err
	}
	return result.code, nil
}

func (h *HttpRpc) SendRawTransaction(data *[]byte) (common.Hash, error) {
	resultChan := make(chan struct {
		txHash common.Hash
		err    error
	})

	go func() {
		var txHash common.Hash
		err := h.provider.Call(&txHash, "eth_sendRawTransaction", hexutil.Bytes(*data))
		resultChan <- struct {
			txHash common.Hash
			err    error
		}{txHash, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return common.Hash{}, result.err
	}
	return result.txHash, nil
}

func (h *HttpRpc) GetTransactionReceipt(txHash *common.Hash) (types.Receipt, error) {
	resultChan := make(chan struct {
		receipt types.Receipt
		err     error
	})

	go func() {
		var receipt types.Receipt
		err := h.provider.Call(&receipt, "eth_getTransactionReceipt", txHash)
		resultChan <- struct {
			receipt types.Receipt
			err     error
		}{receipt, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return types.Receipt{}, result.err
	}
	return result.receipt, nil
}

func (h *HttpRpc) GetTransaction(txHash *common.Hash) (seleneCommon.Transaction, error) {
	resultChan := make(chan struct {
		tx  seleneCommon.Transaction
		err error
	})

	go func() {
		var tx seleneCommon.Transaction
		err := h.provider.Call(&tx, "eth_getTransactionByHash", txHash)
		resultChan <- struct {
			tx  seleneCommon.Transaction
			err error
		}{tx, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return seleneCommon.Transaction{}, result.err
	}
	return result.tx, nil
}

func (h *HttpRpc) GetLogs(filter *ethereum.FilterQuery) ([]types.Log, error) {
	resultChan := make(chan struct {
		logs []types.Log
		err  error
	})

	go func() {
		var logs []types.Log
		err := h.provider.Call(&logs, "eth_getLogs", toFilterArg(*filter))
		resultChan <- struct {
			logs []types.Log
			err  error
		}{logs, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return nil, result.err
	}
	return result.logs, nil
}

func (h *HttpRpc) GetFilterChanges(filterID *uint256.Int) ([]types.Log, error) {
	resultChan := make(chan struct {
		logs []types.Log
		err  error
	})

	go func() {
		var logs []types.Log
		err := h.provider.Call(&logs, "eth_getFilterChanges", filterID.Hex())
		resultChan <- struct {
			logs []types.Log
			err  error
		}{logs, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return nil, result.err
	}
	return result.logs, nil
}

func (h *HttpRpc) UninstallFilter(filterID *uint256.Int) (bool, error) {
	resultChan := make(chan struct {
		result bool
		err    error
	})

	go func() {
		var result bool
		err := h.provider.Call(&result, "eth_uninstallFilter", filterID.Hex())
		resultChan <- struct {
			result bool
			err    error
		}{result, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return false, result.err
	}
	return result.result, nil
}

func (h *HttpRpc) GetNewFilter(filter *ethereum.FilterQuery) (uint256.Int, error) {
	resultChan := make(chan struct {
		filterID uint256.Int
		err      error
	})

	go func() {
		var filterID hexutil.Big
		err := h.provider.Call(&filterID, "eth_newFilter", toFilterArg(*filter))
		filterResult := big.Int(filterID)
		resultChan <- struct {
			filterID uint256.Int
			err      error
		}{*uint256.MustFromBig(&filterResult), err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return uint256.Int{}, result.err
	}
	return result.filterID, nil
}

func (h *HttpRpc) GetNewBlockFilter() (uint256.Int, error) {
	resultChan := make(chan struct {
		filterID uint256.Int
		err      error
	})

	go func() {
		var filterID hexutil.Big
		err := h.provider.Call(&filterID, "eth_newBlockFilter")
		filterResult := big.Int(filterID)
		resultChan <- struct {
			filterID uint256.Int
			err      error
		}{*uint256.MustFromBig(&filterResult), err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return uint256.Int{}, result.err
	}
	return result.filterID, nil
}

func (h *HttpRpc) GetNewPendingTransactionFilter() (uint256.Int, error) {
	resultChan := make(chan struct {
		filterID uint256.Int
		err      error
	})

	go func() {
		var filterID hexutil.Big
		err := h.provider.Call(&filterID, "eth_newPendingTransactionFilter")
		filterResult := big.Int(filterID)
		resultChan <- struct {
			filterID uint256.Int
			err      error
		}{*uint256.MustFromBig(&filterResult), err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return uint256.Int{}, result.err
	}
	return result.filterID, nil
}

func (h *HttpRpc) ChainId() (uint64, error) {
	resultChan := make(chan struct {
		chainID uint64
		err     error
	})

	go func() {
		var chainID hexutil.Uint64
		err := h.provider.Call(&chainID, "eth_chainId")
		resultChan <- struct {
			chainID uint64
			err     error
		}{uint64(chainID), err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return 0, result.err
	}
	return result.chainID, nil
}

func (h *HttpRpc) GetFeeHistory(blockCount uint64, lastBlock uint64, rewardPercentiles *[]float64) (FeeHistory, error) {
	resultChan := make(chan struct {
		feeHistory FeeHistory
		err        error
	})

	go func() {
		var feeHistory FeeHistory
		err := h.provider.Call(&feeHistory, "eth_feeHistory", hexutil.Uint64(blockCount).String(), toBlockNumArg(lastBlock), rewardPercentiles)
		resultChan <- struct {
			feeHistory FeeHistory
			err        error
		}{feeHistory, err}
		close(resultChan)
	}()

	result := <-resultChan
	if result.err != nil {
		return FeeHistory{}, result.err
	}
	return result.feeHistory, nil
}

func toBlockNumArg(number uint64) string {
	if number == 0 {
		return "latest"
	}
	return "0x" + strconv.FormatUint(number, 16)
}

func toFilterArg(q ethereum.FilterQuery) map[string]interface{} {
	arg := make(map[string]interface{})
	if len(q.Addresses) > 0 {
		arg["address"] = q.Addresses
	}
	if len(q.Topics) > 0 {
		arg["topics"] = q.Topics
	}
	if q.FromBlock != nil {
		arg["fromBlock"] = toBlockNumArg(q.FromBlock.Uint64())
	}
	if q.ToBlock != nil {
		arg["toBlock"] = toBlockNumArg(q.ToBlock.Uint64())
	}
	return arg
}

///////////////////////////////////////
/////////  proof.go ////////////////////
/////////////////////////////////////////

// package execution

// import (
// 	"bytes"
// 	"fmt"
// 	"github.com/ethereum/go-ethereum/rlp"
// 	"github.com/holiman/uint256"
// 	"golang.org/x/crypto/sha3"
// )

func VerifyProof(proof [][]byte, root []byte, path []byte, value []byte) (bool, error) {
	expectedHash := root
	pathOffset := 0

	for i, node := range proof {
		if !bytes.Equal(expectedHash, keccak256(node)) {
			return false, nil
		}

		var nodeList [][]byte
		if err := rlp.DecodeBytes(node, &nodeList); err != nil {
			fmt.Println("Error decoding node:", err)
			return false, err
		}

		if len(nodeList) == 17 {
			if i == len(proof)-1 {
				// exclusion proof
				nibble := getNibble(path, pathOffset)
				if len(nodeList[nibble]) == 0 && isEmptyValue(value) {
					return true, nil
				}
			} else {
				nibble := getNibble(path, pathOffset)
				expectedHash = nodeList[nibble]
				pathOffset++
			}
		} else if len(nodeList) == 2 {
			if i == len(proof)-1 {
				// exclusion proof
				if !pathsMatch(nodeList[0], skipLength(nodeList[0]), path, pathOffset) && isEmptyValue(value) {
					return true, nil
				}

				// inclusion proof
				if bytes.Equal(nodeList[1], value) {
					return pathsMatch(nodeList[0], skipLength(nodeList[0]), path, pathOffset), nil
				}
			} else {
				nodePath := nodeList[0]
				prefixLength := sharedPrefixLength(path, pathOffset, nodePath)
				if prefixLength < len(nodePath)*2-skipLength(nodePath) {
					// Proof shows a divergent path , but we're not at the leaf yet
					return false, nil
				}
				pathOffset += prefixLength
				expectedHash = nodeList[1]
			}
		} else {
			return false, nil
		}
	}

	return false, nil
}

func pathsMatch(p1 []byte, s1 int, p2 []byte, s2 int) bool {
	len1 := len(p1)*2 - s1
	len2 := len(p2)*2 - s2

	if len1 != len2 {
		return false
	}

	for offset := 0; offset < len1; offset++ {
		n1 := getNibble(p1, s1+offset)
		n2 := getNibble(p2, s2+offset)
		if n1 != n2 {
			return false
		}
	}

	return true
}

// dead code
func GetRestPath(p []byte, s int) string {
	var ret string
	for i := s; i < len(p)*2; i++ {
		n := getNibble(p, i)
		ret += fmt.Sprintf("%01x", n)
	}
	return ret
}

func isEmptyValue(value []byte) bool {
	emptyAccount := Account{
		Nonce:       0,
		Balance:     uint256.NewInt(0).ToBig(),
		StorageHash: [32]byte{0x56, 0xe8, 0x1f, 0x17, 0x1b, 0xcc, 0x55, 0xa6, 0xff, 0x83, 0x45, 0xe6, 0x92, 0xc0, 0xf8, 0x6e, 0x5b, 0x48, 0xe0, 0x1b, 0x99, 0x6c, 0xad, 0xc0, 0x01, 0x62, 0x2f, 0xb5, 0xe3, 0x63, 0xb4, 0x21},
		CodeHash:    [32]byte{0xc5, 0xd2, 0x46, 0x01, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x03, 0xc0, 0xe5, 0x00, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x04, 0x5d, 0x85, 0xa4, 0x70},
	}
	// EncodeToBytes needs pointer as argument
	encodedEmptyAccount, _ := rlp.EncodeToBytes(&emptyAccount)

	isEmptySlot := len(value) == 1 && value[0] == 0x80
	isEmptyAccount := bytes.Equal(value, encodedEmptyAccount)

	return isEmptySlot || isEmptyAccount
}

func sharedPrefixLength(path []byte, pathOffset int, nodePath []byte) int {
	skipLength := skipLength(nodePath)

	len1 := min(len(nodePath)*2-skipLength, len(path)*2-pathOffset)
	prefixLen := 0

	for i := 0; i < len1; i++ {
		pathNibble := getNibble(path, i+pathOffset)
		nodePathNibble := getNibble(nodePath, i+skipLength)
		if pathNibble != nodePathNibble {
			break
		}
		prefixLen++
	}

	return prefixLen
}

func skipLength(node []byte) int {
	if len(node) == 0 {
		return 0
	}

	nibble := getNibble(node, 0)
	switch nibble {
	case 0, 2:
		return 2
	case 1, 3:
		return 1
	default:
		return 0
	}
}

func getNibble(path []byte, offset int) byte {
	byteVal := path[offset/2]
	if offset%2 == 0 {
		return byteVal >> 4
	}
	return byteVal & 0xF
}

func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Updated as EIP1186ProofResponse was updated
func EncodeAccount(proof *EIP1186ProofResponse) ([]byte, error) {
	account := Account{
		Nonce:       uint64(proof.Nonce),
		Balance:     proof.Balance.ToBig(),
		StorageHash: proof.StorageHash,
		CodeHash:    proof.CodeHash,
	}
	// EncodeToBytes needs pointer as argument
	return rlp.EncodeToBytes(&account)
}

// Make a generic function for it
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

////////////////////////////
/////// rpc.go ////////////
//////////////////////////

// Changes have been made as HttpRpc was not satisfying ExecutionRpc
type ExecutionRpc interface {
	New(rpc *string) (ExecutionRpc, error)
	GetProof(address *seleneCommon.Address, slots *[]common.Hash, block uint64) (EIP1186ProofResponse, error)
	CreateAccessList(opts CallOpts, block seleneCommon.BlockTag) (AccessList, error)
	GetCode(address *seleneCommon.Address, block uint64) ([]byte, error)
	SendRawTransaction(bytes *[]byte) (common.Hash, error)
	GetTransactionReceipt(tx_hash *common.Hash) (types.Receipt, error)
	GetTransaction(tx_hash *common.Hash) (seleneCommon.Transaction, error)
	GetLogs(filter *ethereum.FilterQuery) ([]types.Log, error)
	GetFilterChanges(filer_id *uint256.Int) ([]types.Log, error)
	UninstallFilter(filter_id *uint256.Int) (bool, error)
	GetNewFilter(filter *ethereum.FilterQuery) (uint256.Int, error)
	GetNewBlockFilter() (uint256.Int, error)
	GetNewPendingTransactionFilter() (uint256.Int, error)
	ChainId() (uint64, error)
	GetFeeHistory(block_count uint64, last_block uint64, reward_percentiles *[]float64) (FeeHistory, error)
}

////////////////////////////////////
/////////// state.ggo //////////////
////////////////////////////////////

// package execution

// import (
// 	"github.com/BlocSoc-iitr/selene/common"
// 	"github.com/holiman/uint256"
// 	"sync"
// )

type State struct {
	mu             sync.RWMutex
	blocks         map[uint64]*seleneCommon.Block
	finalizedBlock *seleneCommon.Block
	hashes         map[[32]byte]uint64
	txs            map[[32]byte]TransactionLocation
	historyLength  uint64
}
type TransactionLocation struct {
	Block uint64
	Index int
}

func NewState(historyLength uint64, blockChan <-chan *seleneCommon.Block, finalizedBlockChan <-chan *seleneCommon.Block) *State {
	s := &State{
		blocks:        make(map[uint64]*seleneCommon.Block),
		hashes:        make(map[[32]byte]uint64),
		txs:           make(map[[32]byte]TransactionLocation),
		historyLength: historyLength,
	}
	go func() {
		for {
			select {
			case block := <-blockChan:
				if block != nil {
					s.PushBlock(block)
				}
			case block := <-finalizedBlockChan:
				if block != nil {
					s.PushFinalizedBlock(block)
				}
			}
		}
	}()

	return s
}
func (s *State) PushBlock(block *seleneCommon.Block) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes[block.Hash] = block.Number
	for i, txHash := range block.Transactions.Hashes {
		loc := TransactionLocation{
			Block: block.Number,
			Index: i,
		}
		s.txs[txHash] = loc
	}

	s.blocks[block.Number] = block

	for len(s.blocks) > int(s.historyLength) {
		var oldestNumber uint64 = ^uint64(0)
		for number := range s.blocks {
			if number < oldestNumber {
				oldestNumber = number
			}
		}
		s.removeBlock(oldestNumber)
	}
}
func (s *State) PushFinalizedBlock(block *seleneCommon.Block) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.finalizedBlock = block

	if oldBlock, exists := s.blocks[block.Number]; exists {
		if oldBlock.Hash != block.Hash {
			s.removeBlock(oldBlock.Number)
			s.PushBlock(block)
		}
	} else {
		s.PushBlock(block)
	}
}
func (s *State) removeBlock(number uint64) {
	if block, exists := s.blocks[number]; exists {
		delete(s.blocks, number)
		delete(s.hashes, block.Hash)
		for _, txHash := range block.Transactions.Hashes {
			delete(s.txs, txHash)
		}
	}
}
func (s *State) GetBlock(tag seleneCommon.BlockTag) *seleneCommon.Block {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tag.Latest {
		var latestNumber uint64
		var latestBlock *seleneCommon.Block
		for number, block := range s.blocks {
			if number > latestNumber {
				latestNumber = number
				latestBlock = block
			}
		}
		return latestBlock
	} else if tag.Finalized {
		return s.finalizedBlock
	} else {
		return s.blocks[tag.Number]
	}
}
func (s *State) GetBlockByHash(hash [32]byte) *seleneCommon.Block {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if number, exists := s.hashes[hash]; exists {
		return s.blocks[number]
	}
	return nil
}
func (s *State) GetTransaction(hash [32]byte) *seleneCommon.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if loc, exists := s.txs[hash]; exists {
		if block, exists := s.blocks[loc.Block]; exists {
			if len(block.Transactions.Full) > loc.Index {
				return &block.Transactions.Full[loc.Index]
			}
		}
	}
	return nil
}
func (s *State) GetTransactionByBlockAndIndex(blockHash [32]byte, index uint64) *seleneCommon.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if number, exists := s.hashes[blockHash]; exists {
		if block, exists := s.blocks[number]; exists {
			if int(index) < len(block.Transactions.Full) {
				return &block.Transactions.Full[index]
			}
		}
	}
	return nil
}
func (s *State) GetStateRoot(tag seleneCommon.BlockTag) *[32]byte {
	if block := s.GetBlock(tag); block != nil {
		return &block.StateRoot
	}
	return nil
}
func (s *State) GetReceiptsRoot(tag seleneCommon.BlockTag) *[32]byte {
	if block := s.GetBlock(tag); block != nil {
		return &block.ReceiptsRoot
	}
	return nil
}
func (s *State) GetBaseFee(tag seleneCommon.BlockTag) *uint256.Int {
	if block := s.GetBlock(tag); block != nil {
		return &block.BaseFeePerGas
	}
	return nil
}
func (s *State) GetCoinbase(tag seleneCommon.BlockTag) *seleneCommon.Address {
	if block := s.GetBlock(tag); block != nil {
		return &block.Miner
	}
	return nil
}
func (s *State) LatestBlockNumber() *uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latestNumber uint64
	for number := range s.blocks {
		if number > latestNumber {
			latestNumber = number
		}
	}
	if latestNumber > 0 {
		return &latestNumber
	}
	return nil
}
func (s *State) OldestBlockNumber() *uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var oldestNumber uint64 = ^uint64(0)
	for number := range s.blocks {
		if number < oldestNumber {
			oldestNumber = number
		}
	}
	if oldestNumber < ^uint64(0) {
		return &oldestNumber
	}
	return nil
}

//////////////////////////////////////////////
///////////// errors.go //////////////////////
//////////////////////////////////////////////

// package execution

// import (
// 	"fmt"

// 	"github.com/BlocSoc-iitr/selene/consensus/consensus_core"
// 	"github.com/BlocSoc-iitr/selene/consensus/types"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// )

// ExecutionError represents various execution-related errors
type ExecutionError struct {
	Kind    string
	Details interface{}
}

func (e *ExecutionError) Error() string {
	switch e.Kind {
	case "InvalidAccountProof":
		return fmt.Sprintf("invalid account proof for string: %v", e.Details)
	case "InvalidStorageProof":
		details := e.Details.([]interface{})
		return fmt.Sprintf("invalid storage proof for string: %v, slot: %v", details[0], details[1])
	case "CodeHashMismatch":
		details := e.Details.([]interface{})
		return fmt.Sprintf("code hash mismatch for string: %v, found: %v, expected: %v", details[0], details[1], details[2])
	case "ReceiptRootMismatch":
		return fmt.Sprintf("receipt root mismatch for tx: %v", e.Details)
	case "MissingTransaction":
		return fmt.Sprintf("missing transaction for tx: %v", e.Details)
	case "NoReceiptForTransaction":
		return fmt.Sprintf("could not prove receipt for tx: %v", e.Details)
	case "MissingLog":
		details := e.Details.([]interface{})
		return fmt.Sprintf("missing log for transaction: %v, index: %v", details[0], details[1])
	case "TooManyLogsToProve":
		details := e.Details.([]interface{})
		return fmt.Sprintf("too many logs to prove: %v, current limit is: %v", details[0], details[1])
	case "IncorrectRpcNetwork":
		return "execution RPC is for the incorrect network"
	case "InvalidBaseGasFee":
		details := e.Details.([]interface{})
		return fmt.Sprintf("Invalid base gas fee selene %v vs rpc endpoint %v at block %v", details[0], details[1], details[2])
	case "InvalidGasUsedRatio":
		details := e.Details.([]interface{})
		return fmt.Sprintf("Invalid gas used ratio of selene %v vs rpc endpoint %v at block %v", details[0], details[1], details[2])
	case "BlockNotFoundError":
		return fmt.Sprintf("Block %v not found", e.Details)
	case "EmptyExecutionPayload":
		return "Selene Execution Payload is empty"
	case "InvalidBlockRange":
		details := e.Details.([]interface{})
		return fmt.Sprintf("User query for block %v but selene oldest block is %v", details[0], details[1])
	default:
		return "unknown execution error"
	}
}

// Helper functions to create specific ExecutionError instances
func NewInvalidAccountProofError(address consensusTypes.Address) error {
	return &ExecutionError{"InvalidAccountProof", address}
}

func NewInvalidStorageProofError(address consensusTypes.Address, slot consensus_core.Bytes32) error {
	return &ExecutionError{"InvalidStorageProof", []interface{}{address, slot}}
}

func NewCodeHashMismatchError(address consensusTypes.Address, found consensus_core.Bytes32, expected consensus_core.Bytes32) error {
	return &ExecutionError{"CodeHashMismatch", []interface{}{address, found, expected}}
}

func NewReceiptRootMismatchError(tx consensus_core.Bytes32) error {
	return &ExecutionError{"ReceiptRootMismatch", tx}
}

func NewMissingTransactionError(tx consensus_core.Bytes32) error {
	return &ExecutionError{"MissingTransaction", tx}
}

func NewNoReceiptForTransactionError(tx consensus_core.Bytes32) error {
	return &ExecutionError{"NoReceiptForTransaction", tx}
}

func NewMissingLogError(tx consensus_core.Bytes32, index uint64) error {
	return &ExecutionError{"MissingLog", []interface{}{tx, index}}
}

func NewTooManyLogsToProveError(count int, limit int) error {
	return &ExecutionError{"TooManyLogsToProve", []interface{}{count, limit}}
}

func NewIncorrectRpcNetworkError() error {
	return &ExecutionError{"IncorrectRpcNetwork", nil}
}

func NewInvalidBaseGasFeeError(selene uint64, rpc uint64, block uint64) error {
	return &ExecutionError{"InvalidBaseGasFee", []interface{}{selene, rpc, block}}
}

func NewInvalidGasUsedRatioError(seleneRatio float64, rpcRatio float64, block uint64) error {
	return &ExecutionError{"InvalidGasUsedRatio", []interface{}{seleneRatio, rpcRatio, block}}
}

func NewBlockNotFoundError(block uint64) error {
	return &ExecutionError{"BlockNotFoundError", block}
}

func NewEmptyExecutionPayloadError() error {
	return &ExecutionError{"EmptyExecutionPayload", nil}
}

func NewInvalidBlockRangeError(queryBlock uint64, oldestBlock uint64) error {
	return &ExecutionError{"InvalidBlockRange", []interface{}{queryBlock, oldestBlock}}
}

// EvmError represents EVM-related errors
type EvmError struct {
	Kind    string
	Details interface{}
}

func (e *EvmError) Error() string {
	switch e.Kind {
	case "Revert":
		return fmt.Sprintf("execution reverted: %v", e.Details)
	case "Generic":
		return fmt.Sprintf("evm error: %v", e.Details)
	case "RpcError":
		return fmt.Sprintf("rpc error: %v", e.Details)
	default:
		return "unknown evm error"
	}
}

// Helper functions for creating specific EVM errors
func NewRevertError(data []byte) error {
	return &EvmError{"Revert", data}
}

func NewGenericError(message string) error {
	return &EvmError{"Generic", message}
}

func NewRpcError(report error) error {
	return &EvmError{"RpcError", report}
}

func DecodeRevertReason(data []byte) string {
	reason, err := abi.UnpackRevert(data)
	if err != nil {
		reason = string(err.Error())
	}
	return reason
}

////////////////////////////////////
////////// types.go ////////////////
////////////////////////////////////

// package execution

// import (
//
//	"encoding/json"
//	"fmt"
//	seleneCommon "github.com/BlocSoc-iitr/selene/common"
//	"github.com/BlocSoc-iitr/selene/utils"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/common/hexutil"
//	"github.com/holiman/uint256"
//	"math/big"
//	"reflect"
//
// )
type AccessListItem struct {
	Address     common.Address `json:"address"`
	StorageKeys []B256         `json:"storageKeys"`
}
type AccessList struct {
	AccessList []AccessListItem `json:"accessList"`
	GasUsed    hexutil.Bytes    `json:"gasUsed"`
}
type FeeHistory struct {
	BaseFeePerGas []hexutil.Big
	GasUsedRatio  []float64
	OldestBlock   *hexutil.Big
	Reward        [][]hexutil.Big
}

// defined storage proof	and EIP1186ProofResponse structs
type StorageProof struct {
	Key   common.Hash     `json:"key"`
	Proof []hexutil.Bytes `json:"proof"`
	Value *uint256.Int    `json:"value"`
}
type EIP1186ProofResponse struct {
	Address      common.Address  `json:"address"`
	Balance      *uint256.Int    `json:"balance"`
	CodeHash     common.Hash     `json:"codeHash"`
	Nonce        hexutil.Uint64  `json:"nonce"`
	StorageHash  common.Hash     `json:"storageHash"`
	AccountProof []hexutil.Bytes `json:"accountProof"`
	StorageProof []StorageProof  `json:"storageProof"`
}
type Account struct {
	Balance     *big.Int
	Nonce       uint64
	CodeHash    common.Hash
	Code        []byte
	StorageHash common.Hash
	Slots       []Slot
}

// This is to help in unmarshaling values from rpc response
type Slot struct {
	Key   common.Hash // The key (slot)
	Value *big.Int    // The value (storage value)
}
type CallOpts struct {
	From     *seleneCommon.Address `json:"from,omitempty"`
	To       *seleneCommon.Address `json:"to,omitempty"`
	Gas      *big.Int              `json:"gas,omitempty"`
	GasPrice *big.Int              `json:"gasPrice,omitempty"`
	Value    *big.Int              `json:"value,omitempty"`
	Data     []byte                `json:"data,omitempty"`
}

func (c *CallOpts) String() string {
	return fmt.Sprintf("CallOpts{From: %v, To: %v, Gas: %v, GasPrice: %v, Value: %v, Data: 0x%x}",
		c.From, c.To, c.Gas, c.GasPrice, c.Value, c.Data)
}

func (c *CallOpts) Serialize() ([]byte, error) {
	serialized := make(map[string]interface{})
	v := reflect.ValueOf(*c)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		if !field.IsNil() {
			var value interface{}
			var err error

			switch field.Interface().(type) {
			case *common.Address:
				value = utils.Address_to_hex_string(*field.Interface().(*common.Address))
			case *big.Int:
				value = utils.U64_to_hex_string(field.Interface().(*big.Int).Uint64())
			case []byte:
				value, err = utils.Bytes_serialize(field.Interface().([]byte))
				if err != nil {
					return nil, fmt.Errorf("error serializing %s: %w", fieldName, err)
				}
			default:
				return nil, fmt.Errorf("unsupported type for field %s", fieldName)
			}

			serialized[fieldName] = value
		}
	}

	return json.Marshal(serialized)
}

func (c *CallOpts) Deserialize(data []byte) error {
	var serialized map[string]string
	if err := json.Unmarshal(data, &serialized); err != nil {
		return err
	}

	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		if value, ok := serialized[fieldName]; ok {
			switch field.Interface().(type) {
			case *common.Address:
				addressBytes, err := utils.Hex_str_to_bytes(value)
				if err != nil {
					return fmt.Errorf("error deserializing %s: %w", fieldName, err)
				}
				addr := common.BytesToAddress(addressBytes)
				field.Set(reflect.ValueOf(&addr))
			case *big.Int:
				intBytes, err := utils.Hex_str_to_bytes(value)
				if err != nil {
					return fmt.Errorf("error deserializing %s: %w", fieldName, err)
				}
				bigInt := new(big.Int).SetBytes(intBytes)
				field.Set(reflect.ValueOf(bigInt))
			case []byte:
				byteValue, err := utils.Bytes_deserialize([]byte(value))
				if err != nil {
					return fmt.Errorf("error deserializing %s: %w", fieldName, err)
				}
				field.SetBytes(byteValue)
			default:
				return fmt.Errorf("unsupported type for field %s", fieldName)
			}
		}
	}

	return nil
}

const PARALLEL_QUERY_BATCH_SIZE = 20
