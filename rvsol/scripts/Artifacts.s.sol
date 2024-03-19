// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";
import { Vm } from "forge-std/Vm.sol";
import { Config } from "scripts/Config.sol";

/// @notice Represents a deployment. Is serialized to JSON as a key/value
///         pair. Can be accessed from within scripts.
struct Deployment {
    string name;
    address payable addr;
}

/// @title Artifacts
abstract contract Artifacts {
    /// @notice Foundry cheatcode VM.
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    /// @notice Error for when attempting to fetch a deployment and it does not exist
    error DeploymentDoesNotExist(string);
    /// @notice Error for when trying to save an invalid deployment
    error InvalidDeployment(string);

    /// @notice The set of chain deployments that have been already deployed.
    mapping(string => Deployment) internal _chainDeployments;
    /// @notice The set of deployments that have been done during execution.
    mapping(string => Deployment) internal _namedDeployments;
    /// @notice The path to the deployment artifact that is being written to.
    string internal deploymentOutfile;

    /// @notice Accepts a filepath and then ensures that the directory
    ///         exists for the file to live in.
    function ensurePath(string memory _path) internal {
        (, bytes memory returndata) =
            address(vm).call(abi.encodeWithSignature("split(string,string)", _path, string("/")));
        string[] memory outputs = abi.decode(returndata, (string[]));

        string memory path = "";
        for (uint256 i = 0; i < outputs.length - 1; i++) {
            path = string.concat(path, outputs[i], "/");
        }
        vm.createDir(path, true);
    }

    /// @notice Setup function. The arguments here
    function setUp() public virtual {
        deploymentOutfile = Config.deploymentOutfile();
        console.log("Writing artifact to %s", deploymentOutfile);
        ensurePath(deploymentOutfile);

        // Load addresses from a JSON file if the TARGET_L2_DEPLOYMENT_FILE environment variable
        // is set. Great for loading addresses from `superchain-registry`.
        string memory addresses = Config.chainDeploymentFile();
        if (bytes(addresses).length > 0) {
            console.log("Loading chain addresses from %s", addresses);
            _loadChainAddresses(addresses);
        }
    }

    /// @notice Populates the addresses to be used in a script based on a JSON file.
    ///         The JSON key is the name of the contract and the value is an address.
    function _loadChainAddresses(string memory _path) internal {
        string[] memory commands = new string[](3);
        commands[0] = "bash";
        commands[1] = "-c";
        commands[2] = string.concat("jq -cr < ", _path);
        string memory json = string(vm.ffi(commands));
        string[] memory keys = vm.parseJsonKeys(json, "");
        for (uint256 i; i < keys.length; i++) {
            string memory key = keys[i];
            address addr = stdJson.readAddress(json, string.concat("$.", key));
            Deployment memory deployment = Deployment({ name: key, addr: payable(addr) });
            _chainDeployments[key] = deployment;
            console.log("Loading %s: %s", key, addr);
        }
    }

    /// @notice Appends a deployment to disk as a JSON deploy artifact.
    /// @param _name The name of the deployment.
    /// @param _deployed The address of the deployment.
    function save(string memory _name, address _deployed) public {
        if (bytes(_name).length == 0) {
            revert InvalidDeployment("EmptyName");
        }
        if (bytes(_namedDeployments[_name].name).length > 0) {
            revert InvalidDeployment("AlreadyExists");
        }

        console.log("Saving %s: %s", _name, _deployed);
        Deployment memory deployment = Deployment({ name: _name, addr: payable(_deployed) });
        _namedDeployments[_name] = deployment;
        _appendDeployment(_name, _deployed);
    }

    /// @notice Adds a deployment to the temp deployments file
    function _appendDeployment(string memory _name, address _deployed) internal {
        vm.writeJson({ json: stdJson.serialize("", _name, _deployed), path: deploymentOutfile });
    }

    /// @notice Returns a deployment that is suitable to be used to interact with contracts.
    /// @param _name The name of the deployment.
    /// @return The deployment.
    function getChainDeployment(string memory _name) public view returns (Deployment memory) {
        return _chainDeployments[_name];
    }

    /// @notice Returns the address of a deployment. Also handles the predeploys.
    /// @param _name The name of the deployment.
    /// @return The address of the deployment. May be `address(0)` if the deployment does not
    ///         exist.
    function getChainAddress(string memory _name) public view returns (address payable) {
        Deployment memory existing = _chainDeployments[_name];
        if (existing.addr != address(0)) {
            if (bytes(existing.name).length == 0) {
                return payable(address(0));
            }
            return existing.addr;
        }

        return payable(address(0));
    }

    /// @notice Returns the address of a deployment and reverts if the deployment
    ///         does not exist.
    /// @return The address of the deployment.
    function mustGetChainAddress(string memory _name) public view returns (address payable) {
        address addr = getChainAddress(_name);
        if (addr == address(0)) {
            revert DeploymentDoesNotExist(_name);
        }
        return payable(addr);
    }

    /// @notice Returns the address of a deployment. Also handles the predeploys.
    /// @param _name The name of the deployment.
    /// @return The address of the deployment. May be `address(0)` if the deployment does not
    ///         exist.
    function getAddress(string memory _name) public view returns (address payable) {
        Deployment memory existing = _namedDeployments[_name];
        if (existing.addr != address(0)) {
            if (bytes(existing.name).length == 0) {
                return payable(address(0));
            }
            return existing.addr;
        }

        return payable(address(0));
    }

    /// @notice Returns the address of a deployment and reverts if the deployment
    ///         does not exist.
    /// @return The address of the deployment.
    function mustGetAddress(string memory _name) public view returns (address payable) {
        address addr = getAddress(_name);
        if (addr == address(0)) {
            revert DeploymentDoesNotExist(_name);
        }
        return payable(addr);
    }
}
