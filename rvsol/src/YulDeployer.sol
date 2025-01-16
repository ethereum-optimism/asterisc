pragma solidity 0.8.15;

import "forge-std/Test.sol";

contract YulDeployer is Test {
    /**
     * @notice Deploys a Yul contract and returns the address where the contract was deployed
     * @param fileName - The file name of the Yul contract (e.g., "Example.yul" becomes "Example")
     * @return deployedAddress - The address where the contract was deployed
     */
    function deployContract(string memory fileName) public returns (address) {
        string memory bashCommand = string.concat(
            'cast abi-encode "f(bytes)" $(solc --strict-assembly ./src/yul/',
            string.concat(fileName, ".yul --bin | grep '^[0-9a-fA-Z]*$')")
        );

        string[] memory inputs = new string[](3);
        inputs[0] = "bash";
        inputs[1] = "-c";
        inputs[2] = bashCommand;

        bytes memory bytecode = abi.decode(vm.ffi(inputs), (bytes));

        address deployedAddress;
        assembly {
            deployedAddress := create(0, add(bytecode, 0x20), mload(bytecode))
        }

        require(deployedAddress != address(0), "YulDeployer could not deploy contract");

        return deployedAddress;
    }
}
