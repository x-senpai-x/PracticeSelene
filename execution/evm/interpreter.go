package evm

import (
	"encoding/json"
	"fmt"
	"reflect"
)
type Interpreter struct {
	InstructionPointer *byte // Using *byte as equivalent to *const u8 in Rust
	Gas               Gas
	Contract          Contract
	InstructionResult InstructionResult
	Bytecode          Bytes
	IsEof             bool
	// Is init flag for eof create
	IsEofInit         bool
	// Shared memory.
	// Note: This field is only set while running the interpreter loop.
	// Otherwise, it is taken and replaced with empty shared memory.
	SharedMemory      SharedMemory
	// Stack.
	Stack             Stack
	// EOF function stack.
	FunctionStack     FunctionStack
	// The return data buffer for internal calls.
	// It has multi usage:
	// - It contains the output bytes of call sub call.
	// - When this interpreter finishes execution it contains the output bytes of this contract.
	ReturnDataBuffer  Bytes
	// Whether the interpreter is in "staticcall" mode, meaning no state changes can happen.
	IsStatic          bool
	// Actions that the EVM should do.
	// Set inside CALL or CREATE instructions and RETURN or REVERT instructions.
	// Additionally, those instructions will set InstructionResult to CallOrCreate/Return/Revert so we know the reason.
	NextAction        InterpreterAction
}
type Gas struct {
	Limit   uint64
	Remaining uint64
	Refunded int64
}
type Contract struct{
	Input Bytes
	Bytecode Bytecode
	Hash *B256
	TargetAddress Address
	BytecodeAddress *Address
	Caller Address
	CallValue U256
}
// InstructionResult represents the result of an instruction in the EVM.
type InstructionResult uint8

const (
	// Success codes
	Continue InstructionResult = iota // 0x00
	Stop
	Return
	SelfDestruct
	ReturnContract

	// Revert codes
	Revert                              = 0x10 // revert opcode
	CallTooDeep
	OutOfFunds
	CreateInitCodeStartingEF00
	InvalidEOFInitCode
	InvalidExtDelegateCallTarget

	// Actions
	CallOrCreate = 0x20

	// Error codes
	OutOfGas                         = 0x50
	MemoryOOG
	MemoryLimitOOG
	PrecompileOOG
	InvalidOperandOOG
	OpcodeNotFound
	CallNotAllowedInsideStatic
	StateChangeDuringStaticCall
	InvalidFEOpcode
	InvalidJump
	NotActivated
	StackUnderflow
	StackOverflow
	OutOfOffset
	CreateCollision
	OverflowPayment
	PrecompileError
	NonceOverflow
	CreateContractSizeLimit
	CreateContractStartingWithEF
	CreateInitCodeSizeLimit
	FatalExternalError
	ReturnContractInNotInitEOF
	EOFOpcodeDisabledInLegacy
	EOFFunctionStackOverflow
	EofAuxDataOverflow
	EofAuxDataTooSmall
	InvalidEXTCALLTarget
)

// MarshalJSON customizes the JSON serialization of InstructionResult.
func (ir InstructionResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint8(ir))
}

// UnmarshalJSON customizes the JSON deserialization of InstructionResult.
func (ir *InstructionResult) UnmarshalJSON(data []byte) error {
	var value uint8
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*ir = InstructionResult(value)
	return nil
}
type SharedMemory struct {
	Buffer []byte
	Checkpoints []int
	LastCheckpoint int
	MemoryLimit *uint64 //although implemented as feature memory_limit but kept p=optional for now 
}
/*
pub struct SharedMemory {
    /// The underlying buffer.
    buffer: Vec<u8>,
    /// Memory checkpoints for each depth.
    /// Invariant: these are always in bounds of `data`.
    checkpoints: Vec<usize>,
    /// Invariant: equals `self.checkpoints.last()`
    last_checkpoint: usize,
    /// Memory limit. See [`CfgEnv`](revm_primitives::CfgEnv).
    #[cfg(feature = "memory_limit")]
    memory_limit: u64,
}*/
type Stack struct {
	Data []U256
}
type FunctionStack struct {
    ReturnStack      []FunctionReturnFrame `json:"return_stack"`
    CurrentCodeIdx   uint64                `json:"current_code_idx"` // Changed to uint64 to match Rust's usize
}
// MarshalJSON customizes the JSON serialization of FunctionStack.
func (fs FunctionStack) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ReturnStack    []FunctionReturnFrame `json:"return_stack"`
        CurrentCodeIdx uint64                `json:"current_code_idx"`
    }{
        ReturnStack:    fs.ReturnStack,
        CurrentCodeIdx: fs.CurrentCodeIdx,
    })
}

// UnmarshalJSON customizes the JSON deserialization of FunctionStack.
func (fs *FunctionStack) UnmarshalJSON(data []byte) error {
    var temp struct {
        ReturnStack    []FunctionReturnFrame `json:"return_stack"`
        CurrentCodeIdx uint64                `json:"current_code_idx"`
    }
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }
    fs.ReturnStack = temp.ReturnStack
    fs.CurrentCodeIdx = temp.CurrentCodeIdx
    return nil
}
// FunctionReturnFrame represents a frame in the function return stack for the EVM interpreter.
type FunctionReturnFrame struct {
    // The index of the code container that this frame is executing.
    Idx uint64 `json:"idx"` // Changed to uint64 to match Rust's usize
    // The program counter where frame execution should continue.
    PC uint64 `json:"pc"` // Changed to uint64 to match Rust's usize
}
// MarshalJSON customizes the JSON serialization of FunctionReturnFrame.
func (frf FunctionReturnFrame) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        Idx uint64 `json:"idx"`
        PC  uint64 `json:"pc"`
    }{
        Idx: frf.Idx,
        PC:  frf.PC,
    })
}
// UnmarshalJSON customizes the JSON deserialization of FunctionReturnFrame.
func (frf *FunctionReturnFrame) UnmarshalJSON(data []byte) error {
    var temp struct {
        Idx uint64 `json:"idx"`
        PC  uint64 `json:"pc"`
    }
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }
    frf.Idx = temp.Idx
    frf.PC = temp.PC
    return nil
}
type InterpreterAction struct {
    // We use a type field to distinguish between different actions
    actionType ActionType 
    callInputs     *CallInputs
    createInputs   *CreateInputs
    eofCreateInputs *EOFCreateInputs
    result         *InterpreterResult
}
type InterpreterResult struct {
	// The result of the execution.
	Result InstructionResult
	Output Bytes
	Gas Gas
}
type jsonInterpreterAction struct {
    Type         string          `json:"type"`
    CallInputs   *CallInputs     `json:"call_inputs,omitempty"`
    CreateInputs *CreateInputs   `json:"create_inputs,omitempty"`
    EOFInputs    *EOFCreateInputs      `json:"eof_inputs,omitempty"`
    Result       *InterpreterResult         `json:"result,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface
func (a *InterpreterAction) MarshalJSON() ([]byte, error) {
    var jsonAction jsonInterpreterAction

    switch a.actionType {
    case ActionTypeCall:
        jsonAction.Type = "Call"
        jsonAction.CallInputs = a.callInputs
    case ActionTypeCreate:
        jsonAction.Type = "Create"
        jsonAction.CreateInputs = a.createInputs
    case ActionTypeEOFCreate:
        jsonAction.Type = "EOFCreate"
        jsonAction.EOFInputs = a.eofCreateInputs
    case ActionTypeReturn:
        jsonAction.Type = "Return"
        jsonAction.Result = a.result
    case ActionTypeNone:
        jsonAction.Type = "None"
    default:
        return nil, fmt.Errorf("unknown action type: %v", a.actionType)
    }

    return json.Marshal(jsonAction)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (a *InterpreterAction) UnmarshalJSON(data []byte) error {
    var jsonAction jsonInterpreterAction
    if err := json.Unmarshal(data, &jsonAction); err != nil {
        return err
    }

    switch jsonAction.Type {
    case "Call":
        if jsonAction.CallInputs == nil {
            return fmt.Errorf("Call action missing inputs")
        }
        *a = *NewCallAction(jsonAction.CallInputs)
    case "Create":
        if jsonAction.CreateInputs == nil {
            return fmt.Errorf("Create action missing inputs")
        }
        *a = *NewCreateAction(jsonAction.CreateInputs)
    case "EOFCreate":
        if jsonAction.EOFInputs == nil {
            return fmt.Errorf("EOFCreate action missing inputs")
        }
        *a = *NewEOFCreateAction(jsonAction.EOFInputs)
    case "Return":
        if jsonAction.Result == nil {
            return fmt.Errorf("Return action missing result")
        }
        *a = *NewReturnAction(jsonAction.Result)
    case "None":
        *a = *NewNoneAction()
    default:
        return fmt.Errorf("unknown action type: %s", jsonAction.Type)
    }

    return nil
}

// ActionType represents the type of interpreter action
type ActionType int

const (
    ActionTypeNone ActionType = iota
    ActionTypeCall
    ActionTypeCreate
    ActionTypeEOFCreate
    ActionTypeReturn
)
// NewCallAction creates a new Call action
func NewCallAction(inputs *CallInputs) *InterpreterAction {
    return &InterpreterAction{
        actionType: ActionTypeCall,
        callInputs: inputs,
    }
}

// NewCreateAction creates a new Create action
func NewCreateAction(inputs *CreateInputs) *InterpreterAction {
    return &InterpreterAction{
        actionType: ActionTypeCreate,
        createInputs: inputs,
    }
}

// NewEOFCreateAction creates a new EOFCreate action
func NewEOFCreateAction(inputs *EOFCreateInputs) *InterpreterAction {
    return &InterpreterAction{
        actionType: ActionTypeEOFCreate,
        eofCreateInputs: inputs,
    }
}

// NewReturnAction creates a new Return action
func NewReturnAction(result *InterpreterResult) *InterpreterAction {
    return &InterpreterAction{
        actionType: ActionTypeReturn,
        result: result,
    }
}

// NewNoneAction creates a new None action
func NewNoneAction() *InterpreterAction {
    return &InterpreterAction{
        actionType: ActionTypeNone,
    }
}
/*
// MarshalJSON customizes the JSON serialization of InterpreterAction.
func (ia InterpreterAction) MarshalJSON() ([]byte, error) {
    var action struct {
        ActionType ActionType      `json:"action_type"`
        Inputs     interface{}     `json:"inputs,omitempty"`
        Result     InterpreterResult `json:"result,omitempty"`
    }
    action.ActionType = ia.ActionType
    action.Inputs = ia.Inputs
    action.Result = ia.Result

    return json.Marshal(action)
}

// UnmarshalJSON customizes the JSON deserialization of InterpreterAction.
func (ia *InterpreterAction) UnmarshalJSON(data []byte) error {
    var action struct {
        ActionType ActionType      `json:"action_type"`
        Inputs     interface{}     `json:"inputs,omitempty"`
        Result     InterpreterResult `json:"result,omitempty"`
    }
    if err := json.Unmarshal(data, &action); err != nil {
        return err
    }
    ia.ActionType = action.ActionType
    ia.Inputs = action.Inputs
    ia.Result = action.Result
    return nil
}*/
type Range struct {
    Start int `json:"start"`
    End   int `json:"end"`
}
type CallValue struct {
    ValueType string `json:"value_type"`
    Amount    U256   `json:"amount,omitempty"`
}

// Transfer creates a CallValue representing a transfer.
func Transfer(amount U256) CallValue {
    return CallValue{ValueType: "transfer", Amount: amount}
}

// Apparent creates a CallValue representing an apparent value.
func Apparent(amount U256) CallValue {
    return CallValue{ValueType: "apparent", Amount: amount}
}
type CallScheme int

const (
    Call CallScheme = iota
    CallCode
    DelegateCall
    StaticCall
    ExtCall	
    ExtStaticCall
    ExtDelegateCall
)
type CallInputs struct {
    Input             Bytes       `json:"input"`
    ReturnMemoryOffset Range       `json:"return_memory_offset"`
    GasLimit          uint64      `json:"gas_limit"`
    BytecodeAddress   Address     `json:"bytecode_address"`
    TargetAddress     Address     `json:"target_address"`
    Caller            Address     `json:"caller"`
    Value             CallValue   `json:"value"`
    Scheme            CallScheme  `json:"scheme"`
    IsStatic          bool        `json:"is_static"`
    IsEof            bool        `json:"is_eof"`
}

// JSON serialization for CallValue
func (cv CallValue) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ValueType string `json:"value_type"`
        Amount    *U256  `json:"amount,omitempty"`
    }{
        ValueType: cv.ValueType,
        Amount:    &cv.Amount,
    })
}

// JSON deserialization for CallValue
func (cv *CallValue) UnmarshalJSON(data []byte) error {
    var aux struct {
        ValueType string `json:"value_type"`
        Amount    *U256  `json:"amount,omitempty"`
    }
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    cv.ValueType = aux.ValueType
    if aux.ValueType == "transfer" {
        cv.Amount = *aux.Amount
    } else if aux.ValueType == "apparent" {
        cv.Amount = *aux.Amount
    }
    return nil
}
type CreateInputs struct {
	Caller    Address     `json:"caller"`
	Scheme    CreateScheme `json:"scheme"`
	Value     U256        `json:"value"`
	InitCode  Bytes       `json:"init_code"`
	GasLimit  uint64      `json:"gas_limit"`
}
type CreateScheme struct {
    SchemeType SchemeType `json:"scheme_type"` // can be "create" or "create2"
    Salt       *U256   `json:"salt,omitempty"` // salt is optional for Create
}

// SchemeType represents the type of creation scheme
type SchemeType int

const (
    SchemeTypeCreate SchemeType = iota
    SchemeTypeCreate2
)
type EOFCreateKindType int

const (
    TxK EOFCreateKindType = iota
    OpcodeK
)

/// Inputs for EOF create call.
type EOFCreateInputs struct {
    Caller   Address         `json:"caller"`
    Value    U256           `json:"value"`
    GasLimit uint64          `json:"gas_limit"`
    Kind     EOFCreateKind   `json:"kind"`
}

type EOFCreateKind struct {
    Kind EOFCreateKindType
    Data interface{} // Use an interface to hold the associated data (Tx or Opcode)
}

// TxKind represents a transaction-based EOF create kind.
type Tx struct {
    InitData Bytes `json:"initdata"`
}

// OpcodeKind represents an opcode-based EOF create kind.
type Opcode struct {
    InitCode        Eof     `json:"initcode"`
    Input           Bytes   `json:"input"`
    CreatedAddress  Address  `json:"created_address"`
}
type Eof struct {
    Header EofHeader `json:"header"`
    Body   EofBody   `json:"body"`
    Raw    Bytes     `json:"raw"`
}

// EofHeader represents the header of the EOF.
type EofHeader struct {
    TypesSize            uint16 `json:"types_size"`
    CodeSizes            []uint16 `json:"code_sizes"`
    ContainerSizes       []uint16 `json:"container_sizes"`
    DataSize             uint16 `json:"data_size"`
    SumCodeSizes         int     `json:"sum_code_sizes"`
    SumContainerSizes    int     `json:"sum_container_sizes"`
}


// Marshaler interface for custom marshaling
type Marshaler interface {
    MarshalType(val reflect.Value, b []byte, lastWrittenIdx uint64) (nextIdx uint64, err error)
}

// Unmarshaler interface for custom unmarshaling
type Unmarshaler interface {
    UnmarshalType(target reflect.Value, b []byte, lastReadIdx uint64) (nextIdx uint64, err error)
}

// Implement Marshaler for EOFCreateInputs
func (eci EOFCreateInputs) MarshalType(val reflect.Value, b []byte, lastWrittenIdx uint64) (nextIdx uint64, err error) {
    jsonData, err := json.Marshal(eci)
    if err != nil {
        return lastWrittenIdx, err
    }

    return uint64(len(jsonData)), nil
}

// Implement Unmarshaler for EOFCreateInputs
func (eci *EOFCreateInputs) UnmarshalType(target reflect.Value, b []byte, lastReadIdx uint64) (nextIdx uint64, err error) {
    err = json.Unmarshal(b, eci)
    if err != nil {
        return lastReadIdx, err
    }

    return lastReadIdx + uint64(len(b)), nil
}
type EofBody struct {
    TypesSection     []TypesSection `json:"types_section"`
    CodeSection      []Bytes      `json:"code_section"`      // Using [][]byte for Vec<Bytes>
    ContainerSection []Bytes       `json:"container_section"` // Using [][]byte for Vec<Bytes>
    DataSection      Bytes        `json:"data_section"`      // Using []byte for Bytes
    IsDataFilled     bool           `json:"is_data_filled"`
}

// TypesSection represents a section describing the types used in the EOF body.
type TypesSection struct {
    Inputs        uint8 `json:"inputs"`        // 1 byte
    Outputs       uint8 `json:"outputs"`       // 1 byte
    MaxStackSize  uint16 `json:"max_stack_size"` // 2 bytes
}

// MarshalJSON customizes the JSON marshalling for EOFCreateKind.
func (e *EOFCreateKind) MarshalJSON() ([]byte, error) {
    switch e.Kind {
    case TxK:
        return json.Marshal(map[string]interface{}{
            "type": "Tx",
            "data": e.Data,
        })
    case OpcodeK:
        return json.Marshal(map[string]interface{}{
            "type": "Opcode",
            "data": e.Data,
        })
    default:
        return nil, fmt.Errorf("unknown EOFCreateKind type")
    }
}

// UnmarshalJSON customizes the JSON unmarshalling for EOFCreateKind.
func (e *EOFCreateKind) UnmarshalJSON(data []byte) error {
    var aux struct {
        Type string          `json:"type"`
        Data json.RawMessage `json:"data"`
    }

    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }

    switch aux.Type {
    case "Tx":
        var tx Tx
        if err := json.Unmarshal(aux.Data, &tx); err != nil {
            return err
        }
        e.Kind = TxK
        e.Data = tx
    case "Opcode":
        var opcode Opcode
        if err := json.Unmarshal(aux.Data, &opcode); err != nil {
            return err
        }
        e.Kind = OpcodeK
        e.Data = opcode
    default:
        return fmt.Errorf("unknown EOFCreateKind type: %s", aux.Type)
    }

    return nil
}