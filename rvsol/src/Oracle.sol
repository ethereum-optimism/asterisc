// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;


contract Oracle {

    mapping (bytes32 => uint256) public preimageLengths;
    mapping (bytes32 => mapping(uint256 => bytes32)) preimageParts;
    mapping (bytes32 => mapping(uint256 => bool)) preimagePartOk;

    function readPreimage(bytes32 key, uint256 offset) public returns (bytes32 dat, uint256 datlen) {
        require(preimagePartOk[key][offset], "preimage must exist");
        datlen = 32;
        uint256 length = preimageLengths[key];
        if(offset + 32 >= length) {
            datlen = length - offset;
        }
        dat = preimageParts[key][offset];
    }

    function loadKeccak256PreimagePart(uint256 partOffset, bytes calldata preimage) public {
        uint256 size;
        bytes32 key;
        bytes32 part;
        assembly {
            // calldata layout: 4 (sel) + 0x20 (part offset) + 0x20 (start offset) + 0x20 (size) + preimage payload
            let startOffset := calldataload(0x24)
            if not(eq(startOffset, 0x44)) { // must always point to expected location of the size value.
                revert(0, 0)
            }
            size := calldataload(0x44)
            if iszero(lt(partOffset, size)) { // revert if part offset >= size (i.e. parts must be within bounds)
                revert(0, 0)
            }
            let ptr := 0x80 // we leave solidity slots 0x40 and 0x60 untouched, and everything after as scratch-memory.
            calldatacopy(ptr, 0x64, size) // copy preimage payload into memory so we can hash and read it.
            part := mload(add(ptr, partOffset))  // this will be zero-padded at the end, since memory at end is clean.
            let h := keccak256(ptr, size) // compute preimage keccak256 hash
            key := keccak256(h, 1)  // mix in the keccak256 pre-image type to get preimage key
        }
        preimagePartOk[key][partOffset] = true;
        preimageParts[key][partOffset] = part;
        preimageLengths[key] = size;
    }
}
