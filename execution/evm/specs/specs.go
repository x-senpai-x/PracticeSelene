// +build !optimism

package specs

import (
	"encoding/json"
)

// SpecId represents the specification IDs and their activation block.
type SpecId uint8

// Enumeration of specification IDs
const (
	FRONTIER         SpecId = 0  // Frontier               0
	FRONTIER_THAWING SpecId = 1  // Frontier Thawing       200000
	HOMESTEAD        SpecId = 2  // Homestead              1150000
	DAO_FORK         SpecId = 3  // DAO Fork               1920000
	TANGERINE        SpecId = 4  // Tangerine Whistle      2463000
	SPURIOUS_DRAGON   SpecId = 5  // Spurious Dragon        2675000
	BYZANTIUM        SpecId = 6  // Byzantium              4370000
	CONSTANTINOPLE   SpecId = 7  // Constantinople         7280000 is overwritten with PETERSBURG
	PETERSBURG       SpecId = 8  // Petersburg             7280000
	ISTANBUL        SpecId = 9  // Istanbul               9069000
	MUIR_GLACIER     SpecId = 10 // Muir Glacier           9200000
	BERLIN           SpecId = 11 // Berlin                 12244000
	LONDON           SpecId = 12 // London                 12965000
	ARROW_GLACIER    SpecId = 13 // Arrow Glacier          13773000
	GRAY_GLACIER     SpecId = 14 // Gray Glacier           15050000
	MERGE            SpecId = 15 // Paris/Merge            15537394
	SHANGHAI         SpecId = 16 // Shanghai               17034870
	CANCUN           SpecId = 17 // Cancun                 19426587
	PRAGUE           SpecId = 18 // Praque                 TBD
	PRAGUE_EOF       SpecId = 19 // Praque+EOF             TBD
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

// Spec interface defining the behavior of Ethereum specs.
type Spec interface {
	GetSpecID() SpecId
}

// CreateSpec creates a new specification struct based on the provided SpecId.
func CreateSpec(specId SpecId) Spec {
	switch specId {
	case FRONTIER:
		return &FrontierSpec{}
	case FRONTIER_THAWING:
		// No changes for EVM spec
		return nil
	case HOMESTEAD:
		return &HomesteadSpec{}
	case DAO_FORK:
		// No changes for EVM spec
		return nil
	case TANGERINE:
		return &TangerineSpec{}
	case SPURIOUS_DRAGON:
		return &SpuriousDragonSpec{}
	case BYZANTIUM:
		return &ByzantiumSpec{}
	case PETERSBURG:
		return &PetersburgSpec{}
	case ISTANBUL:
		return &IstanbulSpec{}
	case MUIR_GLACIER:
		// No changes for EVM spec
		return nil
	case BERLIN:
		return &BerlinSpec{}
	case LONDON:
		return &LondonSpec{}
	case ARROW_GLACIER:
		// No changes for EVM spec
		return nil
	case GRAY_GLACIER:
		// No changes for EVM spec
		return nil
	case MERGE:
		return &MergeSpec{}
	case SHANGHAI:
		return &ShanghaiSpec{}
	case CANCUN:
		return &CancunSpec{}
	case PRAGUE:
		return &PragueSpec{}
	case PRAGUE_EOF:
		return &PragueEofSpec{}
	case LATEST:
		return &LatestSpec{}
	default:
		return nil
	}
}

// Spec structures
type FrontierSpec struct{}
func (s *FrontierSpec) GetSpecID() SpecId { return FRONTIER }

type HomesteadSpec struct{}
func (s *HomesteadSpec) GetSpecID() SpecId { return HOMESTEAD }

type TangerineSpec struct{}
func (s *TangerineSpec) GetSpecID() SpecId { return TANGERINE }

type SpuriousDragonSpec struct{}
func (s *SpuriousDragonSpec) GetSpecID() SpecId { return SPURIOUS_DRAGON }

type ByzantiumSpec struct{}
func (s *ByzantiumSpec) GetSpecID() SpecId { return BYZANTIUM }

type PetersburgSpec struct{}
func (s *PetersburgSpec) GetSpecID() SpecId { return PETERSBURG }

type IstanbulSpec struct{}
func (s *IstanbulSpec) GetSpecID() SpecId { return ISTANBUL }

type BerlinSpec struct{}
func (s *BerlinSpec) GetSpecID() SpecId { return BERLIN }

type LondonSpec struct{}
func (s *LondonSpec) GetSpecID() SpecId { return LONDON }

type MergeSpec struct{}
func (s *MergeSpec) GetSpecID() SpecId { return MERGE }

type ShanghaiSpec struct{}
func (s *ShanghaiSpec) GetSpecID() SpecId { return SHANGHAI }

type CancunSpec struct{}
func (s *CancunSpec) GetSpecID() SpecId { return CANCUN }

type PragueSpec struct{}
func (s *PragueSpec) GetSpecID() SpecId { return PRAGUE }

type PragueEofSpec struct{}
func (s *PragueEofSpec) GetSpecID() SpecId { return PRAGUE_EOF }

type LatestSpec struct{}
func (s *LatestSpec) GetSpecID() SpecId { return LATEST }
