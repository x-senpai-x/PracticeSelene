//go:build !optimism_default_handler || negate_optimism_default_handler
// +build !optimism_default_handler negate_optimism_default_handler

package evm

import "testing"

func TestShouldReturnFalse(t *testing.T) {
	// Test logic
	result := getDefaultOptimismSetting()
	if result != false {
		panic("Test failed")
	}
}