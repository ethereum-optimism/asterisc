// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Config } from "scripts/Config.sol";
import { Deployer } from "scripts/Deployer.sol";
import { RISCV } from "../src/RISCV.sol";
import { IPreimageOracle } from "@optimism/src/cannon/interfaces/IPreimageOracle.sol";
import { DisputeGameFactory } from "@optimism/src/dispute/DisputeGameFactory.sol";
import { DelayedWETH } from "@optimism/src/dispute/weth/DelayedWETH.sol";
import { AnchorStateRegistry } from "@optimism/src/dispute/AnchorStateRegistry.sol";
import { PreimageOracle } from "@optimism/src/cannon/PreimageOracle.sol";
import { Types } from "@optimism/scripts/Types.sol";
import { ProxyAdmin } from "@optimism/src/universal/ProxyAdmin.sol";
import { AddressManager } from "@optimism/src/legacy/AddressManager.sol";
import { Proxy } from "@optimism/src/universal/Proxy.sol";
import { EIP1967Helper } from "@optimism/test/mocks/EIP1967Helper.sol";
import { FaultDisputeGame } from "@optimism/src/dispute/FaultDisputeGame.sol";
import { Chains } from "@optimism/scripts/Chains.sol";
import { IBigStepper } from "@optimism/src/dispute/interfaces/IBigStepper.sol";
import "@optimism/src/dispute/lib/Types.sol";
import { console2 as console } from "forge-std/console2.sol";
import { StdAssertions } from "forge-std/StdAssertions.sol";

contract Deploy is Deployer, StdAssertions {
    /// @notice Modifier that wraps a function in broadcasting.
    modifier broadcast() {
        vm.startBroadcast(msg.sender);
        _;
        vm.stopBroadcast();
    }

    /// @inheritdoc Deployer
    function name() public pure override returns (string memory name_) {
        name_ = "Deploy_Stage_1_4";
    }

    /// @notice Deploy all of the L1 contracts necessary for a Stage 1.4 Deployment.
    ///         Intentionally not using Safe contracts for brevity.
    //          Do not need to deploy AddressManager because no legacy contracts deployed.
    function run() public {
        deployProxyAdmin();

        deployProxies();
        deployImplementations();
        initializeImplementations();

        setAsteriscFaultGameImplementation({ _allowUpgrade: false });

        postDeployAssertions();
        printConfigReview();
    }

    /// @notice The create2 salt used for deployment of the contract implementations.
    ///         Using this helps to reduce config across networks as the implementation
    ///         addresses will be the same across networks when deployed with create2.
    function _implSalt() internal view returns (bytes32) {
        return keccak256(bytes(Config.implSalt()));
    }

    /// @notice Deploy RISCV
    function deployRiscv() public broadcast returns (address addr_) {
        console.log("Deploying RISCV implementation");
        RISCV riscv = new RISCV{ salt: _implSalt() }(IPreimageOracle(mustGetChainAddress("PreimageOracle")));
        save("RISCV", address(riscv));
        console.log("RISCV deployed at %s", address(riscv));
        addr_ = address(riscv);
    }

    /// @notice Deploy all of the implementations
    function deployImplementations() public {
        console.log("Deploying implementations");
        deployDisputeGameFactory();
        deployRiscv();
        deployAnchorStateRegistry();
    }

    /// @notice Deploy all of the proxies
    function deployProxies() public {
        console.log("Deploying proxies");
        deployERC1967Proxy("DisputeGameFactoryProxy");
        deployERC1967Proxy("AnchorStateRegistryProxy");
    }

    /// @notice Deploy the ProxyAdmin
    function deployProxyAdmin() public broadcast returns (address addr_) {
        console.log("Deploying ProxyAdmin");
        ProxyAdmin admin = new ProxyAdmin({ _owner: msg.sender });
        require(admin.owner() == msg.sender);

        save("ProxyAdmin", address(admin));
        console.log("ProxyAdmin deployed at %s", address(admin));
        addr_ = address(admin);
    }

    /// @notice Deploy the DisputeGameFactory
    function deployDisputeGameFactory() public broadcast returns (address addr_) {
        console.log("Deploying DisputeGameFactory implementation");
        DisputeGameFactory disputeGameFactory = new DisputeGameFactory{ salt: _implSalt() }();
        save("DisputeGameFactory", address(disputeGameFactory));
        console.log("DisputeGameFactory deployed at %s", address(disputeGameFactory));

        // Check that the contract is initialized
        assertSlotValueIsOne({ _contractAddress: address(disputeGameFactory), _slot: 0, _offset: 0 });
        require(disputeGameFactory.owner() == address(0));

        addr_ = address(disputeGameFactory);
    }

    /// @notice Deploy the AnchorStateRegistry
    function deployAnchorStateRegistry() public broadcast returns (address addr_) {
        console.log("Deploying AnchorStateRegistry implementation");
        AnchorStateRegistry anchorStateRegistry =
            new AnchorStateRegistry{ salt: _implSalt() }(DisputeGameFactory(mustGetAddress("DisputeGameFactory")));
        save("AnchorStateRegistry", address(anchorStateRegistry));
        console.log("AnchorStateRegistry deployed at %s", address(anchorStateRegistry));

        addr_ = address(anchorStateRegistry);
    }

    /// @notice Initialize all of the implementations
    function initializeImplementations() public {
        console.log("Initializing implementations");
        initializeDisputeGameFactory();
        initializeAnchorStateRegistry();
    }

    /// @notice Initialize the DisputeGameFactory
    function initializeDisputeGameFactory() public broadcast {
        console.log("Upgrading and initializing DisputeGameFactory");
        address disputeGameFactoryProxy = mustGetAddress("DisputeGameFactoryProxy");
        address disputeGameFactory = mustGetAddress("DisputeGameFactory");

        _upgradeAndCall({
            _proxy: payable(disputeGameFactoryProxy),
            _implementation: disputeGameFactory,
            _innerCallData: abi.encodeCall(DisputeGameFactory.initialize, (msg.sender))
        });

        string memory version = DisputeGameFactory(disputeGameFactoryProxy).version();
        console.log("DisputeGameFactory version: %s", version);

        // Check that the contract is initialized
        assertSlotValueIsOne({ _contractAddress: address(disputeGameFactoryProxy), _slot: 0, _offset: 0 });
        require(DisputeGameFactory(disputeGameFactoryProxy).owner() == msg.sender);
    }

    // @notice Initialize the AnchorStateRegistry
    //         Only initialize anchors for asterisc
    function initializeAnchorStateRegistry() public broadcast {
        console.log("Upgrading and initializing AnchorStateRegistry");
        address anchorStateRegistryProxy = mustGetAddress("AnchorStateRegistryProxy");
        address anchorStateRegistry = mustGetAddress("AnchorStateRegistry");

        AnchorStateRegistry.StartingAnchorRoot[] memory roots = new AnchorStateRegistry.StartingAnchorRoot[](1);
        roots[0] = AnchorStateRegistry.StartingAnchorRoot({
            gameType: GameTypes.ASTERISC,
            outputRoot: OutputRoot({
                root: Hash.wrap(cfg.faultGameGenesisOutputRoot()),
                l2BlockNumber: cfg.faultGameGenesisBlock()
            })
        });

        _upgradeAndCall({
            _proxy: payable(anchorStateRegistryProxy),
            _implementation: anchorStateRegistry,
            _innerCallData: abi.encodeCall(AnchorStateRegistry.initialize, (roots))
        });

        string memory version = AnchorStateRegistry(payable(anchorStateRegistryProxy)).version();
        console.log("AnchorStateRegistry version: %s", version);
    }

    /// @dev Asserts that for a given contract the value of a storage slot at an offset is 1.
    ///      From ChainAssertions.sol
    function assertSlotValueIsOne(address _contractAddress, uint256 _slot, uint256 _offset) internal view {
        bytes32 slotVal = vm.load(_contractAddress, bytes32(_slot));
        require(
            uint8((uint256(slotVal) >> (_offset * 8)) & 0xFF) == uint8(1),
            "Storage value is not 1 at the given slot and offset"
        );
    }

    /// @notice Call the Proxy Admin's upgrade and call method
    function _upgradeAndCall(address _proxy, address _implementation, bytes memory _innerCallData) internal {
        ProxyAdmin proxyAdmin = ProxyAdmin(mustGetAddress("ProxyAdmin"));
        proxyAdmin.upgradeAndCall(payable(_proxy), _implementation, _innerCallData);
    }

    /// @notice Deploys an ERC1967Proxy contract with the ProxyAdmin as the owner.
    /// @param _name The name of the proxy contract to be deployed.
    /// @return addr_ The address of the deployed proxy contract.
    function deployERC1967Proxy(string memory _name) public returns (address addr_) {
        addr_ = deployERC1967ProxyWithOwner(_name, mustGetAddress("ProxyAdmin"));
    }

    /// @notice Deploys an ERC1967Proxy contract with a specified owner.
    /// @param _name The name of the proxy contract to be deployed.
    /// @param _proxyOwner The address of the owner of the proxy contract.
    /// @return addr_ The address of the deployed proxy contract.
    function deployERC1967ProxyWithOwner(
        string memory _name,
        address _proxyOwner
    )
        public
        broadcast
        returns (address addr_)
    {
        console.log(string.concat("Deploying ERC1967 proxy for ", _name));
        Proxy proxy = new Proxy({ _admin: _proxyOwner });

        require(EIP1967Helper.getAdmin(address(proxy)) == _proxyOwner);

        save(_name, address(proxy));
        console.log("   at %s", address(proxy));
        addr_ = address(proxy);
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
            console.log(
                "[Asterisc Dispute Game] Using absolute prestate from config: %x", cfg.faultGameAbsolutePrestate()
            );
            riscvAbsolutePrestate_ = Claim.wrap(bytes32(cfg.faultGameAbsolutePrestate()));
        }
    }

    /// @notice Sets the implementation for the given fault game type in the `DisputeGameFactory`.
    function setAsteriscFaultGameImplementation(bool _allowUpgrade) public broadcast {
        console.log("Setting Asterisc FaultDisputeGame implementation");
        DelayedWETH weth = DelayedWETH(mustGetChainAddress("DelayedWETHProxy"));
        // use freshly deployed factory and anchorStateRegister
        DisputeGameFactory factory = DisputeGameFactory(mustGetAddress("DisputeGameFactoryProxy"));
        AnchorStateRegistry anchorStateRegistry = AnchorStateRegistry(mustGetAddress("AnchorStateRegistryProxy"));

        if (address(factory.gameImpls(GameTypes.ASTERISC)) != address(0) && !_allowUpgrade) {
            console.log(
                "[WARN] DisputeGameFactoryProxy: `FaultDisputeGame` implementation already set for game type: ASTERISC"
            );
            return;
        }

        FaultDisputeGame fdg = new FaultDisputeGame{ salt: _implSalt() }({
            _gameType: GameTypes.ASTERISC,
            _absolutePrestate: loadRiscvAbsolutePrestate(),
            _maxGameDepth: cfg.faultGameMaxDepth(),
            _splitDepth: cfg.faultGameSplitDepth(),
            _clockExtension: Duration.wrap(uint64(cfg.faultGameClockExtension())),
            _maxClockDuration: Duration.wrap(uint64(cfg.faultGameMaxClockDuration())),
            _vm: IBigStepper(mustGetAddress("RISCV")),
            _weth: weth,
            _anchorStateRegistry: anchorStateRegistry,
            _l2ChainId: cfg.l2ChainID()
        });

        factory.setImplementation(GameTypes.ASTERISC, fdg);

        console.log(
            "DisputeGameFactoryProxy: set `FaultDisputeGame` implementation (Backend: ASTERISC | GameType: %s)",
            vm.toString(GameType.unwrap(GameTypes.ASTERISC))
        );

        factory.setInitBond(GameTypes.ASTERISC, 0.08 ether);
    }

    /// @notice Checks that the deployed system is configured correctly.
    function postDeployAssertions() internal {
        // Ensure that `useFaultProofs` is set to `true`.
        assertTrue(cfg.useFaultProofs());

        ProxyAdmin proxyAdmin = ProxyAdmin(mustGetAddress("ProxyAdmin"));
        assertEq(address(proxyAdmin.owner()), msg.sender);

        // Ensure the contracts are owned by the correct entities.
        address dgfProxyAddr = mustGetAddress("DisputeGameFactoryProxy");
        DisputeGameFactory dgfProxy = DisputeGameFactory(dgfProxyAddr);
        assertEq(address(dgfProxy.owner()), msg.sender);

        PreimageOracle oracle = PreimageOracle(mustGetChainAddress("PreimageOracle"));
        assertEq(oracle.minProposalSize(), cfg.preimageOracleMinProposalSize());
        assertEq(oracle.challengePeriod(), cfg.preimageOracleChallengePeriod());

        RISCV riscv = RISCV(mustGetAddress("RISCV"));
        assertEq(address(riscv.oracle()), address(oracle));

        // Check the AnchorStateRegistry configuration.
        AnchorStateRegistry asr = AnchorStateRegistry(mustGetAddress("AnchorStateRegistryProxy"));
        (Hash root, uint256 l2BlockNumber) = asr.anchors(GameTypes.ASTERISC);
        assertEq(root.raw(), cfg.faultGameGenesisOutputRoot());
        assertEq(l2BlockNumber, cfg.faultGameGenesisBlock());

        // Check the FaultDisputeGame configuration.
        FaultDisputeGame gameImpl = FaultDisputeGame(payable(address(dgfProxy.gameImpls(GameTypes.ASTERISC))));
        assertEq(gameImpl.maxGameDepth(), cfg.faultGameMaxDepth());
        assertEq(gameImpl.splitDepth(), cfg.faultGameSplitDepth());
        assertEq(gameImpl.clockExtension().raw(), cfg.faultGameClockExtension());
        assertEq(gameImpl.maxClockDuration().raw(), cfg.faultGameMaxClockDuration());
        if (block.chainid == Chains.LocalDevnet || block.chainid == Chains.GethDevnet) {
            console.log("Cannot check absolute prestate because used locally generated prestate");
        } else {
            assertEq(gameImpl.absolutePrestate().raw(), bytes32(cfg.faultGameAbsolutePrestate()));
        }
        address wethProxyAddr = mustGetChainAddress("DelayedWETHProxy");
        assertEq(address(gameImpl.weth()), wethProxyAddr);
        assertEq(address(gameImpl.anchorStateRegistry()), address(asr));
        assertEq(address(gameImpl.vm()), address(riscv));
    }

    /// @notice Prints a review of the fault proof configuration section of the deploy config.
    function printConfigReview() internal view {
        console.log(unicode"📖 FaultDisputeGame Config Overview (chainid: %d)", block.chainid);
        console.log("    0. Use Fault Proofs: %s", cfg.useFaultProofs() ? "true" : "false");
        console.log("    1. Absolute Prestate: %x", cfg.faultGameAbsolutePrestate());
        if (block.chainid == Chains.LocalDevnet || block.chainid == Chains.GethDevnet) {
            console.log("    - Deployment did not use prestate provided by config");
        }
        console.log("    2. Max Depth: %d", cfg.faultGameMaxDepth());
        console.log("    3. Output / Execution split Depth: %d", cfg.faultGameSplitDepth());
        console.log("    4. Clock Extension (seconds): %d", cfg.faultGameClockExtension());
        console.log("    5. Max Clock Duration (seconds): %d", cfg.faultGameMaxClockDuration());
        console.log("    6. L2 Genesis block number: %d", cfg.faultGameGenesisBlock());
        console.log("    7. L2 Genesis output root: %x", uint256(cfg.faultGameGenesisOutputRoot()));
        console.log("    8. Proof Maturity Delay (seconds): ", cfg.proofMaturityDelaySeconds());
        console.log("    9. Dispute Game Finality Delay (seconds): ", cfg.disputeGameFinalityDelaySeconds());
        console.log("   10. Respected Game Type: ", cfg.respectedGameType());
        console.log("   11. Preimage Oracle Min Proposal Size (bytes): ", cfg.preimageOracleMinProposalSize());
        console.log("   12. Preimage Oracle Challenge Period (seconds): ", cfg.preimageOracleChallengePeriod());
    }
}
