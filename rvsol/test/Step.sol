// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import "forge-std/Test.sol";
import "../src/Step.sol";

contract CounterTest is Test {
    Step public stepper;

    function setUp() public {
        stepper = new Step(address(42));
    }

    function testStep() public {
        bytes32 out = stepper.step(bytes32(uint256(0)), new bytes(0));
//        assertEq(out.abc, 1);
    }
}
