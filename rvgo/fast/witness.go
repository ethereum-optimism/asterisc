package fast

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/asterisc/rvgo/bindings"
	preimage "github.com/ethereum-optimism/optimism/op-preimage"
	"github.com/ethereum/go-ethereum/common"
)

type LocalContext common.Hash

type StepWitness struct {
	// encoded state witness
	State []byte

	MemProof []byte

	PreimageKey    [32]byte // zeroed when no pre-image is accessed
	PreimageValue  []byte   // including the 8-byte length prefix
	PreimageOffset uint64
}

func (wit *StepWitness) EncodeStepInput(localContext LocalContext) ([]byte, error) {
	abi, err := bindings.RISCVMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	input, err := abi.Pack("step", wit.State, wit.MemProof, localContext)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func (wit *StepWitness) HasPreimage() bool {
	return wit.PreimageKey != ([32]byte{})
}

func (wit *StepWitness) EncodePreimageOracleInput(localContext LocalContext) ([]byte, error) {
	if wit.PreimageKey == ([32]byte{}) {
		return nil, errors.New("cannot encode pre-image oracle input, witness has no pre-image to proof")
	}

	preimageAbi, err := bindings.PreimageOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	switch preimage.KeyType(wit.PreimageKey[0]) {
	case preimage.LocalKeyType:
		if len(wit.PreimageValue) > 32+8 {
			return nil, fmt.Errorf("local pre-image exceeds maximum size of 32 bytes with key 0x%x", wit.PreimageKey)
		}
		preimagePart := wit.PreimageValue[8:]
		var tmp [32]byte
		copy(tmp[:], preimagePart)
		input, err := preimageAbi.Pack("loadLocalData",
			new(big.Int).SetBytes(wit.PreimageKey[1:]),
			localContext,
			tmp,
			new(big.Int).SetUint64(uint64(len(preimagePart))),
			new(big.Int).SetUint64(uint64(wit.PreimageOffset)),
		)
		if err != nil {
			return nil, err
		}
		return input, nil
	case preimage.Keccak256KeyType:
		input, err := preimageAbi.Pack(
			"loadKeccak256PreimagePart",
			new(big.Int).SetUint64(uint64(wit.PreimageOffset)),
			wit.PreimageValue[8:])
		if err != nil {
			return nil, err
		}
		return input, nil
	default:
		return nil, fmt.Errorf("unsupported pre-image type %d, cannot prepare preimage with key %x offset %d for oracle",
			wit.PreimageKey[0], wit.PreimageKey, wit.PreimageOffset)
	}
}
