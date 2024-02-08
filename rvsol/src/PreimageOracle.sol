// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import {IPreimageOracle} from "./interfaces/IPreimageOracle.sol";
import {PreimageKeyLib} from "./PreimageKeyLib.sol";

contract PreimageOracle is IPreimageOracle {
    mapping(bytes32 => uint256) public preimageLengths;
    mapping(bytes32 => mapping(uint256 => bytes32)) public preimageParts;
    mapping(bytes32 => mapping(uint256 => bool)) public preimagePartOk;

    function readPreimage(bytes32 _key, uint256 _offset) external view returns (bytes32 dat_, uint256 datLen_) {
        require(preimagePartOk[_key][_offset], "pre-image must exist");

        // Calculate the length of the pre-image data
        // Add 8 for the length-prefix part
        datLen_ = 32;
        uint256 length = preimageLengths[_key];
        if (_offset + 32 >= length + 8) {
            datLen_ = length + 8 - _offset;
        }

        // Retrieve the pre-image data
        dat_ = preimageParts[_key][_offset];
    }

    // TODO(CLI-4104):
    // we need to mix-in the ID of the dispute for local-type keys to avoid collisions,
    // and restrict local pre-image insertion to the dispute-managing contract.
    // For now we permit anyone to write any pre-image unchecked, to make testing easy.
    // This method is DANGEROUS. And NOT FOR PRODUCTION.
    function cheat(uint256 partOffset, bytes32 key, bytes32 part, uint256 size) external {
        preimagePartOk[key][partOffset] = true;
        preimageParts[key][partOffset] = part;
        preimageLengths[key] = size;
    }

    // temporary method for localization. Will be removed to PreimageKeyLib.sol
    function localize(bytes32 _key, bytes32 _localContext) internal view returns (bytes32 localizedKey_) {
        assembly {
            // Grab the current free memory pointer to restore later.
            let ptr := mload(0x40)
            // Store the local data key and caller next to each other in memory for hashing.
            mstore(0, _key)
            mstore(0x20, caller())
            mstore(0x40, _localContext)
            // Localize the key with the above `localize` operation.
            localizedKey_ := or(and(keccak256(0, 0x60), not(shl(248, 0xFF))), shl(248, 1))
            // Restore the free memory pointer.
            mstore(0x40, ptr)
        }
    }

    // temporary method for localization. Will be removed to PreimageKeyLib.sol
    function cheatLocalKey(uint256 partOffset, bytes32 key, bytes32 part, uint256 size, bytes32 localContext)
        external
    {
        // sanity check key is local key using prefix
        require(uint8(key[0]) == 1, "must be used for local key");

        bytes32 localizedKey = PreimageKeyLib.localize(key, localContext);
        preimagePartOk[localizedKey][partOffset] = true;
        preimageParts[localizedKey][partOffset] = part;
        preimageLengths[localizedKey] = size;
    }

    function loadLocalData(uint256 _ident, bytes32 _localContext, bytes32 _word, uint256 _size, uint256 _partOffset)
        external
        returns (bytes32 key_)
    {
        // TODO: implement me
    }

    // loadKeccak256PreimagePart prepares the pre-image to be read by keccak256 key,
    // starting at the given offset, up to 32 bytes (clipped at preimage length, if out of data).
    function loadKeccak256PreimagePart(uint256 _partOffset, bytes calldata _preimage) external {
        uint256 size;
        bytes32 key;
        bytes32 part;
        assembly {
            // len(sig) + len(partOffset) + len(preimage offset) = 4 + 32 + 32 = 0x44
            size := calldataload(0x44)

            // revert if part offset > size+8 (i.e. parts must be within bounds)
            if gt(_partOffset, add(size, 8)) {
                // Store "PartOffsetOOB()"
                mstore(0, 0xfe254987)
                // Revert with "PartOffsetOOB()"
                revert(0x1c, 4)
            }
            // we leave solidity slots 0x40 and 0x60 untouched,
            // and everything after as scratch-memory.
            let ptr := 0x80
            // put size as big-endian uint64 at start of pre-image
            mstore(ptr, shl(192, size))
            ptr := add(ptr, 8)
            // copy preimage payload into memory so we can hash and read it.
            calldatacopy(ptr, _preimage.offset, size)
            // Note that it includes the 8-byte big-endian uint64 length prefix.
            // this will be zero-padded at the end, since memory at end is clean.
            part := mload(add(sub(ptr, 8), _partOffset))
            let h := keccak256(ptr, size) // compute preimage keccak256 hash
            // mask out prefix byte, replace with type 2 byte
            key := or(and(h, not(shl(248, 0xFF))), shl(248, 2))
        }
        preimagePartOk[key][_partOffset] = true;
        preimageParts[key][_partOffset] = part;
        preimageLengths[key] = size;
    }

    function loadSha256PreimagePart(uint256 _partOffset, bytes calldata _preimage) external {
        // TODO: fetch diff from cannon. Currently for only satisfying interface
    }
}
