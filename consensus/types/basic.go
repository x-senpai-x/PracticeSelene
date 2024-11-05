package types
import ("github.com/ethereum/go-ethereum/common")
type Address = common.Address
type Bytes32 = [32]byte
type LogsBloom = [256]byte
type BLSPubKey = [48]byte
type SignatureBytes = [96]byte
type Transaction = [1073741824]byte //1073741824
