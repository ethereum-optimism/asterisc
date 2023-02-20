// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import "forge-std/Test.sol";
import "../src/Step.sol";

contract CounterTest is Test {
    Step public stepper;

    function setUp() public {
        stepper = new Step();
    }

    function testStep() public {
        out = stepper.step();
//        assertEq(out.abc, 1);
    }
}
