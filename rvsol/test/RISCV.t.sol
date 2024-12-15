// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Test } from "@forge-std/Test.sol";
import { IPreimageOracle } from "@optimism/src/cannon/interfaces/IPreimageOracle.sol";
import { IBigStepper } from "@optimism/src/dispute/interfaces/IBigStepper.sol";
import { PreimageOracle } from "@optimism/src/cannon/PreimageOracle.sol";
import { DeployUtils } from "@optimism/scripts/libraries/DeployUtils.sol";
import { CommonTest } from "./CommonTest.sol";
import "@optimism/src/dispute/lib/Types.sol";

contract RISCV_Test is CommonTest {
    /// @notice Stores the VM state.
    ///         Total state size: 32 + 32 + 8 * 2 + 1 * 2 + 8 * 3 + 32 * 8 = 362 bytes
    ///         Note that struct is not used for step execution and used only for testing
    //          Struct size may be larger than total state size due to memory layouts
    struct State {
        bytes32 memRoot;
        bytes32 preimageKey;
        uint64 preimageOffset;
        uint64 pc;
        uint8 exitCode;
        bool exited;
        uint64 step;
        uint64 heap;
        uint64 loadReservation;
        uint64[32] registers;
    }

    IBigStepper internal riscv;
    PreimageOracle internal oracle;

    function setUp() public virtual override {
        super.setUp();
        oracle = new PreimageOracle(0, 0);
        riscv =
            IBigStepper(DeployUtils.create1({ _name: "RISCV", _args: abi.encode(IPreimageOracle(address(oracle))) }));
        vm.store(address(riscv), 0x0, bytes32(abi.encode(address(oracle))));
        vm.label(address(oracle), "PreimageOracle");
        vm.label(address(riscv), "RISCV");
    }

    function test_step_abi_succeeds() public {
        // state and proof from first step of `simple` binary
        uint64[32] memory registers;
        registers[2] = 0x1000000000000000;
        State memory state = State({
            memRoot: hex"f0df7f266aed88bde90ed121f0de6865f3fa88bf67d3a4657dad876038393b2c",
            preimageKey: bytes32(0),
            preimageOffset: 0,
            pc: 471932,
            exitCode: 0,
            exited: false,
            step: 1,
            heap: 0x7f0000000000,
            loadReservation: 0,
            registers: registers
        });
        bytes memory proof =
            hex"67800f0000000000971f000067800fb40000000000000000033501009305810083348102033401028333810103330101833281008330011d833f01001301811d3c68dba488488bae6478015e476f03a8d0b8f27f087b388bc41ed6c40c492b8ddb41e1d33c6d417324675080ecc5eea5b78f9f539896eb892480de2d33425b20420848eec624fdddc1dac146378ea52a5f03ebb2406d89e01d6304eea742033b42251ce9146b8e43af396434ba823722b4b9977c7062ef2322e5aeb382aefed453b602acc24b2b7d34a8ff2517b7499c9b20510277c2ae05f9cb5fd208ae88a62487d85a07577b9b2c16090488dcfc1fd6ade786ce75056d078abb377db79b211ed2e42c800d3dbb0340afd72bbf760305c444b999a6c6c6d32ee6e9673249d1730c967c62d92e2699234529fa4b749784620a21a0c1a4b2ad81da6507e4fb66fca30cbd5a4da0f9cd5636ab0fc223d399af831578c83d4c10c38972964ba0d670bed1afb5ffc60a2d4dde7e36f5a498f0671d880973cabeeca428a627c5a04b16268248aef083470b7c9e91aeeb49da103cd6519718cca728fda79218038f29e70762ff98d65de0e69f568fa353d115bbf9b5b42dc397706afdcf6d2ff2a68153e7f911d48d5c6292883912b3ee8852e64b8229080b8888b1e9f61524aee439bcdbaf59170f519ccef13111146b601aeba12c990e5f484ea70a617f5ea2f38c538635459bf00023877e777e6c3041df40cfb93eb8637d06ea44eb1f88a91e0adf644bb7710c751982cbbb32a4003bc655cc26cbea017bdd9dcd192c860eff71e1d3b5c807b281e4683cc6d6315cf95b9ade8641defcb32372f1c126e398ef7a5a2dce0a8a7f68bb74560f8f71837c2c2ebbcbf7fffb42ae1896f13f7c7479a0b46a28b6f55540f89444f63de0378e3d121be09e06cc9ded1c20e65876d36aa0c65e9645644786b620e2dd2ad648ddfcbf4a7e5b1a3a4ecfe7f64667a3f0b7e2f4418588ed35a2458cffeb39b93d26f18d2ab13bdce6aee58e7b99359ec2dfd95a9c16dc00d6ef18b7933a6f8dc65ccb55667138776f7dea101070dc8796e3774df84f40ae0c8229d0d6069e5c8f39a7c299677a09d367fc7b05e3bc380ee652cdc72595f74c7b1043d0e1ffbab734648c838dfb0527d971b602bc216c9619ef0abf5ac974a1ed57f4050aa510dd9c74f508277b39d7973bb2dfccc5eeb0618db8cd74046ff337f0a7bf2c8e03e10f642c1886798d71806ab1e888d9e5ee87d0838c5655cb21c6cb83313b5a631175dff4963772cce9108188b34ac87c81c41e662ee4dd2dd7b2bc707961b1e646c4047669dcb6584f0d8d770daf5d7e7deb2e388ab20e2573d171a88108e79d820e98f26c0b84aa8b2f4aa4968dbb818ea32293237c50ba75ee485f4c22adf2f741400bdf8d6a9cc7df7ecae576221665d7358448818bb4ae4562849e949e17ac16e0be16688e156b5cf15e098c627c0056a927ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757bf558bebd2ceec7f3c5dce04a4782f88c2c6036ae78ee206d0bc5289d20461a2e21908c2968c0699040a6fd866a577a99a9d2ec88745c815fd4a472c789244daae824d72ddc272aab68a8c3022e36f10454437c1886f3ff9927b64f232df414f27e429a4bef3083bc31a671d046ea5c1f5b8c3094d72868d9dfdc12c7334ac5f743cc5c365a9a6a15c1f240ac25880c7a9d1de290696cb766074a1d83d9278164adcf616c3bfabf63999a01966c998b7bb572774035a63ead49da73b5987f34775786645d0c5dd7c04a2f8a75dcae085213652f5bce3ea8b9b9bedd1cab3c5e9b88b152c9b8a7b79637d35911848b0c41e7cc7cca2ab4fe9a15f9c38bb4bb9390c4e2d8ce834ffd7a6cd85d7113d4521abb857774845c4291e6f6d010d97e3185bc799d83e3bb31501b3da786680df30fbc18eb41cbce611e8c0e9c72f69571ca10d3ef857d04d9c03ead7c6317d797a090fa1271ad9c7addfbcb412e9643d4fb33b1809c42623f474055fa9400a2027a7a885c8dfa4efe20666b4ee27d7529c134d7f28d53f175f6bf4b62faa2110d5b76f0f770c15e628181c1fcc18f970a9c34d24b2fc8c50ca9c07a7156ef4e5ff4bdf002eda0b11c1d359d0b59a54680704dbb9db631457879b27e0dfdbe50158fd9cf9b4cf77605c4ac4c95bd65fc9f6f9295a686647cb999090819cda700820c282c613cedcd218540bbc6f37b01c6567c4a1ea624f092a3a5cca2d6f0f0db231972fce627f0ecca0dee60f17551c5f8fdaeb5ab560b2ceb781cdb339361a0fbee1b9dffad59115138c8d6a70dda9ccc1bf0bbdd7fee15764845db875f6432559ff8dbc9055324431bc34e5b93d15da307317849eccd90c0c7b98870b9317c15a5959dcfb84c76dcc908c4fe6ba92126339bf06e458f6646df5e83ba7c3d35bc263b3222c8e9040068847749ca8e8f95045e4342aeb521eb3a5587ec268ed3aa6faf32b62b0bc41a9d549521f406fc3ec7d4dabb75e0d3e144d7cc882372d13746b6dcd481b1b229bcaec9f7422cdfb84e35c5d92171376cae5c86300822d729cd3a8479583bef09527027dba5f11263c5cbbeb3834b7a5c1cba9aa5fee0c95ec3f17a33ec3d8047fff799187f5ae2040bbe913c226c34c9fbe4389dd728984257a816892b3cae3e43191dd291f0eb50000000000000000420000000000000035000000000000000000000000000000060000000000000000100000000000001900000000000000480000000000001050edbc06b4bfc3ee108b66f7a8f772ca4d90e1a085f4a8398505920f7465bb44b4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d3021ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85e58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a193440eb01ebfc9ed27500cd4dfc979272d1f0913cc9f66540d7e8005811109e1cf2d887c22bd8750d34016ac3c66b5ff102dacdd73f6b014e710b51e8022af9a1968ffd70157e48063fc33c97a050f7f640233bf646cc98d9524c6b92bcf3ab56f839867cc5f7f196b93bae1e27e6320742445d290f2263827498b54fec539f756afcefad4e508c098b9a7e1d8feb19955fb02ba9675585078710969d3440f5054e0f9dc3e7fe016e050eff260334f18a5d4fe391d82092319f5964f2e2eb7c1c3a5f8b13a49e282f609c317a833fb8d976d11517c571d1221a265d25af778ecf8923490c6ceeb450aecdc82e28293031d10c7d73bf85e57bf041a97360aa2c5d99cc1df82d9c4b87413eae2ef048f94b4d3554cea73d92b0f7af96e0271c691e2bb5c67add7c6caf302256adedf7ab114da0acfe870d449a3a489f781d659e8beccda7bce9f4e8618b6bd2f4132ce798cdc7a60e7e1460a7299e3c6342a579626d22733e50f526ec2fa19a22b31e8ed50f23cd1fdf94c9154ed3a7609a2f1ff981fe1d3b5c807b281e4683cc6d6315cf95b9ade8641defcb32372f1c126e398ef7a5a2dce0a8a7f68bb74560f8f71837c2c2ebbcbf7fffb42ae1896f13f7c7479a0b46a28b6f55540f89444f63de0378e3d121be09e06cc9ded1c20e65876d36aa0c65e9645644786b620e2dd2ad648ddfcbf4a7e5b1a3a4ecfe7f64667a3f0b7e2f4418588ed35a2458cffeb39b93d26f18d2ab13bdce6aee58e7b99359ec2dfd95a9c16dc00d6ef18b7933a6f8dc65ccb55667138776f7dea101070dc8796e3774df84f40ae0c8229d0d6069e5c8f39a7c299677a09d367fc7b05e3bc380ee652cdc72595f74c7b1043d0e1ffbab734648c838dfb0527d971b602bc216c9619ef0abf5ac974a1ed57f4050aa510dd9c74f508277b39d7973bb2dfccc5eeb0618db8cd74046ff337f0a7bf2c8e03e10f642c1886798d71806ab1e888d9e5ee87d0838c5655cb21c6cb83313b5a631175dff4963772cce9108188b34ac87c81c41e662ee4dd2dd7b2bc707961b1e646c4047669dcb6584f0d8d770daf5d7e7deb2e388ab20e2573d171a88108e79d820e98f26c0b84aa8b2f4aa4968dbb818ea32293237c50ba75ee485f4c22adf2f741400bdf8d6a9cc7df7ecae576221665d7358448818bb4ae4562849e949e17ac16e0be16688e156b5cf15e098c627c0056a927ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757bf558bebd2ceec7f3c5dce04a4782f88c2c6036ae78ee206d0bc5289d20461a2e21908c2968c0699040a6fd866a577a99a9d2ec88745c815fd4a472c789244daae824d72ddc272aab68a8c3022e36f10454437c1886f3ff9927b64f232df414f27e429a4bef3083bc31a671d046ea5c1f5b8c3094d72868d9dfdc12c7334ac5f743cc5c365a9a6a15c1f240ac25880c7a9d1de290696cb766074a1d83d9278164adcf616c3bfabf63999a01966c998b7bb572774035a63ead49da73b5987f34775786645d0c5dd7c04a2f8a75dcae085213652f5bce3ea8b9b9bedd1cab3c5e9b88b152c9b8a7b79637d35911848b0c41e7cc7cca2ab4fe9a15f9c38bb4bb9390c4e2d8ce834ffd7a6cd85d7113d4521abb857774845c4291e6f6d010d97e3185bc799d83e3bb31501b3da786680df30fbc18eb41cbce611e8c0e9c72f69571ca10d3ef857d04d9c03ead7c6317d797a090fa1271ad9c7addfbcb412e9643d4fb33b1809c42623f474055fa9400a2027a7a885c8dfa4efe20666b4ee27d7529c134d7f28d53f175f6bf4b62faa2110d5b76f0f770c15e628181c1fcc18f970a9c34d24b2fc8c50ca9c07a7156ef4e5ff4bdf002eda0b11c1d359d0b59a54680704dbb9db631457879b27e0dfdbe50158fd9cf9b4cf77605c4ac4c95bd65fc9f6f9295a686647cb999090819cda700820c282c613cedcd218540bbc6f37b01c6567c4a1ea624f092a3a5cca2d6f0f0db231972fce627f0ecca0dee60f17551c5f8fdaeb5ab560b2ceb781cdb339361a0fbee1b9dffad59115138c8d6a70dda9ccc1bf0bbdd7fee15764845db875f6432559ff8dbc9055324431bc34e5b93d15da307317849eccd90c0c7b98870b9317c15a5959dcfb84c76dcc908c4fe6ba92126339bf06e458f6646df5e83ba7c3d35bc263b3222c8e9040068847749ca8e8f95045e4342aeb521eb3a5587ec268ed3aa6faf32b62b0bc41a9d549521f406fc3bbdff18e513dcd75f7e478e4acb5c91463476a9d83b6b77b4c56ecfe549280ab84e35c5d92171376cae5c86300822d729cd3a8479583bef09527027dba5f11263c5cbbeb3834b7a5c1cba9aa5fee0c95ec3f17a33ec3d8047fff799187f5ae2040bbe913c226c34c9fbe4389dd728984257a816892b3cae3e43191dd291f0eb5";
        bytes32 postState = riscv.step(encodeState(state), proof, 0);
        assertTrue(postState != bytes32(0));
    }

    /* R Type instructions */

    function test_add_succeeds() public {
        uint32 insn = encodeRType(0x33, 18, 0, 5, 14, 0); // add x18, x5, x14
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[5] = 0xe2f9;
        state.registers[14] = 0x6c13;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[18] = state.registers[5] + state.registers[14];
        expect.registers[5] = state.registers[5];
        expect.registers[14] = state.registers[14];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sub_succeeds() public {
        uint32 insn = encodeRType(0x33, 8, 0, 16, 18, 32); // sub x8, x16, x18
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[16] = 0xe0ba;
        state.registers[18] = 0xda71;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[8] = state.registers[16] - state.registers[18];
        expect.registers[16] = state.registers[16];
        expect.registers[18] = state.registers[18];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sll_succeeds() public {
        uint32 insn = encodeRType(0x33, 12, 1, 22, 26, 0); // sll x12, x22, x26
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[22] = 0x576a;
        state.registers[26] = 0x3;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[12] = state.registers[22] << state.registers[26];
        expect.registers[22] = state.registers[22];
        expect.registers[26] = state.registers[26];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_slt_succeeds() public {
        uint32 insn = encodeRType(0x33, 13, 2, 1, 23, 0); // slt x13, x1, x23
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[1] = 0xf17a;
        state.registers[23] = 0x2a22;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[13] = state.registers[1] < state.registers[23] ? 1 : 0;
        expect.registers[1] = state.registers[1];
        expect.registers[23] = state.registers[23];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sltu_succeeds() public {
        uint32 insn = encodeRType(0x33, 2, 3, 14, 16, 0); // sltu x2, x14, x16
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[14] = 0x9cb6;
        state.registers[16] = 0x79e2;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[2] = state.registers[14] < state.registers[16] ? 1 : 0;
        expect.registers[14] = state.registers[14];
        expect.registers[16] = state.registers[16];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_xor_succeeds() public {
        uint32 insn = encodeRType(0x33, 29, 4, 17, 16, 0); // xor x29, x17, x16
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[17] = 0xb2f2;
        state.registers[16] = 0xb5b7;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[29] = state.registers[17] ^ state.registers[16];
        expect.registers[17] = state.registers[17];
        expect.registers[16] = state.registers[16];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_srl_succeeds() public {
        uint32 insn = encodeRType(0x33, 30, 5, 15, 26, 0); // srl x30, x15, x26
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[15] = 0x813d;
        state.registers[26] = 0x7;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[30] = state.registers[15] >> state.registers[26];
        expect.registers[15] = state.registers[15];
        expect.registers[26] = state.registers[26];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sra_succeeds() public {
        uint32 insn = encodeRType(0x33, 14, 5, 23, 2, 32); // sra x14, x23, x2
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        // intentionally set MSB to 1 to check sign preservation
        state.registers[23] = 0xFF12_3412_3412_FFFF;
        state.registers[2] = 0x8;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[14] = state.registers[23] >> state.registers[2];
        bool signBit = (1 << 63) & state.registers[23] > 0;
        if (signBit) {
            uint64 signExtension = uint64(((1 << state.registers[2]) - 1) << (64 - state.registers[2]));
            expect.registers[14] |= signExtension;
        }
        expect.registers[23] = state.registers[23];
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_or_succeeds() public {
        uint32 insn = encodeRType(0x33, 26, 6, 30, 18, 0); // or x26, x30, x18
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[30] = 0x3a8b;
        state.registers[18] = 0xdcff;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[26] = state.registers[30] | state.registers[18];
        expect.registers[30] = state.registers[30];
        expect.registers[18] = state.registers[18];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_and_succeeds() public {
        uint32 insn = encodeRType(0x33, 23, 7, 24, 2, 0); // and x23, x24, x2
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[24] = 0x6d53;
        state.registers[2] = 0xe105;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[23] = state.registers[24] & state.registers[2];
        expect.registers[24] = state.registers[24];
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_addw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 14, 0, 17, 19, 0); // addw x14, x17, x19
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[17] = 0x5742;
        state.registers[19] = 0xfee0;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[14] = state.registers[17] + state.registers[19];
        expect.registers[17] = state.registers[17];
        expect.registers[19] = state.registers[19];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_subw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 28, 0, 13, 14, 32); // subw x28, x13, x14
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[13] = 0x3ea9;
        state.registers[14] = 0x1d1f;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[28] = state.registers[13] - state.registers[14];
        expect.registers[13] = state.registers[13];
        expect.registers[14] = state.registers[14];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sllw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 12, 1, 21, 16, 0); // sllw x12, x21, x16
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[21] = 0xe77a;
        state.registers[16] = 0xc;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[12] = state.registers[21] << state.registers[16];
        expect.registers[21] = state.registers[21];
        expect.registers[16] = state.registers[16];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_srlw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 10, 5, 22, 3, 0); // srlw x10, x22, x3
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[22] = 0xf3be;
        state.registers[3] = 0x5;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[10] = state.registers[22] >> state.registers[3];
        expect.registers[22] = state.registers[22];
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sraw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 17, 5, 11, 4, 32); // sraw x17, x11, x4
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        // intentionally set MSB to 1 to check sign preservation
        state.registers[11] = 0xFF12_3412;
        state.registers[4] = 0x7;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[17] = state.registers[11] >> state.registers[4];
        bool signBit = (1 << 31) & state.registers[11] > 0;
        if (signBit) {
            uint64 signExtension = uint64(((1 << (32 + state.registers[4])) - 1) << (32 - state.registers[4]));
            expect.registers[17] |= signExtension;
        }

        expect.registers[11] = state.registers[11];
        expect.registers[4] = state.registers[4];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_mul_succeeds() public {
        uint32 insn = encodeRType(0x33, 24, 0, 26, 22, 1); // mul x24, x26, x22
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[26] = 0x2612a5fed675423;
        state.registers[22] = 0x3441e4c58579b6b8;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        int256 temp = int256(int64(state.registers[26])) * int256(int64(state.registers[22]));
        expect.registers[24] = uint64(uint256(temp & ((1 << 64) - 1)));
        expect.registers[26] = state.registers[26];
        expect.registers[22] = state.registers[22];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_mulh_succeeds() public {
        uint32 insn = encodeRType(0x33, 20, 1, 31, 2, 1); // mulh x20, x31, x2
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[31] = 0x665af09d9da34d2f;
        state.registers[2] = 0x25ab6e605bdd3e31;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        int256 temp = int256(int64(state.registers[31])) * int256(int64(state.registers[2]));
        expect.registers[20] = uint64(uint256((temp >> 64) & ((1 << 64) - 1)));
        expect.registers[31] = state.registers[31];
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_mulhsu_succeeds() public {
        uint32 insn = encodeRType(0x33, 7, 2, 5, 27, 1); // mulhsu x7, x5, x27
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[5] = 0x6f050b64e4f37291;
        state.registers[27] = 0x2b29ce113892ba69;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        int256 temp = int256(int64(state.registers[5])) * int256(int128(uint128(state.registers[27])));
        expect.registers[7] = uint64(uint256((temp >> 64) & ((1 << 64) - 1)));
        expect.registers[5] = state.registers[5];
        expect.registers[27] = state.registers[27];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_mulhu_succeeds() public {
        uint32 insn = encodeRType(0x33, 27, 3, 6, 18, 1); // mulhu x27, x6, x18
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[6] = 0x4544e15a2fb9c134;
        state.registers[18] = 0xa11b583879ae1211;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint256 temp = uint256(state.registers[6]) * uint256(state.registers[18]);
        expect.registers[27] = uint64(uint256((temp >> 64) & ((1 << 64) - 1)));
        expect.registers[6] = state.registers[6];
        expect.registers[18] = state.registers[18];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_div_succeeds() public {
        uint32 insn = encodeRType(0x33, 18, 4, 4, 29, 1); // div x18, x4, x29
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[4] = 0xf0e9720e82469b35;
        state.registers[29] = 0xf391cc717328ac0;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[18] = uint64(int64(state.registers[4]) / int64(state.registers[29]));
        expect.registers[4] = state.registers[4];
        expect.registers[29] = state.registers[29];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_divu_succeeds() public {
        uint32 insn = encodeRType(0x33, 28, 5, 4, 11, 1); // divu x28, x4, x11
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[4] = 0xf555d9a795ed923c;
        state.registers[11] = 0xf9d68900a39ad4ec;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[28] = state.registers[4] / state.registers[11];
        expect.registers[4] = state.registers[4];
        expect.registers[11] = state.registers[11];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_rem_succeeds() public {
        uint32 insn = encodeRType(0x33, 5, 6, 16, 27, 1); // rem x5, x16, x27
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[16] = 0x4fbb7d7830691641;
        state.registers[27] = 0x18f12d87ce1a5546;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[5] = uint64(int64(state.registers[16]) % int64(state.registers[27]));
        expect.registers[16] = state.registers[16];
        expect.registers[27] = state.registers[27];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_remu_succeeds() public {
        uint32 insn = encodeRType(0x33, 3, 7, 14, 1, 1); // remu x3, x14, x1
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[14] = 0x770cd5ca54465cd1;
        state.registers[1] = 0x691c4f46194b3fa4;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[3] = state.registers[14] % state.registers[1];
        expect.registers[14] = state.registers[14];
        expect.registers[1] = state.registers[1];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_mulw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 18, 0, 21, 7, 1); // mulw x18, x21, x7
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[21] = 0x485b637df7d02127;
        state.registers[7] = 0xc2a29e37cd8ffdae;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        int256 temp = int256(int64(state.registers[21])) * int256(int64(state.registers[7]));
        expect.registers[18] = uint64(uint256(temp & ((1 << 32) - 1)));
        bool signBit = (1 << 31) & expect.registers[18] > 0;
        if (signBit) {
            expect.registers[18] |= ((1 << 32) - 1) << 32;
        }
        expect.registers[21] = state.registers[21];
        expect.registers[7] = state.registers[7];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_divw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 2, 4, 30, 20, 1); // divw x2, x30, x20
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[30] = 0x265b398efecfbcb0;
        state.registers[20] = 0x43175ecbdf9bbd84;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 temp1 = mask32Signed64(state.registers[30]);
        uint64 temp2 = mask32Signed64(state.registers[20]);
        uint64 temp = uint64(int64(temp1) / int64(temp2));
        expect.registers[2] = mask32Signed64(temp);
        expect.registers[30] = state.registers[30];
        expect.registers[20] = state.registers[20];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_divuw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 3, 5, 21, 7, 1); // divuw x3, x21, x7
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[21] = 0x6f2caeeb7e4e97b3;
        state.registers[7] = 0x51cf2e551f6a5e0;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 temp1 = mask32Unsigned64(state.registers[21]);
        uint64 temp2 = mask32Unsigned64(state.registers[7]);
        expect.registers[3] = mask32Unsigned64(temp1 / temp2);
        expect.registers[21] = state.registers[21];
        expect.registers[7] = state.registers[7];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_remw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 27, 6, 22, 21, 1); // remw x27, x22, x21
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[22] = 0x9f0ebf8dfc2febe0;
        state.registers[21] = 0xb704babb86c919bf;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 temp1 = mask32Signed64(state.registers[22]);
        uint64 temp2 = mask32Signed64(state.registers[21]);
        uint64 temp = uint64(int64(temp1) % int64(temp2));
        expect.registers[27] = mask32Signed64(temp);
        expect.registers[22] = state.registers[22];
        expect.registers[21] = state.registers[21];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_remuw_succeeds() public {
        uint32 insn = encodeRType(0x3b, 30, 7, 27, 9, 1); // remuw x30, x27, x9
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[27] = 0x1ccfe2acc3d2fa50;
        state.registers[9] = 0xeb03331a300718a5;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 temp1 = mask32Unsigned64(state.registers[27]);
        uint64 temp2 = mask32Unsigned64(state.registers[9]);
        expect.registers[30] = mask32Unsigned64(temp1 % temp2);
        expect.registers[27] = state.registers[27];
        expect.registers[9] = state.registers[9];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lrw_succeeds() public {
        bytes32 value = hex"1e0acbdd44d41d85";
        uint64 addr = 0x233f3d38d3ce6668;
        uint8 funct3 = 0x2;
        uint8 funct7 = encodeFunct7(0x2, 0x0, 0x0);
        uint8 size = uint8(1 << (funct3 & 0x3));
        uint32 insn = encodeRType(0x2f, 24, funct3, 28, 0, funct7); // lrw x24, x15, (x28)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[28] = addr;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.loadReservation = addr;
        expect.registers[24] = bytes32ToUint64(value, size);
        expect.registers[24] = mask32Signed64(expect.registers[24]);
        expect.registers[28] = state.registers[28];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_scw_succeeds() public {
        uint64 addr = 0x39c314f9013a2b30;
        uint8 funct3 = 0x2;
        uint8 funct7 = encodeFunct7(0x3, 0x0, 0x0);
        uint8 size = uint8(1 << (funct3 & 0x3));
        uint32 insn = encodeRType(0x2f, 23, funct3, 27, 30, funct7); // scw x23, x30, (x27)
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"3ee07aaba5c04760", size);
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.loadReservation = addr;
        state.registers[27] = addr;
        state.registers[30] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, rs2ValueBytes32);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.loadReservation = 0;
        expect.registers[23] = 0; // sc succeeded
        expect.registers[27] = state.registers[27];
        expect.registers[30] = state.registers[30];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoswapw_succeeds() public {
        uint64 addr = 0x44c23256360226b0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x1, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 22, funct3, 4, 14, funct7); // amoswapw x22, x14, (x4)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"e4a97cf4a798bf55", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"23dcb1b1b1ab1969", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[4] = addr;
        state.registers[14] = mask32Signed64(rs2ValueU64);
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of rs2
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, rs2ValueBytes32);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[22] = mask32Signed64(memValueU64);
        expect.registers[4] = state.registers[4];
        expect.registers[14] = state.registers[14];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoaddw_succeeds() public {
        uint64 addr = 0xbf1cd3785c3b5e0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x0, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 23, funct3, 17, 3, funct7); // amoaddw x23, x3, (x17)
        }
        (, uint64 rs2ValueU64) = truncate(hex"37f64a206d30a374", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"99675cd137120f0e", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[17] = addr;
        state.registers[3] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] + x[rs2]
        bytes32 result = uint256ToBytes32(
            uint256(
                mask32Signed64(uint64(int64(int32(int64(rs2ValueU64))) + int64(int32(int64(memValueU64)))))
                    & ((1 << 32) - 1)
            )
        );
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[23] = mask32Signed64(memValueU64);
        expect.registers[17] = state.registers[17];
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoxorw_succeeds() public {
        uint64 addr = 0xd9a8dd911b0547cc;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x4, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 10, funct3, 11, 15, funct7); // amoxorw x10, x15, (x11)
        }
        (, uint64 rs2ValueU64) = truncate(hex"57163d5d64e31c6c", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"1f6d7f7941fde4e5", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[11] = addr;
        state.registers[15] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] ^ x[rs2]
        bytes32 result = uint256ToBytes32(
            uint256(
                mask32Signed64(uint64(int64(int32(int64(rs2ValueU64))) ^ int64(int32(int64(memValueU64)))))
                    & ((1 << 32) - 1)
            )
        );
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[10] = mask32Signed64(memValueU64);
        expect.registers[11] = state.registers[11];
        expect.registers[15] = state.registers[15];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoandw_succeeds() public {
        uint64 addr = 0x5519c1cd82d36828;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0xc, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 22, funct3, 25, 20, funct7); // amoandw x22, x20, (x25)
        }
        (, uint64 rs2ValueU64) = truncate(hex"f52f78fff989efe3", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"315275be66ef0e76", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[25] = addr;
        state.registers[20] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] & x[rs2]
        bytes32 result = uint256ToBytes32(
            uint256(
                mask32Signed64(uint64(int64(int32(int64(rs2ValueU64))) & int64(int32(int64(memValueU64)))))
                    & ((1 << 32) - 1)
            )
        );
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[22] = mask32Signed64(memValueU64);
        expect.registers[25] = state.registers[25];
        expect.registers[20] = state.registers[20];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoorw_succeeds() public {
        uint64 addr = 0x2dbd6638ebe8a250;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x8, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 27, funct3, 16, 1, funct7); // amoorw x27, x1, (x16)
        }
        (, uint64 rs2ValueU64) = truncate(hex"0d204e771480f255", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"8daa13a8b68b622c", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[16] = addr;
        state.registers[1] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] | x[rs2]
        bytes32 result = uint256ToBytes32(
            uint256(
                mask32Signed64(uint64(int64(int32(int64(rs2ValueU64))) | int64(int32(int64(memValueU64)))))
                    & ((1 << 32) - 1)
            )
        );
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[27] = mask32Signed64(memValueU64);
        expect.registers[16] = state.registers[16];
        expect.registers[1] = state.registers[1];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amominw_succeeds() public {
        uint64 addr = 0xbb0517653427ed98;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x10, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 15, funct3, 24, 13, funct7); // amominw x15, x13, (x24)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"37f64a206d30a374", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"f4844f357c630c38", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[24] = addr;
        state.registers[13] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of min(M[x[rs1]], x[rs2])
        bytes32 result = int32(int64(rs2ValueU64)) < int32(int64(memValueU64)) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[15] = mask32Signed64(memValueU64);
        expect.registers[24] = state.registers[24];
        expect.registers[13] = state.registers[13];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amomaxw_succeeds() public {
        uint64 addr = 0xb320adad61ff64b8;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x14, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 4, funct3, 8, 2, funct7); // amomaxw x4, x2, (x8)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"d574e48626033174", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"a525950d1aa4973a", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[8] = addr;
        state.registers[2] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of max(M[x[rs1]], x[rs2])
        bytes32 result = int32(int64(rs2ValueU64)) > int32(int64(memValueU64)) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[4] = mask32Signed64(memValueU64);
        expect.registers[8] = state.registers[8];
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amominuw_succeeds() public {
        uint64 addr = 0xc00b31ae34210ac8;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x18, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 17, funct3, 24, 18, funct7); // amominuw x17, x18, (x24)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"cdeab94408c734f5", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"478fbd468e60ac23", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[24] = addr;
        state.registers[18] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of min(unsigned M[x[rs1]], unsigned x[rs2])
        bytes32 result =
            uint32(int32(int64(rs2ValueU64))) < uint32(int32(int64(memValueU64))) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[17] = mask32Signed64(memValueU64);
        expect.registers[24] = state.registers[24];
        expect.registers[18] = state.registers[18];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amomaxuw_succeeds() public {
        uint64 addr = 0xca0b8f3993fbb894;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x2;
            uint8 funct7 = encodeFunct7(0x1c, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 20, funct3, 14, 23, funct7); // amomaxuw x20, x23, (x14)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"d9341fdf49efa3f6", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"134105b97e200641", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[14] = addr;
        state.registers[23] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of max(unsigned M[x[rs1]], unsigned x[rs2])
        bytes32 result =
            uint32(int32(int64(rs2ValueU64))) > uint32(int32(int64(memValueU64))) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[20] = mask32Signed64(memValueU64);
        expect.registers[14] = state.registers[14];
        expect.registers[23] = state.registers[23];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lrd_succeeds() public {
        bytes32 value = hex"a0b1df92a49eec39";
        uint64 addr = 0xb86a394544c084e0;
        uint8 funct3 = 0x3;
        uint8 funct7 = encodeFunct7(0x2, 0x0, 0x0);
        uint8 size = uint8(1 << (funct3 & 0x3));
        uint32 insn = encodeRType(0x2f, 14, funct3, 7, 13, funct7); // lrd x14, x13, (x7)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[7] = addr;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.loadReservation = addr;
        expect.registers[14] = bytes32ToUint64(value, size);
        expect.registers[14] = expect.registers[14];
        expect.registers[7] = state.registers[7];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_scd_succeeds() public {
        uint64 addr = 0x7d118f395f2decd0;
        uint8 funct3 = 0x3;
        uint8 funct7 = encodeFunct7(0x3, 0x0, 0x0);
        uint8 size = uint8(1 << (funct3 & 0x3));
        uint32 insn = encodeRType(0x2f, 4, funct3, 13, 24, funct7); // scd x4, x24, (x13)
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"3186582d2a2adf7d", size);
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.loadReservation = addr;
        state.registers[13] = addr;
        state.registers[24] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, rs2ValueBytes32);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[4] = 0; // sc succeeded
        expect.registers[13] = state.registers[13];
        expect.registers[24] = state.registers[24];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoswapd_succeeds() public {
        uint64 addr = 0x15f4716cd3aa7308;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x1, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 23, funct3, 30, 3, funct7); // amoswapd x23, x3, (x30)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"c30495901566e553", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"ee2a2e31e99971ad", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[30] = addr;
        state.registers[3] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of rs2
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, rs2ValueBytes32);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[23] = memValueU64;
        expect.registers[30] = state.registers[30];
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoaddd_succeeds() public {
        uint64 addr = 0xeae426a36ff2bb60;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x0, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 14, funct3, 8, 28, funct7); // amoaddd x14, x28, (x8)
        }
        (, uint64 rs2ValueU64) = truncate(hex"a0821b98f6c0d237", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"f47daefa285404dc", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[8] = addr;
        state.registers[28] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] + x[rs2]
        bytes32 result = uint256ToBytes32(uint256(uint128(int128(int64(rs2ValueU64)) + int128(int64(memValueU64)))));
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[14] = memValueU64;
        expect.registers[8] = state.registers[8];
        expect.registers[28] = state.registers[28];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoxord_succeeds() public {
        uint64 addr = 0x2d5ba68f57f1c560;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x4, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 9, funct3, 16, 19, funct7); // amoxord x9, x19, (x16)
        }
        (, uint64 rs2ValueU64) = truncate(hex"cee6d3e92e42e68d", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"a95b29ec1d9bc7d6", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[16] = addr;
        state.registers[19] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] ^ x[rs2]
        bytes32 result = uint256ToBytes32(uint256(rs2ValueU64 ^ memValueU64));
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[9] = memValueU64;
        expect.registers[16] = state.registers[16];
        expect.registers[19] = state.registers[19];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoandd_succeeds() public {
        uint64 addr = 0xd273284a99c8070;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0xc, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 9, funct3, 17, 13, funct7); // amoandd x9, x13, (x17)
        }
        (, uint64 rs2ValueU64) = truncate(hex"ad5ec3eef5264cb6", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"50bd66fb27a4ec4c", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[17] = addr;
        state.registers[13] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] & x[rs2]
        bytes32 result = uint256ToBytes32(uint256(rs2ValueU64 & memValueU64));
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[9] = memValueU64;
        expect.registers[17] = state.registers[17];
        expect.registers[13] = state.registers[13];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amoord_succeeds() public {
        uint64 addr = 0xa0d7a5ea65b35660;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x8, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 24, funct3, 5, 3, funct7); // amoord x24, x3, (x5)
        }
        (, uint64 rs2ValueU64) = truncate(hex"7acf784b9e7764d3", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"bcb6e898d4635f81", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[5] = addr;
        state.registers[3] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] | x[rs2]
        bytes32 result = uint256ToBytes32(uint256(rs2ValueU64 | memValueU64));
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[24] = memValueU64;
        expect.registers[5] = state.registers[5];
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amomind_succeeds() public {
        uint64 addr = 0x1f817b9eab194b0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x10, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 23, funct3, 22, 26, funct7); // amomind x23, x26, (x22)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"7516bf1e13664902", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"67451a124eddc883", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[22] = addr;
        state.registers[26] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of min(M[x[rs1]], x[rs2])
        bytes32 result = int64(rs2ValueU64) < int64(memValueU64) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[23] = memValueU64;
        expect.registers[22] = state.registers[22];
        expect.registers[26] = state.registers[26];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amomaxd_succeeds() public {
        uint64 addr = 0xf41e050aeffd9db0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x14, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 11, funct3, 14, 5, funct7); // amomaxd x11, x5, (x14)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"492c4fe3bf27bf82", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"95066b4c26a3e36c", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[14] = addr;
        state.registers[5] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of max(M[x[rs1]], x[rs2])
        bytes32 result = int64(rs2ValueU64) > int64(memValueU64) ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[11] = memValueU64;
        expect.registers[14] = state.registers[14];
        expect.registers[5] = state.registers[5];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amominud_succeeds() public {
        uint64 addr = 0xe094be571f4baca0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x18, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 31, funct3, 27, 30, funct7); // amominud x31, x30, (x27)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"18b0d1bf989c1b15", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"7ef1928fb292c2dd", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[27] = addr;
        state.registers[30] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of min(unsigned M[x[rs1]], unsigned x[rs2])
        bytes32 result = rs2ValueU64 < memValueU64 ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[31] = memValueU64;
        expect.registers[27] = state.registers[27];
        expect.registers[30] = state.registers[30];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_amomaxud_succeeds() public {
        uint64 addr = 0x2bcfe03b376a17e0;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x1c, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 26, funct3, 27, 6, funct7); // amomaxud x26, x6, (x27)
        }
        (bytes32 rs2ValueBytes32, uint64 rs2ValueU64) = truncate(hex"d679169ee3efcd97", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"5004c91ce741d398", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[27] = addr;
        state.registers[6] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of max(unsigned M[x[rs1]], unsigned x[rs2])
        bytes32 result = rs2ValueU64 > memValueU64 ? rs2ValueBytes32 : memValueBytes32;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[26] = memValueU64;
        expect.registers[27] = state.registers[27];
        expect.registers[6] = state.registers[6];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* I Type instructions */

    function test_addi_succeeds() public {
        uint16 imm = 0x373;
        uint32 insn = encodeIType(0x13, 26, 0, 25, imm); // addi x26, x25, 0x373
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[25] = 0xedf0;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[26] = state.registers[25] + imm;
        expect.registers[25] = state.registers[25];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_slli_succeeds() public {
        uint16 imm = 0x12;
        uint32 insn = encodeIType(0x13, 6, 1, 28, imm); // slli x6, x28, 0x12
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[28] = 0x5b03;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[6] = state.registers[28] << imm;
        expect.registers[28] = state.registers[28];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_slti_succeeds() public {
        uint16 imm = 0x54e;
        uint32 insn = encodeIType(0x13, 20, 2, 19, imm); // slti x20, x19, 0x54e
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[19] = 0x58d3;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[20] = state.registers[19] < imm ? 1 : 0;
        expect.registers[19] = state.registers[19];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sltiu_succeeds() public {
        uint16 imm = 0x2f3;
        uint32 insn = encodeIType(0x13, 22, 3, 14, imm); // sltiu x22, x14, 0x2f3
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[14] = 0x54f3;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[22] = state.registers[14] < imm ? 1 : 0;
        expect.registers[14] = state.registers[14];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_xori_succeeds() public {
        uint16 imm = 0x719;
        uint32 insn = encodeIType(0x13, 28, 4, 14, imm); // xori x28, x14, 0x719
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[14] = 0x9bef;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[28] = state.registers[14] ^ imm;
        expect.registers[14] = state.registers[14];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_srli_succeeds() public {
        uint16 imm = 0x11;
        uint32 insn = encodeIType(0x13, 5, 5, 3, imm); // srli x5, x3, 0x11
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[3] = 0x76b7;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[5] = state.registers[3] >> imm;
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_srai_succeeds() public {
        uint16 shamt = 0xf;
        uint16 imm = (0x10 << 6) | 0xf;
        uint32 insn = encodeIType(0x13, 19, 5, 12, imm); // srai x19, x12, 0xf
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        // intentionally set MSB to 1 to check sign preservation
        state.registers[12] = 0xFF78_3323_1095_FFFF;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[19] = state.registers[12] >> shamt;
        bool signBit = (1 << 63) & (state.registers[12]) > 0;
        if (signBit) {
            uint64 signExtension = uint64(((1 << shamt) - 1) << (64 - shamt));
            expect.registers[19] |= signExtension;
        }
        expect.registers[12] = state.registers[12];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_ori_succeeds() public {
        uint16 imm = 0x41d;
        uint32 insn = encodeIType(0x13, 9, 6, 7, imm); // ori x9, x7, 0x41d
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[7] = 0x9269;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[9] = state.registers[7] | imm;
        expect.registers[7] = state.registers[7];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_andi_succeeds() public {
        uint16 imm = 0x466;
        uint32 insn = encodeIType(0x13, 13, 7, 12, imm); // andi x13, x12, 0x466
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[12] = 0x7f73;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[13] = state.registers[12] & imm;
        expect.registers[12] = state.registers[12];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_addiw_succeeds() public {
        uint16 imm = 0x1f3;
        uint32 insn = encodeIType(0x1b, 31, 0, 16, imm); // addiw x31, x16, 0x1f3
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[16] = 0x6b56;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[31] = (state.registers[16] + imm) & ((1 << 32) - 1);
        expect.registers[16] = state.registers[16];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_slliw_succeeds() public {
        uint16 shamt = 0x15;
        uint16 imm = (0 << 7) | shamt;
        uint32 insn = encodeIType(0x1b, 17, 1, 25, imm); // slliw x17, x25, 0x15
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[25] = 0xf956;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[17] = (state.registers[25] << shamt) & ((1 << 32) - 1);
        expect.registers[25] = state.registers[25];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_srliw_succeeds() public {
        uint16 shamt = 0x7;
        uint16 imm = (0 << 7) | shamt;
        uint32 insn = encodeIType(0x1b, 27, 5, 13, imm); // srliw x27, x13, 0x7
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[13] = 0x88ce;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[27] = (state.registers[13] >> shamt) & ((1 << 32) - 1);
        expect.registers[13] = state.registers[13];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sraiw_succeeds() public {
        uint16 shamt = 0x4;
        uint16 imm = (0x20 << 5) | shamt;
        uint32 insn = encodeIType(0x1b, 30, 5, 28, imm); // sraiw x30, x28, 0x4
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        // intentionally set MSB to 1 to check sign preservation
        state.registers[28] = 0xF6F7_1234;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[30] = (state.registers[28] >> shamt) & ((1 << 32) - 1);
        bool signBit = (1 << 31) & state.registers[28] > 0;
        if (signBit) {
            uint64 signExtension = uint64(((1 << (32 + shamt)) - 1) << (32 - shamt));
            expect.registers[30] |= signExtension;
        }

        expect.registers[28] = state.registers[28];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_jalr_succeeds() public {
        uint16 imm = 0xc;
        uint32 insn = encodeIType(0x67, 15, 0, 3, imm); // jalr x15, x3, 0xc
        uint64 pc = 0x4000;
        (State memory state, bytes memory proof) = constructRISCVState(pc, insn);
        state.registers[15] = 0x1337;
        state.registers[3] = 0x3331;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.registers[15] = state.pc + 4;
        expect.pc = state.registers[3] + imm;
        // pc's LSB is set to zero
        expect.pc -= expect.pc & 1;
        expect.step = state.step + 1;
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_ecall_succeeds() public {
        // some syscalls are not supported
        // lets choose unsupported syscall clone just for testing functionality
        uint16 imm = 0x0;
        uint32 insn = encodeIType(0x73, 0, 0, 0, imm); // ecall
        uint64 pc = 0x1337;
        (State memory state, bytes memory proof) = constructRISCVState(pc, insn);
        state.registers[17] = 220; // syscall number of clone
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[10] = 1;
        expect.registers[11] = 0;
        expect.registers[17] = state.registers[17];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_ebreak_succeeds() public {
        // ebreak is ignored
        uint16 imm = 0x1;
        uint32 insn = encodeIType(0x73, 0, 0, 0, imm); // ebreak
        uint64 pc = 0x4004;
        (State memory state, bytes memory proof) = constructRISCVState(pc, insn);
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lb_succeeds() public {
        bytes32 value = hex"f9d9c609ec075104";
        uint16 offset = 0x376;
        uint64 addr = 0x98be + offset;
        uint8 funct3 = 0;
        uint32 insn = encodeIType(0x3, 3, funct3, 5, offset); // lb x3, offset(x5)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[5] = 0x98be;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[3] = uint8(value[0]);
        bool signBit = (1 << 7) & expect.registers[3] > 0;
        if (signBit) {
            uint64 size = uint64(1 << (funct3 & 0x3)) * 8;
            uint64 signExtension = uint64(((1 << 64 - size)) - 1) << size;
            expect.registers[3] |= signExtension;
        }
        expect.registers[5] = state.registers[5];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lh_succeeds() public {
        bytes32 value = hex"d7ffe3d2157e8954";
        uint16 offset = 0x358;
        uint64 addr = 0xc9b + offset;
        uint8 funct3 = 1;
        uint32 insn = encodeIType(0x3, 17, funct3, 15, offset); // lh x17, offset(x15)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[15] = 0xc9b;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[17] = bytes32ToUint64(value, 2);
        bool signBit = (1 << 15) & expect.registers[17] > 0;
        if (signBit) {
            uint64 size = uint64(1 << (funct3 & 0x3)) * 8;
            uint64 signExtension = uint64(((1 << 64 - size)) - 1) << size;
            expect.registers[17] |= signExtension;
        }
        expect.registers[15] = state.registers[15];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lw_succeeds() public {
        bytes32 value = hex"dcd20df1dbcda033";
        uint16 offset = 0x358;
        uint64 addr = 0xc9b + offset;
        uint8 funct3 = 2;
        uint32 insn = encodeIType(0x3, 27, funct3, 28, offset); // lw x27, offset(x28)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[28] = 0xc9b;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[27] = bytes32ToUint64(value, 4);
        bool signBit = (1 << 31) & expect.registers[27] > 0;
        if (signBit) {
            uint64 size = uint64(1 << (funct3 & 0x3)) * 8;
            uint64 signExtension = uint64(((1 << 64 - size)) - 1) << size;
            expect.registers[27] |= signExtension;
        }
        expect.registers[28] = state.registers[28];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_ld_succeeds() public {
        bytes32 value = hex"196faeb2a761c3f7";
        uint16 offset = 0x358;
        uint64 addr = 0xc9b + offset;
        uint32 insn = encodeIType(0x3, 3, 3, 15, offset); // ld x3, offset(x15)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[15] = 0xc9b;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // no need to sign extend
        expect.registers[3] = bytes32ToUint64(value, 8);
        expect.registers[15] = state.registers[15];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lbu_succeeds() public {
        bytes32 value = hex"f8d721d58e12f0bf";
        uint16 offset = 0x6bf;
        uint64 addr = 0xd34d + offset;
        uint32 insn = encodeIType(0x3, 9, 4, 25, offset); // lbu x9, offset(x25)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[25] = 0xd34d;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[9] = uint8(value[0]);
        expect.registers[25] = state.registers[25];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lhu_succeeds() public {
        bytes32 value = hex"59fb11d66dcc9d48";
        uint16 offset = 0x6bf;
        uint64 addr = 0xd34d + offset;
        uint32 insn = encodeIType(0x3, 21, 5, 4, offset); // lhu x21, offset(x4)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[4] = 0xd34d;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[21] = bytes32ToUint64(value, 2);
        expect.registers[4] = state.registers[4];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_lwu_succeeds() public {
        bytes32 value = hex"b02ec02db9d2ff8b";
        uint16 offset = 0x19;
        uint64 addr = 0x7bcc + offset;
        uint32 insn = encodeIType(0x3, 3, 6, 23, offset); // lwu x3, offset(x23)
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, value);
        state.registers[23] = 0x7bcc;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[3] = bytes32ToUint64(value, 4);
        expect.registers[23] = state.registers[23];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrw_succeeds() public {
        uint16 imm = 0x29a;
        uint32 insn = encodeIType(0x73, 13, 1, 2, imm); // csrrw x13, 0x29a, x2
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[2] = 0x4797;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[13] = 0; // CSR is not supported
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrs_succeeds() public {
        uint16 imm = 0x7cc;
        uint32 insn = encodeIType(0x73, 7, 2, 10, imm); // csrrs x7, 0x7cc, x10
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[10] = 0x8cdc;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[7] = 0; // CSR is not supported
        expect.registers[10] = state.registers[10];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrc_succeeds() public {
        uint16 imm = 0x567;
        uint32 insn = encodeIType(0x73, 1, 3, 25, imm); // csrrc x1, 0x567, x25
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[25] = 0xd088;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[1] = 0; // CSR is not supported
        expect.registers[25] = state.registers[25];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrwi_succeeds() public {
        uint16 imm = 0x3d;
        uint32 insn = encodeIType(0x73, 31, 5, 29, imm); // csrrwi x31, 0x3d, x29
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[29] = 0x398d;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[31] = 0; // CSR is not supported
        expect.registers[29] = state.registers[29];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrsi_succeeds() public {
        uint16 imm = 0x1;
        uint32 insn = encodeIType(0x73, 17, 6, 22, imm); // csrrsi x17, 0x1, x22
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[22] = 0x856a;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[17] = 0; // CSR is not supported
        expect.registers[22] = state.registers[22];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_csrrci_succeeds() public {
        uint16 imm = 0x2ca;
        uint32 insn = encodeIType(0x73, 23, 7, 18, imm); // csrrci x23, 0x2ca, x18
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[18] = 0xbeb3;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[23] = 0; // CSR is not supported
        expect.registers[18] = state.registers[18];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* S Type instructions */

    function test_sb_succeeds() public {
        uint16 imm = 0xe64;
        uint64 rs1Value = 0x6b856cf8;
        uint8 funct3 = 0;
        uint8 size = uint8(1 << (funct3 & 0x3));
        bytes32 value = hex"ce9a61c0068bd030";
        (bytes32 target, uint64 rs2Value) = truncate(value, size);
        bool signBit = (1 << 11) & imm > 0;
        uint64 addr = rs1Value + imm;
        if (signBit) {
            addr -= 1 << 12;
        }
        uint32 insn = encodeSType(0x23, funct3, 6, 3, imm); // sb x3, offset(x6)
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.registers[6] = rs1Value;
        state.registers[3] = rs2Value;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, target);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[6] = state.registers[6];
        expect.registers[3] = state.registers[3];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sh_succeeds() public {
        uint16 imm = 0xb51;
        uint64 rs1Value = 0x7ce31e7a;
        uint8 funct3 = 1;
        uint8 size = uint8(1 << (funct3 & 0x3));
        bytes32 value = hex"4f045df3ef2c2817";
        (bytes32 target, uint64 rs2Value) = truncate(value, size);
        bool signBit = (1 << 11) & imm > 0;
        uint64 addr = rs1Value + imm;
        if (signBit) {
            addr -= 1 << 12;
        }
        uint32 insn = encodeSType(0x23, funct3, 19, 25, imm); // sh x25, offset(x19)
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.registers[19] = rs1Value;
        state.registers[25] = rs2Value;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, target);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[19] = state.registers[19];
        expect.registers[25] = state.registers[25];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sw_succeeds() public {
        uint16 imm = 0xc04;
        uint64 rs1Value = 0xcb3053d5;
        uint8 funct3 = 2;
        uint8 size = uint8(1 << (funct3 & 0x3));
        bytes32 value = hex"43c10f060b84afdf";
        (bytes32 target, uint64 rs2Value) = truncate(value, size);
        bool signBit = (1 << 11) & imm > 0;
        uint64 addr = rs1Value + imm;
        if (signBit) {
            addr -= 1 << 12;
        }
        uint32 insn = encodeSType(0x23, funct3, 12, 29, imm); // sw x29, offset(x12)
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.registers[12] = rs1Value;
        state.registers[29] = rs2Value;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, target);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[12] = state.registers[12];
        expect.registers[29] = state.registers[29];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_sd_succeeds() public {
        uint16 imm = 0x431;
        uint64 rs1Value = 0x9ab94a99;
        uint8 funct3 = 3;
        uint8 size = uint8(1 << (funct3 & 0x3));
        bytes32 value = hex"5298cefada934bc7";
        (bytes32 target, uint64 rs2Value) = truncate(value, size);
        bool signBit = (1 << 11) & imm > 0;
        uint64 addr = rs1Value + imm;
        if (signBit) {
            addr -= 1 << 12;
        }
        uint32 insn = encodeSType(0x23, funct3, 1, 2, imm); // sd x2, offset(x1)
        // note. asterisc memory is zero-initialized.
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, 0);
        state.registers[1] = rs1Value;
        state.registers[2] = rs2Value;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, target);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        expect.registers[1] = state.registers[1];
        expect.registers[2] = state.registers[2];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* B Type instructions */

    function test_beq_succeeds() public {
        uint16 imm = 0x19cd;
        uint32 insn = encodeBType(0x63, 0, 23, 20, imm); // beq x23, x20, offset
        (State memory state, bytes memory proof) = constructRISCVState(0x139a, insn);
        state.registers[23] = 0x2152;
        state.registers[20] = 0x2152;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (state.registers[23] == state.registers[20]) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[23] = state.registers[23];
        expect.registers[20] = state.registers[20];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_bne_succeeds() public {
        uint16 imm = 0x1d7e;
        uint32 insn = encodeBType(0x63, 1, 20, 26, imm); // bne x20, x26, offset
        (State memory state, bytes memory proof) = constructRISCVState(0x1afc, insn);
        state.registers[20] = 0x14b6;
        state.registers[26] = 0x4156;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (state.registers[20] != state.registers[26]) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[20] = state.registers[20];
        expect.registers[26] = state.registers[26];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_blt_succeeds() public {
        uint16 imm = 0x664;
        uint32 insn = encodeBType(0x63, 4, 9, 19, imm); // blt x9, x19, offset
        (State memory state, bytes memory proof) = constructRISCVState(0xcc8, insn);
        state.registers[9] = 0xffffffff_ffff18af;
        state.registers[19] = 0x8e5e;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (int64(state.registers[9]) < int64(state.registers[19])) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[9] = state.registers[9];
        expect.registers[19] = state.registers[19];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_bge_succeeds() public {
        uint16 imm = 0x1350;
        uint32 insn = encodeBType(0x63, 5, 27, 11, imm); // bge x27, x11, offset
        (State memory state, bytes memory proof) = constructRISCVState(0x26a0, insn);
        state.registers[27] = 0xbad7;
        state.registers[11] = 0xffffffff_ffff5c1f;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (int64(state.registers[27]) >= int64(state.registers[11])) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[27] = state.registers[27];
        expect.registers[11] = state.registers[11];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_bltu_succeeds() public {
        uint16 imm = 0x171d;
        uint32 insn = encodeBType(0x63, 6, 13, 22, imm); // bltu x13, x22, offset
        (State memory state, bytes memory proof) = constructRISCVState(0x2e3a, insn);
        state.registers[13] = 0xa0cc;
        state.registers[22] = 0xffffffff_ffff795c;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (state.registers[13] < state.registers[22]) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[13] = state.registers[13];
        expect.registers[22] = state.registers[22];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_bgeu_succeeds() public {
        uint16 imm = 0x14b5;
        uint32 insn = encodeBType(0x63, 7, 7, 16, imm); // bgeu x7, x16, offset
        (State memory state, bytes memory proof) = constructRISCVState(0x296a, insn);
        state.registers[7] = 0xffffffff_ffff35e5;
        state.registers[16] = 0x7c3c;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc;
        if (state.registers[7] >= state.registers[16]) {
            expect.pc += imm - (imm & 1);
            bool signBit = (1 << 12) & imm > 0;
            if (signBit) {
                expect.pc -= 1 << 13;
            }
        } else {
            expect.pc += 4;
        }
        expect.step = state.step + 1;
        expect.registers[7] = state.registers[7];
        expect.registers[16] = state.registers[16];

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* U Type instructions */

    function test_lui_succeeds() public {
        uint32 imm = 0xd4638aaa;
        uint32 insn = encodeUType(0x37, 2, imm); // lui x2, imm
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 immSignExtended = (imm >> 12) << 12;
        bool signBit = (1 << 31) & imm > 0;
        if (signBit) {
            immSignExtended |= ((1 << 32) - 1) << 32;
        }
        expect.registers[2] = immSignExtended;
        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_auipc_succeeds() public {
        uint32 imm = 0xf00dcd79;
        uint32 insn = encodeUType(0x17, 7, imm); // auipc x7, imm
        uint64 pc = 0x9fbdc310; // 0x9fbdc319 fails
        (State memory state, bytes memory proof) = constructRISCVState(pc, insn);
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        uint64 immSignExtended = (imm >> 12) << 12;
        bool signBit = (1 << 31) & imm > 0;
        if (signBit) {
            immSignExtended |= ((1 << 32) - 1) << 32;
        }
        expect.registers[7] = uint64((uint128(immSignExtended) + pc) & ((1 << 64) - 1));

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* J Type instructions */

    function test_jal_succeeds() public {
        uint32 imm = 0xbef054ae;
        uint32 insn = encodeJType(0x6f, 5, imm); // jal x5, imm
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        bytes memory encodedState = encodeState(state);

        State memory expect;
        expect.memRoot = state.memRoot;
        expect.step = state.step + 1;
        uint64 offsetSignExtended = (imm & ((1 << 21) - 1)) - (imm & 1);
        bool signBit = (1 << 20) & imm > 0;
        if (signBit) {
            offsetSignExtended |= ((1 << (64 - 21)) - 1) << 21;
        }
        expect.registers[5] = state.pc + 4;
        expect.pc = uint64((uint128(offsetSignExtended) + state.pc) & ((1 << 64) - 1));

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* Syscalls */

    function test_preimage_read_succeeds() public {
        uint16 imm = 0x0;
        uint32 insn = encodeIType(0x73, 0, 0, 0, imm); // ecall
        uint64 pc = 0x1337;
        bytes32 value = hex"576b9d12b141a35a";
        uint64 addr = 0xd20ec023b82c68b2;
        (bytes32 memRoot, bytes memory proof) = ffi.getAsteriscMemoryProof(pc, insn, addr, value);

        uint64 size = 8;
        uint64[32] memory registers;
        registers[17] = 63; // syscall number of read
        registers[10] = 5; // A0 = fd: preimage read
        registers[11] = addr; // A1 = *buf addr
        registers[12] = size; // A2 = count

        State memory state = State({
            memRoot: memRoot,
            preimageKey: bytes32(uint256(1) << 248 | 0x01), // local key
            preimageOffset: 8, // start reading past the pre-image length prefix
            pc: pc,
            exitCode: 0,
            exited: false,
            step: 1,
            heap: 0,
            loadReservation: 0,
            registers: registers
        });
        bytes memory encodedState = encodeState(state);

        // prime the pre-image oracle
        bytes32 word = bytes32(uint256(0xdeadbeefcafebebe) << (256 - 32 * 2));
        uint8 partOffset = 8;
        oracle.loadLocalData(uint256(state.preimageKey), 0, word, size, partOffset);

        State memory expect = state;
        expect.preimageOffset += size;
        expect.pc += 4;
        expect.step += 1;
        expect.registers[10] = size; // return
        expect.registers[11] = 0; // error code
        // recompute merkle root of written pre-image
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(pc, insn, addr, word);

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    function test_preimage_write_succeeds() public {
        uint16 imm = 0x0;
        uint32 insn = encodeIType(0x73, 0, 0, 0, imm); // ecall
        uint64 pc = 0x1337;
        bytes32 value = hex"000056deb0bd0018";
        uint64 addr = 0xe3e356dce663a260;
        (bytes32 memRoot, bytes memory proof) = ffi.getAsteriscMemoryProof(pc, insn, addr, value);

        uint64 size = 7;
        uint64[32] memory registers;
        registers[17] = 64; // syscall number of write
        registers[10] = 6; // A0 = fd: preimage write
        registers[11] = addr; // A1 = *buf addr
        registers[12] = size; // A2 = count

        State memory state = State({
            memRoot: memRoot,
            preimageKey: bytes32(0),
            preimageOffset: 0x7331,
            pc: pc,
            exitCode: 0,
            exited: false,
            step: 1,
            heap: 0,
            loadReservation: 0,
            registers: registers
        });
        bytes memory encodedState = encodeState(state);

        State memory expect = state;
        expect.preimageOffset = 0; // preimage write resets offset
        expect.pc += 4;
        expect.step += 1;
        // vm memory layout: [LSB -> MSB]
        // [0x00, 0x00, 0x56, 0xde, 0xb0, 0xbd, 0x00, 0x18]
        // pre-image request size = 7 bytes
        // preimage key: [right padding with null] + [0x00, 0x00, 0x56, 0xde, 0xb0, 0xbd, 0x00]
        expect.preimageKey = hex"00000000000000000000000000000000000000000000000000000056deb0bd00";
        expect.registers[10] = size; // return
        expect.registers[11] = 0; // error code

        bytes32 postState = riscv.step(encodedState, proof, 0);
        assertEq(postState, outputState(expect), "unexpected post state");
    }

    /* Revert cases */

    function test_unknown_instruction() public {
        uint32 insn = encodeRType(0xff, 0, 0, 0, 0, 0); // 0xff is unknown instruction opcode
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        bytes memory encodedState = encodeState(state);

        vm.expectRevert(hex"00000000000000000000000000000000000000000000000000000000f001c0de");
        riscv.step(encodedState, proof, 0);
    }

    function test_invalid_proof() public {
        uint32 insn = encodeRType(0xff, 0, 0, 0, 0, 0);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        bytes memory encodedState = encodeState(state);
        proof = hex"00"; // Invalid memory proof

        vm.expectRevert(hex"00000000000000000000000000000000000000000000000000000000badf00d1");
        riscv.step(encodedState, proof, 0);
    }

    function test_unrecognized_resource_limit() public {
        uint16 imm = 0x0;
        uint32 insn = encodeIType(0x73, 0, 0, 0, imm); // ecall
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        state.registers[17] = 163;
        state.registers[10] = 0;
        bytes memory encodedState = encodeState(state);

        vm.expectRevert(hex"00000000000000000000000000000000000000000000000000000000000f0012");
        riscv.step(encodedState, proof, 0);
    }

    function test_invalid_amo_size() public {
        uint32 insn;
        uint8 funct3 = 0x1; // invalid amo size
        uint8 funct7 = encodeFunct7(0x0, 0x0, 0x0);
        insn = encodeRType(0x2f, 23, funct3, 17, 3, funct7);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn);
        bytes memory encodedState = encodeState(state);

        vm.expectRevert(hex"0000000000000000000000000000000000000000000000000000000000bada70");
        riscv.step(encodedState, proof, 0);
    }

    function test_unaligned_address() public {
        uint64 addr = 0xeae426a36ff2bb65; // unaligned address

        // Valid amoadd instr
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0x0, 0x0, 0x0);
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 14, funct3, 8, 28, funct7); // amoaddd x14, x28, (x8)
        }
        (, uint64 rs2ValueU64) = truncate(hex"a0821b98f6c0d237", size);
        (bytes32 memValueBytes32, uint64 memValueU64) = truncate(hex"f47daefa285404dc", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[8] = addr;
        state.registers[28] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        State memory expect;
        // check memory stores value of M[x[rs1]] + x[rs2]
        bytes32 result = uint256ToBytes32(uint256(uint128(int128(int64(rs2ValueU64)) + int128(int64(memValueU64)))));
        (expect.memRoot,) = ffi.getAsteriscMemoryProof(0, insn, addr, result);
        expect.pc = state.pc + 4;
        expect.step = state.step + 1;
        // check rd value stores original mem value.
        expect.registers[14] = memValueU64;
        expect.registers[8] = state.registers[8];
        expect.registers[28] = state.registers[28];

        vm.expectRevert(hex"00000000000000000000000000000000000000000000000000000000bad10ad0");
        riscv.step(encodedState, proof, 0);
    }

    function test_unknown_atomic_operation() public {
        uint64 addr = 0xeae426a36ff2bb68;
        uint32 insn;
        uint8 size;
        {
            uint8 funct3 = 0x3;
            uint8 funct7 = encodeFunct7(0xff, 0x0, 0x0); // unknown atomic operation
            size = uint8(1 << (funct3 & 0x3));
            insn = encodeRType(0x2f, 14, funct3, 8, 28, funct7); // amoaddd x14, x28, (x8)
        }
        (, uint64 rs2ValueU64) = truncate(hex"a0821b98f6c0d237", size);
        (bytes32 memValueBytes32,) = truncate(hex"f47daefa285404dc", size);
        (State memory state, bytes memory proof) = constructRISCVState(0, insn, addr, memValueBytes32);
        state.registers[8] = addr;
        state.registers[28] = rs2ValueU64;
        bytes memory encodedState = encodeState(state);

        vm.expectRevert(hex"000000000000000000000000000000000000000000000000000000000f001a70");
        riscv.step(encodedState, proof, 0);
    }

    /* Helper methods */

    function encodeState(State memory state) internal pure returns (bytes memory) {
        bytes memory registers;
        for (uint256 i = 0; i < state.registers.length; i++) {
            registers = bytes.concat(registers, abi.encodePacked(state.registers[i]));
        }
        bytes memory stateData = abi.encodePacked(
            state.memRoot,
            state.preimageKey,
            state.preimageOffset,
            state.pc,
            state.exitCode,
            state.exited,
            state.step,
            state.heap,
            state.loadReservation,
            registers
        );
        return stateData;
    }

    /// @dev RISCV VM status codes:
    ///      0. Exited with success (Valid)
    ///      1. Exited with success (Invalid)
    ///      2. Exited with failure (Panic)
    ///      3. Unfinished
    function vmStatus(State memory state) internal pure returns (VMStatus out_) {
        if (!state.exited) {
            return VMStatuses.UNFINISHED;
        } else if (state.exitCode == 0) {
            return VMStatuses.VALID;
        } else if (state.exitCode == 1) {
            return VMStatuses.INVALID;
        } else {
            return VMStatuses.PANIC;
        }
    }

    function outputState(State memory state) internal pure returns (bytes32 out_) {
        bytes memory enc = encodeState(state);
        VMStatus status = vmStatus(state);
        assembly {
            out_ := keccak256(add(enc, 0x20), 362)
            out_ := or(and(not(shl(248, 0xFF)), out_), shl(248, status))
        }
    }

    function constructRISCVState(
        uint64 pc,
        uint32 insn,
        uint64 addr,
        bytes32 val
    )
        internal
        returns (State memory state, bytes memory proof)
    {
        (state.memRoot, proof) = ffi.getAsteriscMemoryProof(pc, insn, addr, val);
        state.pc = pc;
    }

    function constructRISCVState(uint64 pc, uint32 insn) internal returns (State memory state, bytes memory proof) {
        (state.memRoot, proof) = ffi.getAsteriscMemoryProof(pc, insn);
        state.pc = pc;
    }

    function encodeRType(
        uint8 opcode,
        uint8 rd,
        uint8 funct3,
        uint8 rs1,
        uint8 rs2,
        uint8 funct7
    )
        internal
        pure
        returns (uint32 insn)
    {
        // insn := [funct7] | [rs2] | [rs1] | [funct3] | [rd]  | [opcode]
        // example: 0000000 | 00011 | 00010 | 000      | 00001 | 0110011
        insn = uint32(funct7 & 0x7F) << (7 + 5 + 3 + 5 + 5);
        insn |= uint32(rs2 & 0x1F) << (7 + 5 + 3 + 5);
        insn |= uint32(rs1 & 0x1F) << (7 + 5 + 3);
        insn |= uint32(funct3 & 0x7) << (7 + 5);
        insn |= uint32(rd & 0x1F) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeIType(
        uint8 opcode,
        uint8 rd,
        uint8 funct3,
        uint8 rs1,
        uint16 imm
    )
        internal
        pure
        returns (uint32 insn)
    {
        // insn   := [imm[11:0]]  | [rs1] | [funct3] | [rd]  | [opcode]
        // example:  000000000111 | 00101 | 000      | 00110 | 0010011
        insn = uint32(imm & 0xFFF) << (7 + 5 + 3 + 5);
        insn |= uint32(rs1 & 0x1F) << (7 + 5 + 3);
        insn |= uint32(funct3 & 0x7) << (7 + 5);
        insn |= uint32(rd & 0x1F) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeSType(
        uint8 opcode,
        uint8 funct3,
        uint8 rs1,
        uint8 rs2,
        uint16 imm
    )
        internal
        pure
        returns (uint32 insn)
    {
        // insn   := [imm[11:5]]| [rs2] | [rs1] | [funct3] | [imm[4:0]] | [opcode]
        // example:  0001010    | 01011 | 00001 | 000      | 11011      | 0100011
        insn = uint32((imm >> 5) & 0x7F) << (7 + 5 + 3 + 5 + 5);
        insn |= uint32(rs2 & 0x1F) << (7 + 5 + 3 + 5);
        insn |= uint32(rs1 & 0x1F) << (7 + 5 + 3);
        insn |= uint32(funct3 & 0x7) << (7 + 5);
        insn |= uint32(imm & 0x1F) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeBType(
        uint8 opcode,
        uint8 funct3,
        uint8 rs1,
        uint8 rs2,
        uint16 imm
    )
        internal
        pure
        returns (uint32 insn)
    {
        // we lose information of lsb of imm, assuming always zero
        // insn   := [imm[12]] | [imm[10:5]] | [rs2] | [rs1] | [funct3] | [imm[4:1]] | imm[11] | [opcode]
        // example:  0         | 010101      | 01100 | 00100 | 000      | 0010       | 1       | 1100011
        insn = uint32((imm >> 12) & 0x1) << (7 + 1 + 4 + 3 + 5 + 5 + 6);
        insn |= uint32((imm >> 5) & 0x3f) << (7 + 1 + 4 + 3 + 5 + 5);
        insn |= uint32(rs2 & 0x1F) << (7 + 1 + 4 + 3 + 5);
        insn |= uint32(rs1 & 0x1F) << (7 + 1 + 4 + 3);
        insn |= uint32(funct3 & 0x7) << (7 + 1 + 4);
        insn |= uint32((imm >> 1) & 0xF) << (7 + 1);
        insn |= uint32((imm >> 11) & 0x1) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeUType(uint8 opcode, uint8 rd, uint32 imm) internal pure returns (uint32 insn) {
        // insn   := [imm[31:12]]         | [rd]  | [opcode]
        // example:  00110010101010000001 | 01011 | 0110111
        insn = uint32((imm >> 12) & 0xFFFFF) << (7 + 5);
        insn |= uint32(rd & 0x1F) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeJType(uint8 opcode, uint8 rd, uint32 imm) internal pure returns (uint32 insn) {
        // insn   := [imm[20|10:1|11|19:12]] | [rd]  | [opcode]
        // example:  00110010101010000001    | 00101 | 1101111
        insn = uint32((imm >> 20) & 0x1) << (7 + 5 + 8 + 1 + 10);
        insn |= uint32((imm >> 1) & 0x3ff) << (7 + 5 + 8 + 1);
        insn |= uint32((imm >> 11) & 0x1) << (7 + 5 + 8);
        insn |= uint32((imm >> 12) & 0xFF) << (7 + 5);
        insn |= uint32(rd & 0x1F) << 7;
        insn |= uint32(opcode & 0x7F);
    }

    function encodeFunct7(uint8 funct5, uint8 aq, uint8 rl) internal pure returns (uint8 funct7) {
        // funct7  := [funct5] | [aq] | [rl]
        // example :  01100    | 0    | 0
        funct7 = (funct5 & 0x1f) << 2;
        funct7 |= (aq & 0x1) << 1;
        funct7 |= rl & 0x1;
    }

    function mask32Signed64(uint64 val) internal pure returns (uint64) {
        uint64 result = mask32Unsigned64(val);
        if ((1 << 31) & result > 0) {
            result |= uint64(((1 << 32) - 1) << 32);
        }
        return result;
    }

    function mask32Unsigned64(uint64 val) internal pure returns (uint64) {
        return uint64(val & ((1 << 32) - 1));
    }

    function uint256ToBytes32(uint256 val) internal pure returns (bytes32) {
        // we cannot use direct casting using bytes32 because of endianess
        uint256 temp = 0;
        for (uint8 i = 0; i < 32; i++) {
            temp += uint256((val >> (8 * i)) & 0xFF) << (256 - 8 - 8 * i);
        }
        return bytes32(temp);
    }

    function truncate(bytes32 val, uint8 size) internal pure returns (bytes32 valueBytes32, uint64 valueU64) {
        valueU64 = bytes32ToUint64(val, size);
        uint256 temp = 0;
        for (uint8 i = 0; i < size; i++) {
            temp += uint8(val[i]) * uint256(1 << (256 - 8 - 8 * i));
        }
        valueBytes32 = bytes32(temp);
    }

    function bytes32ToUint64(bytes32 val, uint8 size) internal pure returns (uint64 result) {
        for (uint8 i = 0; i < size; i++) {
            result += uint8(val[i]) * uint64(1 << 8 * i);
        }
    }
}
