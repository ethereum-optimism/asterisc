// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Vm } from "forge-std/Vm.sol";

import { Chains } from "scripts/Chains.sol";

/// @title Config
library Config {
    /// @notice Foundry cheatcode VM.
    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    /// @notice Returns the path on the local filesystem where the deployment artifact is
    ///         written to disk after doing a deployment.
    function deploymentOutfile() internal view returns (string memory _env) {
        _env = vm.envOr(
            "DEPLOYMENT_OUTFILE", string.concat(vm.projectRoot(), "/deployments/", _getDeploymentContext(), "/.deploy")
        );
    }

    /// @notice Returns the path on the local filesystem where the deploy config is
    function deployConfigPath() internal view returns (string memory _env) {
        _env = vm.envOr(
            "TARGET_L2_DEPLOY_CONFIG",
            string.concat(vm.projectRoot(), "/deploy-config/", _getDeploymentContext(), ".json")
        );
    }

    /// @notice Returns the path on the local filesystem where the target chain deployment artifact is written.
    function chainDeploymentFile() internal view returns (string memory _env) {
        _env = vm.envOr("TARGET_L2_DEPLOYMENT_FILE", string("./.deploy"));
    }

    function chainL1AllocPath() internal view returns (string memory _env) {
        _env = vm.envOr("TARGET_L1_ALLOC", string("./allocs-L1.json"));
    }

    function asteriscPrestatePath() internal view returns (string memory _env) {
        _env = vm.envOr("ASTERISC_PRESTATE", string.concat(vm.projectRoot(), "/../rvgo/bin/prestate-proof.json"));
    }

    /// @notice Returns the chainid from the EVM context or the value of the CHAIN_ID env var as
    ///         an override.
    function chainID() internal view returns (uint256 _env) {
        _env = vm.envOr("CHAIN_ID", block.chainid);
    }

    /// @notice Returns the deployment context which was only useful in the hardhat deploy style
    ///         of deployments. It is now DEPRECATED and will be removed in the future.
    function deploymentContext() internal view returns (string memory _env) {
        _env = vm.envOr("DEPLOYMENT_CONTEXT", string(""));
    }

    /// @notice The CREATE2 salt to be used when deploying the implementations.
    function implSalt() internal view returns (string memory _env) {
        _env = vm.envOr("IMPL_SALT", string("ethers phoenix"));
    }

    /// @notice The context of the deployment is used to namespace the artifacts.
    ///         An unknown context will use the chainid as the context name.
    ///         This is legacy code and should be removed in the future.
    function _getDeploymentContext() private view returns (string memory) {
        string memory context = deploymentContext();
        if (bytes(context).length > 0) {
            return context;
        }

        uint256 chainid = Config.chainID();
        if (chainid == Chains.Mainnet) {
            return "mainnet";
        } else if (chainid == Chains.Goerli) {
            return "goerli";
        } else if (chainid == Chains.OPGoerli) {
            return "optimism-goerli";
        } else if (chainid == Chains.OPMainnet) {
            return "optimism-mainnet";
        } else if (chainid == Chains.LocalDevnet || chainid == Chains.GethDevnet) {
            return "devnetL1";
        } else if (chainid == Chains.Hardhat) {
            return "hardhat";
        } else if (chainid == Chains.Sepolia) {
            return "sepolia";
        } else if (chainid == Chains.OPSepolia) {
            return "optimism-sepolia";
        } else {
            return vm.toString(chainid);
        }
    }
}
