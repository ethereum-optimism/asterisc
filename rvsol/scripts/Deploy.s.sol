// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "@forge-std/Script.sol";
import { console2 as console } from "@forge-std/console2.sol";

import { Chains } from "scripts/lib/Chains.sol";
import { Config } from "scripts/lib/Config.sol";
import { Deployer } from "scripts/lib/Deployer.sol";
import { RISCV } from "../src/RISCV.sol";

import { IBigStepper } from "@optimism/src/dispute/interfaces/IBigStepper.sol";
import { IPreimageOracle } from "@optimism/src/cannon/interfaces/IPreimageOracle.sol";
import { IDisputeGameFactory } from "@optimism/src/dispute/interfaces/IDisputeGameFactory.sol";
import { IDisputeGame } from "@optimism/src/dispute/interfaces/IDisputeGame.sol";
import { IFaultDisputeGame } from "@optimism/src/dispute/interfaces/IFaultDisputeGame.sol";
import { FaultDisputeGame } from "@optimism/src/dispute/FaultDisputeGame.sol";
import { IDelayedWETH } from "@optimism/src/dispute/interfaces/IDelayedWETH.sol";
import { IAnchorStateRegistry } from "@optimism/src/dispute/interfaces/IAnchorStateRegistry.sol";
import "@optimism/src/dispute/lib/Types.sol";

contract Deploy is Deployer {
    /// @notice FaultDisputeGameParams is a struct that contains the parameters necessary to call
    ///         the function _setFaultGameImplementation. This struct exists because the EVM needs
    ///         to finally adopt PUSHN and get rid of stack too deep once and for all.
    ///         Someday we will look back and laugh about stack too deep, today is not that day.
    struct FaultDisputeGameParams {
        IAnchorStateRegistry anchorStateRegistry;
        IDelayedWETH weth;
        GameType gameType;
        Claim absolutePrestate;
        IBigStepper faultVm;
        uint256 maxGameDepth;
        Duration maxClockDuration;
    }

    /// @notice Modifier that wraps a function in broadcasting.
    modifier broadcast() {
        vm.startBroadcast(msg.sender);
        _;
        vm.stopBroadcast();
    }

    /// @inheritdoc Deployer
    function name() public pure override returns (string memory name_) {
        name_ = "Deploy";
    }

    /// @notice The create2 salt used for deployment of the contract implementations.
    ///         Using this helps to reduce config across networks as the implementation
    ///         addresses will be the same across networks when deployed with create2.
    function _implSalt() internal view returns (bytes32) {
        return keccak256(bytes(Config.implSalt()));
    }

    function run() public {
        deployRiscv();
        setAsteriscFaultGameImplementation(false);
    }

    function runForDevnetAlloc() public {
        vm.loadAllocs(l1Allocfile);
        run();
        string memory path = vm.envOr(
            "STATE_DUMP_PATH", string.concat(vm.projectRoot(), "/", name(), "-", vm.toString(block.chainid), ".json")
        );
        vm.dumpState(path);
    }

    /// @notice Deploy RISCV
    function deployRiscv() public broadcast returns (address addr_) {
        console.log("Deploying RISCV implementation");
        addr_ = _deploy("RISCV", "RISCV", abi.encode(IPreimageOracle(mustGetAddress("PreimageOracle"))));
    }

    /// @notice Loads the riscv absolute prestate from the prestate-proof for devnets otherwise
    ///         from the config.
    function loadRiscvAbsolutePrestate() internal returns (Claim riscvAbsolutePrestate_) {
        if (block.chainid == Chains.LocalDevnet || block.chainid == Chains.GethDevnet) {
            // Fetch the absolute prestate dump
            string memory filePath = asteriscPrestatefile;
            string[] memory commands = new string[](3);
            commands[0] = "bash";
            commands[1] = "-c";
            commands[2] = string.concat("[[ -f ", filePath, " ]] && echo \"present\"");
            if (vm.ffi(commands).length == 0) {
                revert("Asterisc prestate dump not found, generate it with `make prestate` in the Asterisc root.");
            }
            commands[2] = string.concat("cat ", filePath, " | jq -r .pre");
            riscvAbsolutePrestate_ = Claim.wrap(abi.decode(vm.ffi(commands), (bytes32)));
            console.log(
                "[Asterisc Dispute Game] Using devnet RISCV Absolute prestate: %s",
                vm.toString(Claim.unwrap(riscvAbsolutePrestate_))
            );
        } else {
            revert("Currently Asterisc only supports local devnet");
            // TODO: Add Asterisc absolute prestate into OP stack deploy config
        }
    }

    /// @notice Sets the implementation for the given fault game type in the `DisputeGameFactory`.
    function setAsteriscFaultGameImplementation(bool _allowUpgrade) public broadcast {
        console.log("Setting Asterisc FaultDisputeGame implementation");
        IDisputeGameFactory factory = IDisputeGameFactory(mustGetAddress("DisputeGameFactoryProxy"));
        IDelayedWETH weth = IDelayedWETH(mustGetAddress("DelayedWETHProxy"));

        // Set the Asterisc FaultDisputeGame implementation in the factory.
        _setFaultGameImplementation({
            _factory: factory,
            _allowUpgrade: _allowUpgrade,
            _params: FaultDisputeGameParams({
                anchorStateRegistry: IAnchorStateRegistry(mustGetAddress("AnchorStateRegistryProxy")),
                weth: weth,
                gameType: GameTypes.ASTERISC,
                absolutePrestate: loadRiscvAbsolutePrestate(),
                faultVm: IBigStepper(mustGetAddress("RISCV")),
                maxGameDepth: cfg.faultGameMaxDepth(),
                maxClockDuration: Duration.wrap(uint64(cfg.faultGameMaxClockDuration()))
            })
        });
    }

    /// @notice Sets the implementation for the given fault game type in the `DisputeGameFactory`.
    function _setFaultGameImplementation(
        IDisputeGameFactory _factory,
        bool _allowUpgrade,
        FaultDisputeGameParams memory _params
    )
        internal
    {
        if (address(_factory.gameImpls(_params.gameType)) != address(0) && !_allowUpgrade) {
            console.log(
                "[WARN] DisputeGameFactoryProxy: `FaultDisputeGame` implementation already set for game type: %s",
                vm.toString(GameType.unwrap(_params.gameType))
            );
            return;
        }

        uint32 rawGameType = GameType.unwrap(_params.gameType);
        IDisputeGame dg = IDisputeGame(
            _deploy(
                "FaultDisputeGame",
                string.concat("FaultDisputeGame_", vm.toString(rawGameType)),
                abi.encode(
                    _params.gameType,
                    _params.absolutePrestate,
                    _params.maxGameDepth,
                    cfg.faultGameSplitDepth(),
                    cfg.faultGameClockExtension(),
                    _params.maxClockDuration,
                    _params.faultVm,
                    _params.weth,
                    _params.anchorStateRegistry,
                    cfg.l2ChainID()
                )
            )
        );

        _factory.setImplementation(_params.gameType, dg);

        console.log(
            "DisputeGameFactoryProxy: set `FaultDisputeGame` implementation (Backend: Asterisc | GameType: %s)",
            vm.toString(rawGameType)
        );
    }

    /// @notice Deploys a contract via CREATE2.
    /// @param _name The name of the contract.
    /// @param _constructorParams The constructor parameters.
    function _deploy(string memory _name, bytes memory _constructorParams) internal returns (address addr_) {
        return _deploy(_name, _name, _constructorParams);
    }

    /// @notice Deploys a contract via CREATE2.
    /// @param _name The name of the contract.
    /// @param _nickname The nickname of the contract.
    /// @param _constructorParams The constructor parameters.
    function _deploy(
        string memory _name,
        string memory _nickname,
        bytes memory _constructorParams
    )
        internal
        returns (address addr_)
    {
        console.log("Deploying %s", _nickname);
        bytes32 salt = _implSalt();
        bytes memory initCode = abi.encodePacked(vm.getCode(_name), _constructorParams);
        address preComputedAddress = vm.computeCreate2Address(salt, keccak256(initCode));
        require(preComputedAddress.code.length == 0, "Deploy: contract already deployed");
        assembly {
            addr_ := create2(0, add(initCode, 0x20), mload(initCode), salt)
        }
        require(addr_ != address(0), "deployment failed");
        save(_nickname, addr_);
        console.log("%s deployed at %s", _nickname, addr_);
    }
}
