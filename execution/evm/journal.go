package evm
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