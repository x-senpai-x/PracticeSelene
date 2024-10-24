// +build optimism

package specs

import (
	"encoding/json"
)

// SpecId represents the specification IDs and their activation block for optimism.
type SpecId uint8

// Enumeration of specification IDs for optimism
const (
	FRONTIER         SpecId = 0  // Frontier
	FRONTIER_THAWING SpecId = 1  // Frontier Thawing
	HOMESTEAD        SpecId = 2  // Homestead
	DAO_FORK         SpecId = 3  // DAO Fork
	TANGERINE        SpecId = 4  // Tangerine Whistle
	SPURIOUS_DRAGON   SpecId = 5  // Spurious Dragon
	BYZANTIUM        SpecId = 6  // Byzantium
	CONSTANTINOPLE   SpecId = 7  // Constantinople
	PETERSBURG       SpecId = 8  // Petersburg
	ISTANBUL        SpecId = 9  // Istanbul
	MUIR_GLACIER     SpecId = 10 // Muir Glacier
	BERLIN           SpecId = 11 // Berlin
	LONDON           SpecId = 12 // London
	ARROW_GLACIER    SpecId = 13 // Arrow Glacier
	GRAY_GLACIER     SpecId = 14 // Gray Glacier
	MERGE            SpecId = 15 // Paris/Merge
	BEDROCK          SpecId = 16 // Bedrock
	REGOLITH         SpecId = 17 // Regolith
	SHANGHAI         SpecId = 18 // Shanghai
	CANYON           SpecId = 19 // Canyon
	CANCUN           SpecId = 20 // Cancun
	ECOTONE          SpecId = 21 // Ecotone
	FJORD            SpecId = 22 // Fjord
	PRAGUE           SpecId = 23 // Praque
	PRAGUE_EOF       SpecId = 24 // Praque+EOF
	LATEST           SpecId = 255 // Latest (u8::MAX)
)

// MarshalJSON implements the json.Marshaler interface for SpecId.
func (s SpecId) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint8(s))
}

// UnmarshalJSON implements the json.Unmarshaler interface for SpecId.
func (s *SpecId) UnmarshalJSON(data []byte) error {
	var val uint8
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	*s = SpecId(val)
	return nil
}
