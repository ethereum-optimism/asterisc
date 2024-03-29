package fast

import "github.com/ethereum/go-ethereum/crypto"

var (
	StepBytes4                      = crypto.Keccak256([]byte("step(bytes,bytes,bytes32)"))[:4]
	LoadKeccak256PreimagePartBytes4 = crypto.Keccak256([]byte("loadKeccak256PreimagePart(uint256,bytes)"))[:4]
	LoadLocalDataBytes4             = crypto.Keccak256([]byte("loadLocalData(uint256,bytes32,bytes32,uint256,uint256)"))[:4]
)
