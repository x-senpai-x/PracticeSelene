package evm

import (
	"testing"
	"math/big"

	"github.com/stretchr/testify/assert"
	// "github.com/BlocSoc-iitr/selene/common"
)

// TestNewJournalState checks if NewJournalState initializes JournaledState correctly.
func TestNewJournalStateI(t *testing.T) {
	spec := DefaultSpecId()
	warmPreloadedAddresses := map[Address]struct{}{
		 {0x1}: {},
	}
	journalState := NewJournalState(spec, warmPreloadedAddresses)

	assert.Nil(t, journalState.State, "State should be initialized as nil")
	assert.Nil(t, journalState.TransientStorage, "TransientStorage should be initialized as nil")
	assert.Equal(t, spec, journalState.Spec, "Spec should be set to the provided SpecId")
	assert.Equal(t, warmPreloadedAddresses, journalState.WarmPreloadedAddresses, "WarmPreloadedAddresses should be set correctly")
	assert.Equal(t, uint(0), journalState.Depth, "Depth should be initialized to 0")
	assert.Empty(t, journalState.Journal, "Journal should be initialized empty")
	assert.Empty(t, journalState.Logs, "Logs should be initialized empty")
}

// TestSetSpecId verifies that the setSpecId method correctly updates the SpecId in JournaledState.
func TestSetSpecIdI(t *testing.T) {
	spec := DefaultSpecId()
	journalState := NewJournalState(spec, nil)

	newSpec := SpecId(2)
	journalState.setSpecId(newSpec)
	assert.Equal(t, newSpec, journalState.Spec, "SpecId should be updated to the new SpecId")
}

// TestInnerEvmContextWithJournalState ensures InnerEvmContext integrates correctly with JournaledState.
func TestInnerEvmContextWithJournalState(t *testing.T) {
	db := NewEmptyDB()
	ctx := NewInnerEvmContext(db)

	assert.NotNil(t, ctx.JournaledState, "JournaledState should be initialized")
	assert.Nil(t, ctx.JournaledState.State, "State in JournaledState should be nil on initialization")
	assert.Equal(t, ctx.JournaledState.Spec, DefaultSpecId(), "Spec in JournaledState should match DefaultSpecId")
}

// TestTransientStorageInitialization ensures TransientStorage is initialized correctly in JournaledState.
func TestTransientStorageInitialization(t *testing.T) {
	spec := DefaultSpecId()
	journalState := NewJournalState(spec, nil)
	
	// Initialize a TransientStorage entry
	key := Key{
		Account: Address  {0x1},
		Slot:    U256(big.NewInt(0)),
	}
	value := U256(big.NewInt(12345))
	journalState.TransientStorage = make(TransientStorage)
	journalState.TransientStorage[key] = value
	
	assert.Equal(t, value, journalState.TransientStorage[key], "TransientStorage should store the correct value for the key")
}

// TestLogInitialization verifies that Logs are initialized and added to JournaledState.
func TestLogInitialization(t *testing.T) {
	spec := DefaultSpecId()
	journalState := NewJournalState(spec, nil)

	log := Log[LogData]{
		Address: Address  {0x1},
		Data: LogData{
			Topics: []B256{{0x1, 0x2, 0x3, 0x4}},
			Data:   Bytes{0x10, 0x20, 0x30},
		},
	}
	journalState.Logs = append(journalState.Logs, log)

	assert.Len(t, journalState.Logs, 1, "Logs should contain one entry")
	assert.Equal(t, log, journalState.Logs[0], "Log entry should match the added log")
}