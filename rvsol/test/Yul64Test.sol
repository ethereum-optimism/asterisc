pragma solidity 0.8.15;

import "forge-std/Test.sol";
import "../src/YulDeployer.sol";
import { console } from "forge-std/console.sol";

interface Yul64 { }

// Run with: forge test --match-path ./test/Yul64Test.sol -vvvvv
contract Yul64Test is Test {
    YulDeployer yulDeployer = new YulDeployer();
    Yul64Test yul64;
    mapping(string => string) public slowToEVM;
    string[11] slowFunctions;

    /// @notice deploys yul wrapper and sets up string for slow (golang) and mapping for solidity
    function setUp() public {
        yul64 = Yul64Test(yulDeployer.deployContract("Yul64"));

        slowFunctions[0] = "ParseTypeI";
        slowToEVM[slowFunctions[0]] = "parseImmTypeI(uint64)";

        slowFunctions[1] = "ParseTypeS";
        slowToEVM[slowFunctions[1]] = "parseImmTypeS(uint64)";

        slowFunctions[2] = "ParseTypeB";
        slowToEVM[slowFunctions[2]] = "parseImmTypeB(uint64)";

        slowFunctions[3] = "ParseTypeU";
        slowToEVM[slowFunctions[3]] = "parseImmTypeU(uint64)";

        slowFunctions[4] = "ParseTypeJ";
        slowToEVM[slowFunctions[4]] = "parseImmTypeJ(uint64)";

        slowFunctions[5] = "ParseOpcode";
        slowToEVM[slowFunctions[5]] = "parseOpcode(uint64)";

        slowFunctions[6] = "ParseRd";
        slowToEVM[slowFunctions[6]] = "parseRd(uint64)";

        slowFunctions[7] = "ParseFunct3";
        slowToEVM[slowFunctions[7]] = "parseFunct3(uint64)";

        slowFunctions[8] = "ParseRs1";
        slowToEVM[slowFunctions[8]] = "parseRs1(uint64)";

        slowFunctions[9] = "ParseRs2";
        slowToEVM[slowFunctions[9]] = "parseRs2(uint64)";

        slowFunctions[10] = "ParseFunct7";
        slowToEVM[slowFunctions[10]] = "parseFunct7(uint64)";
    }

    function testFuzz_parseTypeI(uint32 input) public {
        runDiffTest(slowFunctions[0], input);
    }

    function testFuzz_parseTypeS(uint32 input) public {
        runDiffTest(slowFunctions[1], input);
    }

    function testFuzz_parseTypeB(uint32 input) public {
        runDiffTest(slowFunctions[2], input);
    }

    function testFuzz_parseTypeU(uint32 input) public {
        runDiffTest(slowFunctions[3], input);
    }

    function testFuzz_parseTypeJ(uint32 input) public {
        runDiffTest(slowFunctions[4], input);
    }

    function testFuzz_parseOpcode(uint32 input) public {
        runDiffTest(slowFunctions[5], input);
    }

    function testFuzz_parseRd(uint32 input) public {
        runDiffTest(slowFunctions[6], input);
    }

    function testFuzz_parseFunct3(uint32 input) public {
        runDiffTest(slowFunctions[7], input);
    }

    function testFuzz_parseRs1(uint32 input) public {
        runDiffTest(slowFunctions[8], input);
    }

    function testFuzz_parseRs2(uint32 input) public {
        runDiffTest(slowFunctions[9], input);
    }

    function testFuzz_parseFunct7(uint32 input) public {
        runDiffTest(slowFunctions[10], input);
    }

    // Helper functions

    /// @notice Executes evm code then ffi code and checks consistent behavior
    /// @dev Used by all testFuzz_ functions given funcToCall
    /// @param funcToCall Defines function under test
    /// @param input Defines uint32 input to represent instruction
    function runDiffTest(string memory funcToCall, uint64 input) private {
        (bool evmSuccess, bytes memory evmOutput) = executeEVM(funcToCall, input);
        bytes memory ffiOutput = executeFFI(funcToCall, input);
        assertConsistent(evmSuccess, evmOutput, ffiOutput);
    }

    /// @notice Generates calldata given signature and input, and calls yul64 wrapper
    /// @dev Helper function is used by all testFuzz_functions in runDiffTest
    /// @param funcToCall key to slowToEVM mapping to define which function to test
    /// @param input Defines uint32 input to represent instruction
    function executeEVM(
        string memory funcToCall,
        uint64 input
    )
        private
        returns (bool evmSuccess, bytes memory evmOutput)
    {
        // generate calldata for EVM call
        bytes memory callDataBytes = abi.encodeWithSignature(slowToEVM[funcToCall], input);
        (evmSuccess, evmOutput) = address(yul64).call(callDataBytes);
    }
    /// @notice Generates ffi command to invoke ./rvsol/test/slow with arguments
    /// @dev Required to build diff.go prior to use
    /// @param funcToCall key to slowToEVM mapping to define which function to test
    /// @param input Defines uint32 input to represent instruction

    function executeFFI(string memory funcToCall, uint64 input) private returns (bytes memory ffiOutput) {
        string[] memory inputs = generateCommand(funcToCall, input);
        ffiOutput = vm.ffi(inputs);
    }

    /// @notice Asserts false if evm and ffi results are unaligned
    /// @param evmSuccess True if EVM call was successful (did not revert)
    /// @param evmOutput Result of EVM parse function
    /// @param ffiOutput Result of slow parse function
    function assertConsistent(bool evmSuccess, bytes memory evmOutput, bytes memory ffiOutput) private pure {
        console.logBool(didPanic(ffiOutput));
        if (didPanic(ffiOutput)) {
            // if slow vm determined there was an invalid value
            console.logString("SLOW VM // invalid value");
            if (!evmSuccess) {
                // and so did EVM, consider non-deviating behaviour
                console.logString("EVM // expected revert");
            } else {
                // EVM should have been successful
                console.logString("! EVM FAIL ! // should have failed");
                assert(false);
            }
        } else {
            console.logString("SLOW VM // successful");
            if (!evmSuccess) {
                console.log("! EVM FAIL ! // EVM failed");
                assert(false);
            } else {
                assertEq(ffiOutput, evmOutput);
            }
        }
    }

    /// @notice Generates ffi command to run slow.
    /// @dev This requires slow to print nothing else when run except result. If changed, ensure script does not print
    /// with a newline
    /// @param functionToCall Defines function to call with ffi
    /// @param numberInput Specifies input from fuzzer to call with
    /// @return inputs list of strings with the full bash command to run ffi
    function generateCommand(string memory functionToCall, uint64 numberInput) private pure returns (string[] memory) {
        string memory bashCommand = string.concat(
            "../rvgo/scripts/parse-diff-ffi/slow -fuzz=", functionToCall, " -number=", (vm.toString(numberInput))
        );
        string[] memory inputs = new string[](3);
        inputs[0] = "bash";
        inputs[1] = "-c";
        inputs[2] = bashCommand;
        return inputs;
    }

    /// @notice Returns true if didPanic panic string
    function didPanic(bytes memory whatBytes) private pure returns (bool found) {
        string memory err =
            hex"70616E69633A20696E76616C69642076616C75650A0A676F726F7574696E652031205B72756E6E696E675D3A0A6D61696E2E6D61696E28290A";
        return keccak256(abi.encodePacked(string(whatBytes))) == keccak256(abi.encodePacked(err));
    }
}
