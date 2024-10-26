package execution

import (
	"log"
	"math/big"
	"sync"

	Common "github.com/ethereum/go-ethereum/common" //geth common imported as Common
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/x-senpai-x/PracticeSelene/common"
	"github.com/x-senpai-x/PracticeSelene/execution/evm"
	"github.com/x-senpai-x/PracticeSelene/execution/logging"
	"go.uber.org/zap"
)
type BlockTag = common.BlockTag
type U256 = *big.Int
type B256 = Common.Hash
type Address = common.Address
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
func (e *Evm) Call(opts *CallOpts) ([]byte, evmError) {
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
func (e *Evm) EstimateGas(opts *CallOpts) (uint64, error) {
	tx, err := e.callInner(opts)
	if err != nil {
		return 0, err
	}

	switch tx.Result.Type {
	case "Success":
		return tx.Result.GasUsed, nil
	case "Revert":
		return tx.Result.GasUsed,nil
	case "Halt":
		return tx.Result.GasUsed, nil
	default:
		return 0, fmt.Errorf("unexpected execution result")
	}
}
func (e *Evm) callInner(opts *CallOpts) (*evm.ExecutionResult, error) {
	db, err := NewProofDB( e.tag, e.execution)
	if err != nil {
		return nil, err
	}
	if err := db.State.PrefetchState(opts); err != nil {
		return nil, err
	}
	env := e.getEnv(opts, e.tag)
	evm:=evm.NewEvmBuilder().WithDB(db).WithEnv(&env).Build()
	//ctx:=evm.IntoContextWithHandlerCfg()
// ctx is a ContextWithHandlerCfg instance created from evm using IntoContextWithHandlerCfg.
	ctx := evm.IntoContextWithHandlerCfg[interface{}, ProofDB](evm)
	//ctx:=evm.IntoContextWithHandlerCfg[interface{}, ProofDB](evm)
	for {
		db:=ctx.Context.Evm.Inner.DB
		if db.
		if db.State.NeedsUpdate() {
			if err := db.State.UpdateState(); err != nil {
				return nil, err
			}
		}
		// Create a new types.Transaction
		tx := types.NewTransaction(
			0, // Nonce is not provided in CallOpts, so we use 0
			*opts.To,
			opts.Value,
			opts.Gas.Uint64(),
			opts.GasPrice,
			opts.Data,
		)

		// Create a Message from the transaction
		signer := types.NewEIP155Signer(chainConfig.ChainID)
		msg, err := core.TransactionToMessage(tx, signer, blockContext.BaseFee)
		if err != nil {
			return nil, err
		}

		// Override the From address with the one provided in CallOpts
		msg.From = *opts.From

		// Create a new EVM and apply the message
		result, err := core.ApplyMessage(evm, msg, new(core.GasPool).AddGas(opts.Gas.Uint64()))
		if err != nil {
			return nil, err
		}

		if result.Err == nil {
			return &ExecutionResult{
				Success: true,
				Output:  result.ReturnData,
				GasUsed: result.UsedGas,
			}, nil
		}

		if _, ok := result.Err.(vm.ErrExecutionReverted); ok {
			return &ExecutionResult{
				Success:  false,
				Output:   result.ReturnData,
				GasUsed:  result.UsedGas,
				Reverted: true,
			}, nil
		}

		// If the error is not a revert, we break the loop and return the error
		return nil, result.Err
	}
}
func (e *Evm) getEnv( opts *CallOpts, tag BlockTag) evm.Env {
	env:=evm.NewEnv()//needs to be implemented
	//env.Tx.Transact_to=evm.TransactTo:Call
	env.Tx.Caller=*opts.From //is conversion required?
	env.Tx.Value=opts.Value
	env.Tx.Data=opts.Data
	env.Tx.Gas_limit=opts.Gas.Uint64()
	env.Tx.Gas_price=opts.GasPrice
	block, err := e.execution.GetBlock(tag, false)
	if err != nil {
        // Handle the error appropriately, e.g., log it or set default values
        log.Printf("Error getting block: %v", err)
        // Return a default or empty env in case of error
        return evm.Env{}
    }
	env.Block.Number=block.Number
	env.Block.Coinbase=block.Miner
	env.Block.Timestamp=block.Timestamp
	env.Block.Difficulty=block.Difficulty
	env.Cfg.ChainID=e.chainID
	
	return env
}
type ProofDB struct {
	State *EvmState
}

func NewProofDB( tag BlockTag, execution *ExecutionClient) (*ProofDB, error) {
	state := NewEvmState(execution, tag)
	return &ProofDB{
		State: state,
	}, nil
}

type StateAccess struct {
	Basic     *Address
	BlockHash *uint64
	Storage   map[Address]U256
}
type AccountInfo struct {
	Balance  U256
	Nonce    uint64
	CodeHash B256
	Code     hexutil.Bytes //Doubtful
}

func NewAccountInfo(balance U256, nonce uint64, codeHash B256, code hexutil.Bytes) AccountInfo {
	return AccountInfo{
		Balance:  balance,
		Nonce:    nonce,
		CodeHash: codeHash,
		Code:     code,
	}
}

type EvmState struct {
	Basic      map[Address]evm.AccountInfo
	BlockHash  map[uint64]B256
	Storage    map[Address]map[U256]U256
	Block      BlockTag
	Access     *StateAccess
	Execution  *ExecutionClient
	AccessList map[Address]struct{}
} //added just now : UPDATE
func NewEvmState(execution *ExecutionClient, block BlockTag) *EvmState {
	return &EvmState{
		Basic:      make(map[Address]evm.AccountInfo),
		BlockHash:  make(map[uint64]B256),
		Storage:    make(map[Address]map[U256]U256),
		Block:      block,
		Access:     nil,
		Execution:  execution,
		AccessList: make(map[Address]struct{}), //added just now : UPDATE
	}
}
func (e *EvmState) UpdateState() error {
	if e.Access == nil {
		return nil
	}
	access := e.Access
	e.Access = nil // Equivalent to Rust's self.access.take()
	switch {
	case access.Basic != nil:
		account, err := e.Execution.GetAccount( access.Basic, nil, e.Block)
		if err != nil {
			return err
		}
		bytecode := NewBytecodeRaw(account.Code)
		codeHash := B256FromSlice(account.CodeHash[:])
		balance := ConvertU256(account.Balance)
		accountInfo := evm.AccountInfo{
			Balance:  balance,
			Nonce:    account.Nonce,
			CodeHash: codeHash,
			Code:     bytecode,
		}
		e.Basic[*access.Basic] = accountInfo
	case access.Storage != nil:
		for address, slotValue := range access.Storage {
			slot := Common.BigToHash(slotValue) // Use slotValue directly
			slots := []B256{slot}
			account, err := e.Execution.GetAccount(&address, slots, e.Block)
			if err != nil {
				return err
			}
			storage, ok := e.Storage[address]
			if !ok {
				storage = make(map[U256]U256) // Initialize with *big.Int
				e.Storage[address] = storage
			}
			value, ok := account.Slots[slot]
			if !ok {
				return errors.New("slot not found in account")
			}
			storage[slotValue] = value
		}
	case access.BlockHash != nil:
		block, err := e.Execution.GetBlock(BlockTag{Number: *access.BlockHash}, false)
		if err != nil {
			return err
		}
		hash := B256FromSlice(block.Hash[:]) // Convert [32]byte to []byte
		e.BlockHash[*access.BlockHash] = hash
	default:
		return errors.New("invalid access type")
	}
	return nil
}
func (e *EvmState) NeedsUpdate() bool {
	return e.Access != nil //Checks if access Field is non zero
}
func (e *EvmState) GetBasic(address Address) (evm.AccountInfo, error) {
	if account, exists := e.Basic[address]; exists {
		return account, nil
	} else {
		e.Access = &StateAccess{Basic: &address}
		return AccountInfo{}, errors.New("state missing")
	}
}
func (e *EvmState) GetStorage(address Address, slot U256) (U256, error) {
	storage := e.Storage[address]
	if value, exists := storage[slot]; exists {
		return value, nil
	} else {
		e.Access = &StateAccess{Storage: map[Address]U256{address: slot}}
		return &big.Int{}, errors.New("state missing") // Return an empty U256 and the error
	}
}
func (e *EvmState) GetBlockHash(block uint64) (B256, error) {
	if hash, exists := e.BlockHash[block]; exists {
		return hash, nil
	} else {
		e.Access = &StateAccess{BlockHash: &block}
		return B256{}, errors.New("state missing")
	}
}
func (e *EvmState) PrefetchState( opts *CallOpts) error {
	list, err := e.Execution.Rpc.CreateAccessList(opts, e.Block)
	if err != nil {
		return err
	}
	fromAccessEntry := AccessListItem{
		Address:     *opts.From,
		StorageKeys: []B256{},
	}
	toAccessEntry := AccessListItem{
		Address:     *opts.To,
		StorageKeys: []B256{},
	}
	coinbase, err := e.Execution.GetBlock( e.Block, false)
	if err != nil {
		return err
	}
	producerAccessEntry := AccessListItem{
		Address:     coinbase.Miner,
		StorageKeys: []B256{},
	}
	listAddresses := make(map[Address]bool)
	for _, item := range list {
		listAddresses[item.Address] = true
	}

	if !listAddresses[fromAccessEntry.Address] {
		list = append(list, fromAccessEntry)
	}
	if !listAddresses[toAccessEntry.Address] {
		list = append(list, toAccessEntry)
	}
	if !listAddresses[producerAccessEntry.Address] {
		list = append(list, producerAccessEntry)
	}
	accountMap := make(map[Address]Account)
	var wg sync.WaitGroup
	sem := make(chan struct{}, PARALLEL_QUERY_BATCH_SIZE)
	var mu sync.Mutex
	for _, account := range list {
		wg.Add(1)
		sem <- struct{}{}
		go func(account AccessListItem) {
			defer wg.Done()
			defer func() { <-sem }()

			acc, err := e.Execution.GetAccount(&account.Address, account.StorageKeys, e.Block)
			if err == nil {
				mu.Lock()
				accountMap[account.Address] = *acc
				mu.Unlock()
			}
		}(account)
	}
	wg.Wait()
	for address, account := range accountMap {
		bytecode := NewBytecodeRaw(account.Code)
		codeHash := Common.BytesToHash(account.CodeHash[:])
		balance := ConvertU256(account.Balance)
		info := NewAccountInfo(balance, account.Nonce, codeHash, bytecode)
		e.Basic[address] = info
		for slot, value := range account.Slots {
			slotHash := B256FromSlice(slot[:])
			valueU256 := ConvertU256(value)
			storage, exists := e.Storage[address]
			if !exists {
				storage = make(map[U256]U256)
				e.Storage[address] = storage
			}
			slotBigInt := new(big.Int).SetBytes(slotHash.Bytes())
			storage[slotBigInt] = valueU256
		}
	}
	return nil
}

type AccessListItem struct {
	Address     Address //I used Common here instead of common
	StorageKeys []B256
}
type Bytecode []byte

func NewBytecodeRaw(code []byte) hexutil.Bytes {
	return hexutil.Bytes(code)
}
func B256FromSlice(slice []byte) Common.Hash {
	return Common.BytesToHash(slice)
}
func ConvertU256(value *big.Int) *big.Int {
	valueSlice := make([]byte, 32)
	value.FillBytes(valueSlice)
	result := new(big.Int).SetBytes(valueSlice)
	return result
}

type Database interface {
	Basic(address Address) (AccountInfo, error)
	BlockHash(number uint64) (B256, error)
	Storage(address Address, slot *big.Int) (*big.Int, error)
	CodeByHash(codeHash B256) (Bytecode, error)
}

func (db *ProofDB) Basic(address Address) (evm.AccountInfo, error) {
	if isPrecompile(address) {
		return evm.AccountInfo{}, nil // Return a default AccountInfo
	}
	logging.Trace("fetch basic evm state for address", zap.String("address", address.Hex()))
	return db.State.GetBasic(address)
}
func (db *ProofDB) BlockHash(number uint64) (B256, error) {
	logging.Trace("fetch block hash for block number", zap.Uint64("number", number))
	return db.State.GetBlockHash(number)
}
func (db *ProofDB) Storage(address Address, slot *big.Int) (*big.Int, error) {
	logging.Trace("fetch storage for address and slot",
		zap.String("address", address.Hex()),
		zap.String("slot", slot.String()))
	return db.State.GetStorage(address, slot)
}
func (db *ProofDB) CodeByHash(_ B256) (evm.Bytecode, error) {
    return evm.Bytecode{}, errors.New("should never be called")
}

func isPrecompile(address Address) bool {
	precompileAddress := Address{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09}
	return address.Cmp(precompileAddress) <= 0 && address.Cmp(Address{}) > 0
}


//Skipped the testing for now
//Proposal: We should be using geth Address instead of locally defined address in common/types.go in the entire codebase
/*package execution

import (
	"fmt"
	//"context"
	"encoding/hex"
	//"log"
	"math/big"
	Common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
//"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	//"github.com/ethereum/go-ethereum/ethdb"

	//"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/params"
	//"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/triedb"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/18aaddy/selene-practics/common"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/crypto"

)
type B256 = Common.Hash
type U256 = big.Int
type HeaderReader interface {
	GetHeader(hash B256, number uint64) *types.Header
}
type Evm struct {
	execution *ExecutionClient
	chainID   uint64
	tag       common.BlockTag
}
func NewEvm(execution *ExecutionClient, chainID uint64, tag common.BlockTag) *Evm {
	return &Evm{
		execution: execution,
		chainID:   chainID,
		tag:       tag,
	}
}
func (e *Evm) CallInner(opts *CallOpts) (*core.ExecutionResult, error) {
	txContext := vm.TxContext{
		Origin:   *opts.From,
		GasPrice: opts.GasPrice,
	}
	tag:= e.tag
	block, err := e.execution.GetBlock(tag, false)
	if err != nil {
		return nil, err
	}
	blockContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash: func(n uint64) B256 {
			return B256{} // You might want to implement this properly
		},
		Coinbase:    block.Miner.Addr,
		BlockNumber: new(U256).SetUint64(block.Number),
		Time:        block.Timestamp,
		Difficulty:  block.Difficulty.ToBig(),
		GasLimit:    block.GasLimit,
		BaseFee:     block.BaseFeePerGas.ToBig(),
	}
	db:= rawdb.NewMemoryDatabase()
	tdb:= triedb.NewDatabase(db, nil)
	sdb:= state.NewDatabase(tdb, nil)
	//root:= trie.NewSecure(common.Hash{}, trie.NewDatabase(sdb))
	state, err := state.New(types.EmptyRootHash, sdb)
	//witness:=stateless.NewWitness(block,)
	//state.StartPrefetcher("hello",witness)
	// Create a new vm object
	var chainConfig *params.ChainConfig
	chainID:=e.chainID
	switch (int64(chainID)) {
		case MainnetID:
			chainConfig = params.MainnetChainConfig
		case HoleskyID:
			chainConfig = params.HoleskyChainConfig
		case SepoliaID:
			chainConfig = params.SepoliaChainConfig
		case LocalDevID:
			chainConfig = params.AllEthashProtocolChanges
		default:
			// Handle unknown chain ID
			chainConfig = nil
		}
		//Note other chainids not implemented(local testing)
		//	"github.com/ethereum/go-ethereum/params"

	config:= vm.Config{}
	nonceBytes, err := hex.DecodeString(block.Nonce)
	var nonce types.BlockNonce
	copy(nonce[:], nonceBytes)
	//Prefetch database: 
	var witness *stateless.Witness
	//need uncle hash for context so manually creatuing it
	header := &types.Header{
		ParentHash: 		  block.ParentHash,
		UncleHash: 			  block.Sha3Uncles,
		Coinbase: 			  block.Miner.Addr,
		Root: 				  block.StateRoot,
		TxHash: 			  block.TransactionsRoot,
		ReceiptHash: 		  block.ReceiptsRoot,
		Bloom: 				  types.Bloom(block.LogsBloom),
		Difficulty: 		  new(U256).SetUint64(block.Difficulty.Uint64()),
		Number: 			  new(U256).SetUint64(block.Number),
		GasLimit: 			  block.GasLimit,
		GasUsed: 			  block.GasUsed,
		Time: 				  block.Timestamp,
		Extra: 				  block.ExtraData,
		MixDigest: 			  block.MixHash,
		Nonce: 				  nonce,
		BaseFee: 			  new(U256).SetUint64(block.BaseFeePerGas.Uint64()),
		//WithdrawalsHash: 	  block.WithdrawalsRoot,
		BlobGasUsed: 		  block.BlobGasUsed,
		ExcessBlobGas: 		  block.ExcessBlobGas,
		//ParentBeaconBlockRoot: block.ParentBeaconBlockRoot,
		//RequestsHash: 		  block.RequestsRoot,
	}
	var(
	key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	addr   = crypto.PubkeyToAddress(key.PublicKey)
)
	genspec := &core.Genesis{
		Config:    params.AllCliqueProtocolChanges,
		Alloc: map[Common.Address]types.Account{
			addr: {Balance: big.NewInt(10000000000000000)},
		},
		BaseFee: big.NewInt(params.InitialBaseFee),
	}//using base fees as same as eip1559 blocks
	var engine = clique.New(params.AllCliqueProtocolChanges.Clique, db)
	//WithdrawalsHash,ParentBeaconBlockRoot,RequestsHash not found in block struct
	chain,_:=core.NewBlockChain(db, nil, genspec, nil,engine,config,nil)
	//don't know whether to use sdb or db
	//doubt in implementtation of genspec
	witness,err = stateless.NewWitness(header,chain)
	state.StartPrefetcher("evm", witness)
	evm := vm.NewEVM(blockContext,txContext,state,chainConfig,config)
// Prepare the call message
	msg := core.Message{
		From:              *opts.From,
		To:                opts.To,
		Value:             opts.Value,
		GasLimit:          opts.Gas.Uint64(),
		GasPrice:          opts.GasPrice,
		GasFeeCap:         nil, // Set if using EIP-1559
		GasTipCap:         nil, // Set if using EIP-1559
		Data:              opts.Data,
		AccessList:        nil, // Set if using EIP-2930
		SkipNonceChecks: false,
	}
	// Execute the call
	result, err := core.ApplyMessage(evm, &msg, new(core.GasPool).AddGas(opts.Gas.Uint64()))
	if err != nil {
		return nil, fmt.Errorf("failed to apply message: %w", err)
	}

	return result, nil
}
func (e *Evm) Call(opts *CallOpts) ([]byte, error) {
	result, err := e.CallInner(opts)
	if err != nil {
		return nil, fmt.Errorf("call failed: %w", err)
	}

	switch {
	case result.Failed():
		return nil, &EvmError{Kind: "execution reverted", Details: result.Revert()}
	default:
		return result.Return(), nil
	}
}
func (e *Evm) EstimateGas(opts *CallOpts) (uint64, error) {
	result, err := e.CallInner(opts)
	if err != nil {
		return 0, fmt.Errorf("gas estimation failed: %w", err)
	}

	return result.UsedGas, nil
}*/