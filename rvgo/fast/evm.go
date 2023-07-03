package fast

import "github.com/ethereum/go-ethereum/crypto"

var (
	StepBytes4                      = crypto.Keccak256([]byte("Step(bytes,bytes)"))[:4]
	CheatBytes4                     = crypto.Keccak256([]byte("cheat(uint256,bytes32,bytes32,uint256)"))[:4]
	LoadKeccak256PreimagePartBytes4 = crypto.Keccak256([]byte("loadKeccak256PreimagePart(uint256,bytes)"))[:4]
)
