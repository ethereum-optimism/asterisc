pragma solidity 0.8.15;

import {Test} from "forge-std/Test.sol";
import {Step} from "src/Step.sol";
import {PreimageOracle} from "@optimism/packages/contracts-bedrock/src/cannon/PreimageOracle.sol";

contract Step_Test is Test {
    PreimageOracle internal oracle;
    Step internal step;

    function setUp() public {
        oracle = new PreimageOracle(0, 0, 0);
        step = new Step(oracle);
        vm.store(address(step), 0x0, bytes32(abi.encode(address(oracle))));
        vm.label(address(oracle), "PreimageOracle");
        vm.label(address(step), "Step");
    }
}
