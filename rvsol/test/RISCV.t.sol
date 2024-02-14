pragma solidity 0.8.15;

import {Test} from "forge-std/Test.sol";
import {RISCV} from "src/RISCV.sol";
import {PreimageOracle} from "@optimism/src/cannon/PreimageOracle.sol";

contract RISCV_Test is Test {
    RISCV internal riscv;
    PreimageOracle internal oracle;

    function setUp() public {
        oracle = new PreimageOracle(0, 0, 0);
        riscv = new RISCV(oracle);
        vm.store(address(riscv), 0x0, bytes32(abi.encode(address(oracle))));
        vm.label(address(oracle), "PreimageOracle");
        vm.label(address(riscv), "RISCV");
    }
}
