package execution

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (a *Account) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{a.Nonce, a.Balance, a.StorageHash, a.CodeHash})
}

func (a *Account) DecodeRLP(s *rlp.Stream) error {
	var accountRLP struct {
		Nonce       uint64
		Balance     *big.Int
		StorageHash common.Hash
		CodeHash    common.Hash
	}
	if err := s.Decode(&accountRLP); err != nil {
		return err
	}
	a.Nonce = accountRLP.Nonce
	a.Balance = accountRLP.Balance
	a.StorageHash = accountRLP.StorageHash
	a.CodeHash = accountRLP.CodeHash

	return nil
}

func TestSharedPrefixLength(t *testing.T) {
	t.Run("Match first 5 nibbles", func(t *testing.T) {
		path := []byte{0x12, 0x13, 0x14, 0x6f, 0x6c, 0x64, 0x21}
		pathOffset := 6
		nodePath := []byte{0x6f, 0x6c, 0x63, 0x21}

		sharedLen := sharedPrefixLength(path, pathOffset, nodePath)
		assert.Equal(t, 5, sharedLen)
	})

	t.Run("Match first 7 nibbles with skip", func(t *testing.T) {
		path := []byte{0x12, 0x13, 0x14, 0x6f, 0x6c, 0x64, 0x21}
		pathOffset := 5
		nodePath := []byte{0x14, 0x6f, 0x6c, 0x64, 0x11}

		sharedLen := sharedPrefixLength(path, pathOffset, nodePath)
		assert.Equal(t, 7, sharedLen)
	})

	t.Run("No match", func(t *testing.T) {
		path := []byte{0x12, 0x34, 0x56}
		pathOffset := 0
		nodePath := []byte{0x78, 0x9a, 0xbc}

		sharedLen := sharedPrefixLength(path, pathOffset, nodePath)
		assert.Equal(t, 0, sharedLen)
	})

	t.Run("Complete match", func(t *testing.T) {
		path := []byte{0x12, 0x34, 0x56}
		pathOffset := 0
		nodePath := []byte{0x00, 0x12, 0x34, 0x56}

		sharedLen := sharedPrefixLength(path, pathOffset, nodePath)
		assert.Equal(t, 6, sharedLen)
	})
}

func TestSkipLength(t *testing.T) {
	testCases := []struct {
		name     string
		node     []byte
		expected int
	}{
		{"Empty node", []byte{}, 0},
		{"Nibble 0", []byte{0x00}, 2},
		{"Nibble 1", []byte{0x10}, 1},
		{"Nibble 2", []byte{0x20}, 2},
		{"Nibble 3", []byte{0x30}, 1},
		{"Nibble 4", []byte{0x40}, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := skipLength(tc.node)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetNibble(t *testing.T) {
	path := []byte{0x12, 0x34, 0x56}

	testCases := []struct {
		offset   int
		expected byte
	}{
		{0, 0x1},
		{1, 0x2},
		{2, 0x3},
		{3, 0x4},
		{4, 0x5},
		{5, 0x6},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Offset %d", tc.offset), func(t *testing.T) {
			result := getNibble(path, tc.offset)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsEmptyValue(t *testing.T) {
	t.Run("Empty slot", func(t *testing.T) {
		value := []byte{0x80}
		assert.True(t, isEmptyValue(value))
	})

	t.Run("Empty account", func(t *testing.T) {
		emptyAccount := Account{
			Nonce:       0,
			Balance:     big.NewInt(0),
			StorageHash: [32]byte{0x56, 0xe8, 0x1f, 0x17, 0x1b, 0xcc, 0x55, 0xa6, 0xff, 0x83, 0x45, 0xe6, 0x92, 0xc0, 0xf8, 0x6e, 0x5b, 0x48, 0xe0, 0x1b, 0x99, 0x6c, 0xad, 0xc0, 0x01, 0x62, 0x2f, 0xb5, 0xe3, 0x63, 0xb4, 0x21},
			CodeHash:    [32]byte{0xc5, 0xd2, 0x46, 0x01, 0x86, 0xf7, 0x23, 0x3c, 0x92, 0x7e, 0x7d, 0xb2, 0xdc, 0xc7, 0x03, 0xc0, 0xe5, 0x00, 0xb6, 0x53, 0xca, 0x82, 0x27, 0x3b, 0x7b, 0xfa, 0xd8, 0x04, 0x5d, 0x85, 0xa4, 0x70},
		}
		encodedEmptyAccount, err := rlp.EncodeToBytes(&emptyAccount)
		require.NoError(t, err)
		assert.True(t, isEmptyValue(encodedEmptyAccount))
	})

	t.Run("Non-empty value", func(t *testing.T) {
		value := []byte{0x01, 0x23, 0x45}
		assert.False(t, isEmptyValue(value))
	})
}

func TestEncodeAccount(t *testing.T) {
	proof := &EIP1186ProofResponse{
		Nonce:       1,
		Balance:     uint256.NewInt(1000),
		StorageHash: [32]byte{1, 2, 3},
		CodeHash:    [32]byte{4, 5, 6},
	}

	encoded, err := EncodeAccount(proof)
	require.NoError(t, err)

	var decodedAccount Account
	err = rlp.DecodeBytes(encoded, &decodedAccount)
	require.NoError(t, err)

	// Assert the decoded values match the original
	assert.Equal(t, uint64(1), decodedAccount.Nonce)
	assert.Equal(t, big.NewInt(1000), decodedAccount.Balance)
	assert.Equal(t, proof.StorageHash, decodedAccount.StorageHash)
	assert.Equal(t, proof.CodeHash, decodedAccount.CodeHash)
}

func TestKeccak256(t *testing.T) {
	input := []byte("hello world")
	expected, _ := hex.DecodeString("47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad")

	result := keccak256(input)
	assert.Equal(t, expected, result)
}

type Proof struct {
	Address      string   `json:"address"`
	AccountProof []string `json:"accountProof"`
	Balance      string   `json:"balance"`
	CodeHash     string   `json:"codeHash"`
	Nonce        string   `json:"nonce"`
	StorageHash  string   `json:"storageHash"`
	StorageProof string   `json:"storageProof"`
}

func TestVerifyProof(t *testing.T) {
	jsonString := `{
	"address":"0x457a22804cf255ee8c1b7628601c5682b3d70c71",
	"accountProof":[
		"0xf90211a0c48ed6536f719a4e24c1195b099fd19f127448c0e2b4ad5729dcdcccb41a56b5a0a338e0f373eac20e4ad9528f9895da93762623350bb5defa693263ea6f9af4afa0304e41961a40be33e5213d3e583c0d30989e795b541c810e8456c12a7478c841a05b26d8545de569451cd3baee25ef6d525f6c27c89844b76ccebf99801708c1d9a069b7051e21066a40847fbe6e4217628f6979b1a361819c7ebe0980c063d684f5a06f5948452762bfe7061717765ab322a5e086ece9c4eddf442a047e253b557dc4a0ef37b2aaf540e598e0e046ee6a8c51c1c100d68dcda70bea1ebe4924ef7911f3a03218af1883a1bfb9828b0c3f07f16a8df859e81543633808ec625164db6b3644a0121b6fdd1951f7915bd55c96dc2cc546fa42bfadb4cfa3fc1a2cc3de232df289a0fdf5a771f4ee044d9c6a7ddd6d797ca102022d4786db1332c5530091f52a2238a08e59837befa581ee0e4d2d9b2caeac403106090117844590d876df445df09509a0d0eb738155e240e6c069c765102f526b54124d788c8e501ed96a6a29a326f4d7a0c14d799f056ca9d6bf91a10272dff64cf34eb4aff42e7b56bf14176e2e36a58aa03ae0cf5dc0b96aa9755207ec060c45b149e071f1453e7416acd282660a66a371a0495a02815d4434895bf7997771016ab2bf29c1f970c778c696444d4b5573ae87a008d340cb14293c576f934b7f8cbab6f0b07f6d4804957d6f8d6c48e2465d027f80",
		"0xf90211a0a0b5e06bee86e09cc7f63d64ef1f3bd9a970ba8881d52b9a1a60849eef16b535a022756912563afa1c6cc7a49d0f34704f5919df048091561e21511bff585fc6d2a0eaa5bd3c14a851aeaeff1a7dbafb5456d4a218f12c63f0795ba2679d8f412ba6a07f8b49a363a58d81b7c51eab08998532ff3efdec14e6015d43429180a5a74929a0e62a8be403d29d6972426603c4a0962a24d5c56325bf1661c0984414605eaa55a0d567f6a8520f25ed64ece7d37b405374a51c3af48356c6fa72efec22f7d42e2aa0696255c12473694333c6724ff234a04f6995b8eb376e8e8be0c171965767ca2ea0b1b08ecb635708f8c0f42dffe86e09ac6b3bfe8319e544797a4199bde4d1fd15a05c0e9906e3d20bbec170f769824250862fd7dad378d6cfeda9b713acf3b72568a03d160d95900a9bc2e7e9c08766b4392a9ac86b7e006979c426e0885229e5f44da00bc337a5b553d0ffb67859b28cc59246262bf532c2b14dd456f160d778540b34a005d784ba4f478f2f315f1b678fbfc39ed2c0b01b02b55ed82ea9d62369cd35faa023f24fd63f87a947ee148c8001bc4039a5f481fca1e0f3f10a3791a9e0451571a01ffb2845581388c6b915135a9ee856bcc896f2def7eeeb88bdfd53ae32323431a09d7bde3d951bcc3505684616bf760dfdd69e03e6b2ca2a180bc7b6fc7922d9c6a01b29670cbbe0bb02bf764032d347af708db1cd34dd2a9ceaa056d0c96ebcc06380",
		"0xf90211a0532c4b817232369b476e2f512431248940bc2b62b268750eafb19725b1ea3f61a023d73bdeaa7d23f1ae3c7e45a961f6828a8c84da890802d9e76c4ad040f9dfa8a0b2b86151f708e0970051496defa3a21a9e38add3710dcdf3128f350eed7545b7a0c6f14ab140c8c849900d90e72430cc8ed9e05bb11c0bc1ef74653813fe184908a085382b28cc5d147da5be0ad99cf65ba09a70031f9ad6ddb5a1b8676adf620d27a05002c30689fcafbe153d3dd120d8e2ed2d867ed4dc38e29c488914cfb2d773e1a0061763ee76bdc20123bd89151b0d3311d3500ffe9502013f79147983497323a2a0b27cccc06c21e8ec8a41076e872360263864d2c2a033bff15cdc4c422f5efbc4a061d3b9c9582aaa05e4c0c4c5e269194ad2337465c88fe15721dce420872df88aa0bf0cf79d1546129128ee59aa52f987f0a92d4ee8b50d34178d49ccbc3b542fbda08dcb375e470c4b7e33e9b30a18c77557c3afaf29a3423ebc457193ffa4afafeda00f37ee5c7e78a14120c393f96aaa2eb77c7a2b290b3e390d868f901d7935b91ea0189ba4abb7a368dc21faa9009e56f08dbabbf2b169237b2c3343cdce84b9734ba001a06747b252f00070e5c6857db4d2bd984ac1e3ca8e2c1768a1b15c7caaa32aa06f68846bda765c1ebe37c4f460a5501ea49d32e2873e0333b1db3f8b11353345a046ee87393f4507c48fffeaefb60f86d39be7f3164101a254c978154bdf440e0680",
		"0xf90211a01d3ed8ac26e69d88bf7c5ede07a54e9e58497b152539ca91e50a92287db07f0ea085e2443725c20842d26d4989ba1a56dfb13fbafbd95628dd6ea0ecfaebbd956ea06d11f90d431eabe083bde15c02ca665d6bd270102108105d1179aa3e350ea327a0c95a53dd831c6b4f71406d0549369d20a6a19831ee70cb2bbfce0089328c3ceba0b84a29614ca66250d59c07792808522acf7137ed2027f1a0148ffe1b467c5623a01abe3994fd1fa6276a00a2bf2db562a3418bcd6ef53c2daf62f899d95e31e15aa01b26905ddf68b2ec3ee23558c6c931f2d86436caeaf078d772b5f659e900f782a02b7f348cd686aa44c2dc0d8ab03dccc700728d2acbb645d84036ee5b26dcb4aaa0ed6662d9fcde2a8fd39bc7f26c8537bec1a6193955cd63d739cd6c8eeb5f2c35a090a7e073bb5df6de34707cb679c2ef359e66550c557abd28c8dbfec598a7a5b6a0d8899a7e2839d10b61c59546af49af8b723dab8ddcafc68ecdf4994d42b5badaa09aa471b32e1f8ac6b811137e6cc15245fcb4d8e40b6f27c0955e31ae3e428619a00e6f9245fc44c37708a41ccde7bb764b9e54b69f12c79c5c1ad9eedc21075d59a0d83ec73c7ef8bfea0448174f53a174ea07397d8ba9a55cd1589ca9e9a4d41b7da0b3ceb20f69ce1a99b8e110bb3c56213ed047167cf9b794b9ab0d5dd3227f9d62a0a9e3612f28c1cd13e9f1ac7b6e5f984afa226b2b6037ebb2f8d6fcc4505c7f1c80",
		"0xf90211a0fd032ce4009ac0aa95170293240d6bef6cefd073fcf5d430e2a526410dab705aa09ed14f5069549d06740b6946a651b4bd90bff6ad44133b779fdcada5b621d690a0b833494b9e3f58f773aa5cec7645ceac4ebfdcd985f234d33eb283415a14cdd4a0c2e1ff80fdcff52539b161fb2ce25cd2825bc61453e63f72347d5be029441ed6a082cb4d21a71f220cc1fc05e6d544a79a102c296a0a30634006757aa9dcf0dacba03930cac361c2eb6edbbe3cb453c1d301aedb25bfc0c12b32adb56849545a8390a0bdbcc436e5eded0ca5daf6f37ea2626d3977f745245190a62d24f0fe4e12b8eaa05ab7a229f5a628f7606948ff694c05a5a08f87e0e3c3ad1f45b747eec521c5eea0e9a0e18f800afc38bcabe3cbca21695a0d399b235285de2819791d88baa7c339a0bbb17f8b5dfbb24a4a15e6d606c54d67c103dd44f0f1246d6604e9b71a671866a048fe8956aa652d57f0322990f94c9df6af0f7c28c41776b1c80cc0eafa5ab458a01ed5fcbfd2efe7381e1f98081bc98896086615236cb8f7f7a44a04493286a0a0a020c202d8b2b22ed074692df8dd21df64105643b9a97d1fdd0ec235a59711909da0d8ac47fa535011dc28cd3f649f9391f881ba3822d05b4d25421a21ce3c50499ea025db02e0e16e04f2b3e69dacac680d3c8fbfcc5cadbf441fdf0f303f43b7aaa1a0f646b6d69c1fce8ad6fc064c1341da989a59c4157b5da5ff25b760f041707f0680",
		"0xf90211a0a57894cce27679d5a469be71ae0c582159b51a71381f0365404bbe1b2af0d8d3a03243d088fb7f888a461e06d9339dda036d10fdc13fdc18d68680680117133cc8a0fe66886652b91264cb2e9e99cdae776928be462c0e76a1b2af3673de8c1e09f4a08a47de5ebfaa75fe79082627efa51108ad86235e6576c6dab5cad9f374dc3a33a0663fcc26c76d08a85004b4ca2e02baaf067ecbf1e0a121dc95be500cf541d56ea092f22428ae816de7a5a95e46ac3579d614e4c44d55be89d869a815a2e654e599a04adf75bc024ab9dc2bc280249a610422166f52c9fbc0bf9b401984b79e03211aa0d6ae689b2f5dabbc30e1e612cafc960fbb342f523617f7f6a8df1845a7564154a0bee5ca129557312716be21cf0e91273532e573ed162f0c8cdf0b1f1da38a30cca01aa1ca8998d04f89182c022aebcf6828d5d808ef4e663bf690e51aeb133ef1d5a0147cac105fa6c368d8a9013cd1873271ddacda2be2345051fdd3fa50a2c003c8a0caf1ff1075e0f30fb12b29e1278b61d69bf5cbd059131ba18f4b89422c188cd2a0fc0dc4f79e218c884075a69b52bfe84b8bbbb98e9e252299bda742ef77ee4264a05153dc26c8c558e394488fc70a45c0d863f2a930a215fb6fa0e1fdc92d990b31a010899ac22f43b3031f16cc9c660d7c5a93bf3ce5538a545056d9b1d364b06195a060e86b07bb6e96e4ea87eaeaa162108408bd5fadcc3496910c9fc59df5c0320480",
		"0xf90191a0a1f0ab277fc1d351526eadeb8bcd95db08d710f0d3d1a0cc0cc8609d961a47d1a0c756a865d28d6bdcff435c90f5bf251107c230ad7558f57583d14be37e1ccaafa0828901d38fd39ade9bd74cbe07996e8cc45f8c43c6637fe7b7a43ab467e52098a022e4fb3dc11ca5e38ed9527e4eb534162c878ac048ac428abba29633b607d6258080a018d60336d1d9723e0d43567745464372e11f9f7a146abf163b0f43b338c0d827a00b44f7040c54b2bd63b0741062a96495dc4b92f7fec136c8effbc213cfffd02b80a0b81fcd189fbc66780a932367f874cbd4c85fed1fd37af109acc052f8ba80e51d80a0bbbcbe635193161c1e53442c06d0d14ed1af96af105a9fdd46a493195d62c1bba00050d231707972af60f163aa425401abb860f6c04b770fe0983e2f0f90ad6254a0804a9abf1b69bd16b1f6dc7cbbbf7592cd93fdf785b9cc61b01816583d9cd928a0ad85c29c3adbf2291cead9df6602560b93e485af7da680697ef9d1703f6f91cea040134ec8f47e1c599c0b7b4d237b8d3349f026fcfbacf885673438ffd2e9c72e80",
		"0xf851a01a5a4c757ab8653aad3a83cbb491142c776b91e252ffe849c253dab5d109819380808080808080a09fc998b2c0347275d9f98a9ac5bbb205387c04387672cda1a0360c4ef7cffb1a8080808080808080",
		"0xf86d9d205eb14112905ead3a2b16b733af37911d729cec51f559baf570332371b84df84b8087cc756b8ad2e000a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	],
	"balance":"0xcc756b8ad2e000",
	"codeHash":"0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
	"nonce":"0x0",
	"storageHash":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
	,"storageProof":""
	}`

	accountProof := []string{
		"0xf90211a0c48ed6536f719a4e24c1195b099fd19f127448c0e2b4ad5729dcdcccb41a56b5a0a338e0f373eac20e4ad9528f9895da93762623350bb5defa693263ea6f9af4afa0304e41961a40be33e5213d3e583c0d30989e795b541c810e8456c12a7478c841a05b26d8545de569451cd3baee25ef6d525f6c27c89844b76ccebf99801708c1d9a069b7051e21066a40847fbe6e4217628f6979b1a361819c7ebe0980c063d684f5a06f5948452762bfe7061717765ab322a5e086ece9c4eddf442a047e253b557dc4a0ef37b2aaf540e598e0e046ee6a8c51c1c100d68dcda70bea1ebe4924ef7911f3a03218af1883a1bfb9828b0c3f07f16a8df859e81543633808ec625164db6b3644a0121b6fdd1951f7915bd55c96dc2cc546fa42bfadb4cfa3fc1a2cc3de232df289a0fdf5a771f4ee044d9c6a7ddd6d797ca102022d4786db1332c5530091f52a2238a08e59837befa581ee0e4d2d9b2caeac403106090117844590d876df445df09509a0d0eb738155e240e6c069c765102f526b54124d788c8e501ed96a6a29a326f4d7a0c14d799f056ca9d6bf91a10272dff64cf34eb4aff42e7b56bf14176e2e36a58aa03ae0cf5dc0b96aa9755207ec060c45b149e071f1453e7416acd282660a66a371a0495a02815d4434895bf7997771016ab2bf29c1f970c778c696444d4b5573ae87a008d340cb14293c576f934b7f8cbab6f0b07f6d4804957d6f8d6c48e2465d027f80",
		"0xf90211a0a0b5e06bee86e09cc7f63d64ef1f3bd9a970ba8881d52b9a1a60849eef16b535a022756912563afa1c6cc7a49d0f34704f5919df048091561e21511bff585fc6d2a0eaa5bd3c14a851aeaeff1a7dbafb5456d4a218f12c63f0795ba2679d8f412ba6a07f8b49a363a58d81b7c51eab08998532ff3efdec14e6015d43429180a5a74929a0e62a8be403d29d6972426603c4a0962a24d5c56325bf1661c0984414605eaa55a0d567f6a8520f25ed64ece7d37b405374a51c3af48356c6fa72efec22f7d42e2aa0696255c12473694333c6724ff234a04f6995b8eb376e8e8be0c171965767ca2ea0b1b08ecb635708f8c0f42dffe86e09ac6b3bfe8319e544797a4199bde4d1fd15a05c0e9906e3d20bbec170f769824250862fd7dad378d6cfeda9b713acf3b72568a03d160d95900a9bc2e7e9c08766b4392a9ac86b7e006979c426e0885229e5f44da00bc337a5b553d0ffb67859b28cc59246262bf532c2b14dd456f160d778540b34a005d784ba4f478f2f315f1b678fbfc39ed2c0b01b02b55ed82ea9d62369cd35faa023f24fd63f87a947ee148c8001bc4039a5f481fca1e0f3f10a3791a9e0451571a01ffb2845581388c6b915135a9ee856bcc896f2def7eeeb88bdfd53ae32323431a09d7bde3d951bcc3505684616bf760dfdd69e03e6b2ca2a180bc7b6fc7922d9c6a01b29670cbbe0bb02bf764032d347af708db1cd34dd2a9ceaa056d0c96ebcc06380",
		"0xf90211a0532c4b817232369b476e2f512431248940bc2b62b268750eafb19725b1ea3f61a023d73bdeaa7d23f1ae3c7e45a961f6828a8c84da890802d9e76c4ad040f9dfa8a0b2b86151f708e0970051496defa3a21a9e38add3710dcdf3128f350eed7545b7a0c6f14ab140c8c849900d90e72430cc8ed9e05bb11c0bc1ef74653813fe184908a085382b28cc5d147da5be0ad99cf65ba09a70031f9ad6ddb5a1b8676adf620d27a05002c30689fcafbe153d3dd120d8e2ed2d867ed4dc38e29c488914cfb2d773e1a0061763ee76bdc20123bd89151b0d3311d3500ffe9502013f79147983497323a2a0b27cccc06c21e8ec8a41076e872360263864d2c2a033bff15cdc4c422f5efbc4a061d3b9c9582aaa05e4c0c4c5e269194ad2337465c88fe15721dce420872df88aa0bf0cf79d1546129128ee59aa52f987f0a92d4ee8b50d34178d49ccbc3b542fbda08dcb375e470c4b7e33e9b30a18c77557c3afaf29a3423ebc457193ffa4afafeda00f37ee5c7e78a14120c393f96aaa2eb77c7a2b290b3e390d868f901d7935b91ea0189ba4abb7a368dc21faa9009e56f08dbabbf2b169237b2c3343cdce84b9734ba001a06747b252f00070e5c6857db4d2bd984ac1e3ca8e2c1768a1b15c7caaa32aa06f68846bda765c1ebe37c4f460a5501ea49d32e2873e0333b1db3f8b11353345a046ee87393f4507c48fffeaefb60f86d39be7f3164101a254c978154bdf440e0680",
		"0xf90211a01d3ed8ac26e69d88bf7c5ede07a54e9e58497b152539ca91e50a92287db07f0ea085e2443725c20842d26d4989ba1a56dfb13fbafbd95628dd6ea0ecfaebbd956ea06d11f90d431eabe083bde15c02ca665d6bd270102108105d1179aa3e350ea327a0c95a53dd831c6b4f71406d0549369d20a6a19831ee70cb2bbfce0089328c3ceba0b84a29614ca66250d59c07792808522acf7137ed2027f1a0148ffe1b467c5623a01abe3994fd1fa6276a00a2bf2db562a3418bcd6ef53c2daf62f899d95e31e15aa01b26905ddf68b2ec3ee23558c6c931f2d86436caeaf078d772b5f659e900f782a02b7f348cd686aa44c2dc0d8ab03dccc700728d2acbb645d84036ee5b26dcb4aaa0ed6662d9fcde2a8fd39bc7f26c8537bec1a6193955cd63d739cd6c8eeb5f2c35a090a7e073bb5df6de34707cb679c2ef359e66550c557abd28c8dbfec598a7a5b6a0d8899a7e2839d10b61c59546af49af8b723dab8ddcafc68ecdf4994d42b5badaa09aa471b32e1f8ac6b811137e6cc15245fcb4d8e40b6f27c0955e31ae3e428619a00e6f9245fc44c37708a41ccde7bb764b9e54b69f12c79c5c1ad9eedc21075d59a0d83ec73c7ef8bfea0448174f53a174ea07397d8ba9a55cd1589ca9e9a4d41b7da0b3ceb20f69ce1a99b8e110bb3c56213ed047167cf9b794b9ab0d5dd3227f9d62a0a9e3612f28c1cd13e9f1ac7b6e5f984afa226b2b6037ebb2f8d6fcc4505c7f1c80",
		"0xf90211a0fd032ce4009ac0aa95170293240d6bef6cefd073fcf5d430e2a526410dab705aa09ed14f5069549d06740b6946a651b4bd90bff6ad44133b779fdcada5b621d690a0b833494b9e3f58f773aa5cec7645ceac4ebfdcd985f234d33eb283415a14cdd4a0c2e1ff80fdcff52539b161fb2ce25cd2825bc61453e63f72347d5be029441ed6a082cb4d21a71f220cc1fc05e6d544a79a102c296a0a30634006757aa9dcf0dacba03930cac361c2eb6edbbe3cb453c1d301aedb25bfc0c12b32adb56849545a8390a0bdbcc436e5eded0ca5daf6f37ea2626d3977f745245190a62d24f0fe4e12b8eaa05ab7a229f5a628f7606948ff694c05a5a08f87e0e3c3ad1f45b747eec521c5eea0e9a0e18f800afc38bcabe3cbca21695a0d399b235285de2819791d88baa7c339a0bbb17f8b5dfbb24a4a15e6d606c54d67c103dd44f0f1246d6604e9b71a671866a048fe8956aa652d57f0322990f94c9df6af0f7c28c41776b1c80cc0eafa5ab458a01ed5fcbfd2efe7381e1f98081bc98896086615236cb8f7f7a44a04493286a0a0a020c202d8b2b22ed074692df8dd21df64105643b9a97d1fdd0ec235a59711909da0d8ac47fa535011dc28cd3f649f9391f881ba3822d05b4d25421a21ce3c50499ea025db02e0e16e04f2b3e69dacac680d3c8fbfcc5cadbf441fdf0f303f43b7aaa1a0f646b6d69c1fce8ad6fc064c1341da989a59c4157b5da5ff25b760f041707f0680",
		"0xf90211a0a57894cce27679d5a469be71ae0c582159b51a71381f0365404bbe1b2af0d8d3a03243d088fb7f888a461e06d9339dda036d10fdc13fdc18d68680680117133cc8a0fe66886652b91264cb2e9e99cdae776928be462c0e76a1b2af3673de8c1e09f4a08a47de5ebfaa75fe79082627efa51108ad86235e6576c6dab5cad9f374dc3a33a0663fcc26c76d08a85004b4ca2e02baaf067ecbf1e0a121dc95be500cf541d56ea092f22428ae816de7a5a95e46ac3579d614e4c44d55be89d869a815a2e654e599a04adf75bc024ab9dc2bc280249a610422166f52c9fbc0bf9b401984b79e03211aa0d6ae689b2f5dabbc30e1e612cafc960fbb342f523617f7f6a8df1845a7564154a0bee5ca129557312716be21cf0e91273532e573ed162f0c8cdf0b1f1da38a30cca01aa1ca8998d04f89182c022aebcf6828d5d808ef4e663bf690e51aeb133ef1d5a0147cac105fa6c368d8a9013cd1873271ddacda2be2345051fdd3fa50a2c003c8a0caf1ff1075e0f30fb12b29e1278b61d69bf5cbd059131ba18f4b89422c188cd2a0fc0dc4f79e218c884075a69b52bfe84b8bbbb98e9e252299bda742ef77ee4264a05153dc26c8c558e394488fc70a45c0d863f2a930a215fb6fa0e1fdc92d990b31a010899ac22f43b3031f16cc9c660d7c5a93bf3ce5538a545056d9b1d364b06195a060e86b07bb6e96e4ea87eaeaa162108408bd5fadcc3496910c9fc59df5c0320480",
		"0xf90191a0a1f0ab277fc1d351526eadeb8bcd95db08d710f0d3d1a0cc0cc8609d961a47d1a0c756a865d28d6bdcff435c90f5bf251107c230ad7558f57583d14be37e1ccaafa0828901d38fd39ade9bd74cbe07996e8cc45f8c43c6637fe7b7a43ab467e52098a022e4fb3dc11ca5e38ed9527e4eb534162c878ac048ac428abba29633b607d6258080a018d60336d1d9723e0d43567745464372e11f9f7a146abf163b0f43b338c0d827a00b44f7040c54b2bd63b0741062a96495dc4b92f7fec136c8effbc213cfffd02b80a0b81fcd189fbc66780a932367f874cbd4c85fed1fd37af109acc052f8ba80e51d80a0bbbcbe635193161c1e53442c06d0d14ed1af96af105a9fdd46a493195d62c1bba00050d231707972af60f163aa425401abb860f6c04b770fe0983e2f0f90ad6254a0804a9abf1b69bd16b1f6dc7cbbbf7592cd93fdf785b9cc61b01816583d9cd928a0ad85c29c3adbf2291cead9df6602560b93e485af7da680697ef9d1703f6f91cea040134ec8f47e1c599c0b7b4d237b8d3349f026fcfbacf885673438ffd2e9c72e80",
		"0xf851a01a5a4c757ab8653aad3a83cbb491142c776b91e252ffe849c253dab5d109819380808080808080a09fc998b2c0347275d9f98a9ac5bbb205387c04387672cda1a0360c4ef7cffb1a8080808080808080",
		"0xf86d9d205eb14112905ead3a2b16b733af37911d729cec51f559baf570332371b84df84b8087cc756b8ad2e000a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
	}
	address := common.HexToAddress("0x457a22804cf255ee8c1b7628601c5682b3d70c71")
	proof := [][]byte{}

	for _, s := range accountProof {
		bytes1, _ := hex.DecodeString(s[2:])
		proof = append(proof, bytes1)
	}

	stateRoot := "b666170175a47b5dde7514e732d935e58f7b379b0cf2774051c3fad5e9224000"
	root, err := hex.DecodeString(stateRoot)
	accountPath := crypto.Keccak256(address.Bytes())

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, root, keccak256(proof[0]), "Not Equal")
	var testData Proof

	err = json.Unmarshal([]byte(jsonString), &testData)
	assert.NoError(t, err, "Error in unmarshaling json string")

	proofStruct := EIP1186ProofResponse{
		Address:     common.HexToAddress(testData.Address),
		Balance:     uint256.MustFromHex(testData.Balance),
		CodeHash:    common.HexToHash(testData.CodeHash),
		Nonce:       hexutil.Uint64(hexutil.MustDecodeUint64(testData.Nonce)),
		StorageHash: common.HexToHash(testData.StorageHash),
		AccountProof: func() []hexutil.Bytes {
			var bytes []hexutil.Bytes
			for _, h := range testData.AccountProof {
				bytes = append(bytes, common.Hex2Bytes(h))
			}
			return bytes
		}(),
		StorageProof: []StorageProof{},
	}

	value, err := EncodeAccount(&proofStruct)
	assert.NoError(t, err, "Error in encoding account")

	isValid, err := VerifyProof(proof, root, accountPath, value)
	assert.NoError(t, err, "Found Error")
	assert.Equal(t, true, isValid, "ProofValidity not matching")
}