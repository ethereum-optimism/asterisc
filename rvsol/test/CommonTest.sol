// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Test } from "forge-std/Test.sol";
import { Vm } from "forge-std/Vm.sol";

/// @title FFIInterface
/// @notice This contract is set into state using `etch` and therefore must not have constructor logic.
///         It also MUST be compiled with `0.8.15` because `vm.getDeployedCode` will break if there
///         are multiple artifacts for different compiler versions.
contract FFIInterface {
    Vm internal constant vm = Vm(0x7109709ECfa91a80626fF3989D68f67F5b1DD12D);

    function getAsteriscMemoryProof(uint64 pc, uint32 insn) external returns (bytes32, bytes memory) {
        string[] memory cmds = new string[](5);
        cmds[0] = "../rvgo/scripts/go-ffi/go-ffi";
        cmds[1] = "diff";
        cmds[2] = "asteriscMemoryProof";
        cmds[3] = vm.toString(pc);
        cmds[4] = vm.toString(insn);
        bytes memory result = vm.ffi(cmds);
        (bytes32 memRoot, bytes memory proof) = abi.decode(result, (bytes32, bytes));
        return (memRoot, proof);
    }

    function getAsteriscMemoryProof(
        uint64 pc,
        uint32 insn,
        uint64 memAddr,
        bytes32 memVal
    )
        external
        returns (bytes32, bytes memory)
    {
        string[] memory cmds = new string[](7);
        cmds[0] = "../rvgo/scripts/go-ffi/go-ffi";
        cmds[1] = "diff";
        cmds[2] = "asteriscMemoryProof";
        cmds[3] = vm.toString(pc);
        cmds[4] = vm.toString(insn);
        cmds[5] = vm.toString(memAddr);
        cmds[6] = vm.toString(memVal); // 0x prefixed hex string
        bytes memory result = vm.ffi(cmds);
        (bytes32 memRoot, bytes memory proof) = abi.decode(result, (bytes32, bytes));
        return (memRoot, proof);
    }
}

/// @title CommonTest
/// @dev An extension to `Test` that sets up the optimism smart contracts.
contract CommonTest is Test {
    FFIInterface constant ffi = FFIInterface(address(uint160(uint256(keccak256(abi.encode("optimism.ffi"))))));

    function setUp() public virtual {
        vm.etch(address(ffi), vm.getDeployedCode("CommonTest.sol:FFIInterface"));
        vm.label(address(ffi), "FFIInterface");
    }
}
