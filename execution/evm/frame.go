package evm

type CallOutcome struct{}
type CreateOutcome struct{}

// FrameResultType represents the type of FrameResult.
type FrameResultType int

const (
	Call FrameResultType = iota
	Create
	EOFCreate
)

// FrameResult represents the outcome of a frame, which can be Call, Create, or EOFCreate.
type FrameResult struct {
	ResultType FrameResultType
	Call       *CallOutcome
	Create     *CreateOutcome
}