object "RISCV" {
    code {
        // Deploy the contract
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // Dispatcher
            switch selector()
            case 0xb00c8ce3 /* parseImmTypeI(uint64)->uint64*/ {
                returnUint(parseImmTypeI(toU64(decodeAsUint64(0))))
            }
            case 0xbf5b9caa /* parseImmTypeS(uint64)->uint256*/ {
                returnUint(parseImmTypeS(decodeAsUint64(0)))
            }
            case 0x40233f6d /* parseImmTypeB(uint64)->uint256*/{
                returnUint(parseImmTypeB(decodeAsUint64(0)))
            }
            case 0x9039dd19 /* parseImmTypeU(uint64)->uint256*/ {
                returnUint(parseImmTypeU(decodeAsUint64(0)))
            }
            case 0x6933e90c /* parseImmTypeJ(uint64)->uint256*/ {
                returnUint(parseImmTypeJ(decodeAsUint64(0)))
            }
            case 0xb488c140 /* ParseOpcode(uint64)returns(uint256)" */ {
                returnUint(parseOpcode(decodeAsUint64(0)))
            }
            case 0x2c8fcf96 /* ParseRd(uint64)returns(uint256) */ {
                returnUint(parseRd(decodeAsUint64(0)))
            }
            case 0x0596de79 /* ParseFunct3(uint64)returns(uint256) */ {
                returnUint(parseFunct3(decodeAsUint64(0)))
            }
            case 0xa2494672 /* ParseRs1(uint64)returns(uint256)*/ {
                returnUint(parseRs1(decodeAsUint64(0)))
            }
            case 0xb3bc5703 /* ParseRs2(uint64)returns(uint256) */ {
                returnUint(parseRs2(decodeAsUint64(0)))
            }
            case 0xf80141d6 /* ParseFunct7(uint64)returns(uint256) */ {
                returnUint(parseFunct7(decodeAsUint64(0)))
            }
            default {
                revert(0, 0)
            }


            /* ---------- calldata decoding functions ----------- */
            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }

            function decodeAsUint64(offset) -> v {
                let pos := add(4, mul(offset, 0x8))  // We use 0x8 instead of 0x20 since uint64 is 8 bytes
                if lt(calldatasize(), add(pos, 0x8)) {  // Check if the calldata size is sufficient for 8 bytes
                    revert(0, 0)
                }
                v := calldataload(pos)  // Load 8 bytes (uint64)
                v := and(v, 0xFFFFFFFFFFFFFFFF)  // Mask to ensure we only get the lower 8 bytes (uint64)
            }
            /* ---------- calldata encoding functions ---------- */
            function returnUint(v) {
                mstore(0, v)
                return(0, 0x20)
            }

            /* ---------- Parse - functions to parse RISC-V instructions ---------- */
            /* ---------- maps to parse.go in golang  ---------- */
            function parseImmTypeI(instr) -> out {
                out := signExtend64(shr64(toU64(20), instr), toU64(11))
            }

            function parseImmTypeS(instr) -> out {
                out :=
                    signExtend64(
                        or64(shl64(toU64(5), shr64(toU64(25), instr)), and64(shr64(toU64(7), instr), toU64(0x1F))),
                        toU64(11)
                    )
            }

            function parseImmTypeB(instr) -> out {
                out :=
                    signExtend64(
                        or64(
                            or64(
                                shl64(toU64(1), and64(shr64(toU64(8), instr), toU64(0xF))),
                                shl64(toU64(5), and64(shr64(toU64(25), instr), toU64(0x3F)))
                            ),
                            or64(
                                shl64(toU64(11), and64(shr64(toU64(7), instr), toU64(1))),
                                shl64(toU64(12), shr64(toU64(31), instr))
                            )
                        ),
                        toU64(12)
                    )
            }

            function parseImmTypeU(instr) -> out {
                out := signExtend64(shr64(toU64(12), instr), toU64(19))
            }

            function parseImmTypeJ(instr) -> out {
                out :=
                    signExtend64(
                        or64(
                            or64(
                                and64(shr64(toU64(21), instr), shortToU64(0x3FF)), // 10 bits for index 0:9
                                shl64(toU64(10), and64(shr64(toU64(20), instr), toU64(1))) // 1 bit for index 10
                            ),
                            or64(
                                shl64(toU64(11), and64(shr64(toU64(12), instr), toU64(0xFF))), // 8 bits for index 11:18
                                shl64(toU64(19), shr64(toU64(31), instr)) // 1 bit for index 19
                            )
                        ),
                        toU64(19)
                    )
            }

            function parseOpcode(instr) -> out {
                out := and64(instr, toU64(0x7F))
            }

            function parseRd(instr) -> out {
                out := and64(shr64(toU64(7), instr), toU64(0x1F))
            }

            function parseFunct3(instr) -> out {
                out := and64(shr64(toU64(12), instr), toU64(0x7))
            }

            function parseRs1(instr) -> out {
                out := and64(shr64(toU64(15), instr), toU64(0x1F))
            }

            function parseRs2(instr) -> out {
                out := and64(shr64(toU64(20), instr), toU64(0x1F))
            }

            function parseFunct7(instr) -> out {
                out := shr64(toU64(25), instr)
            }

            /* ---------- Yul64 - functions to implement yul ---------- */
            /* ---------- maps to yul64.go in golang  ---------- */
            function u64Mask() -> out {
                // max uint64
                out := shr(192, not(0)) // 256-64 = 192
            }

            function u32Mask() -> out {
                out := U64(shr(toU256(224), not(0))) // 256-32 = 224
            }

            function toU64(v) -> out {
                out := v
            }

            function shortToU64(v) -> out {
                out := v
            }

            function shortToU256(v) -> out {
                out := v
            }

            function longToU256(v) -> out {
                out := v
            }

            function u256ToU64(v) -> out {
                out := and(v, U256(u64Mask()))
            }

            function u64ToU256(v) -> out {
                out := v
            }

            function mask32Signed64(v) -> out {
                out := signExtend64(and64(v, u32Mask()), toU64(31))
            }

            function u64Mod() -> out {
                // 1 << 64
                out := shl(toU256(64), toU256(1))
            }

            function u64TopBit() -> out {
                // 1 << 63
                out := shl(toU256(63), toU256(1))
            }

            function signExtend64(v, bit) -> out {
                switch and(v, shl(bit, 1))
                case 0 {
                    // fill with zeroes, by masking
                    out := U64(and(U256(v), shr(sub(toU256(63), bit), U256(u64Mask()))))
                }
                default {
                    // fill with ones, by or-ing
                    out := U64(or(U256(v), shl(bit, shr(bit, U256(u64Mask())))))
                }
            }

            function signExtend64To256(v) -> out {
                switch and(U256(v), u64TopBit())
                case 0 { out := v }
                default { out := or(shl(toU256(64), not(0)), v) }
            }

            function add64(x, y) -> out {
                out := U64(mod(add(U256(x), U256(y)), u64Mod()))
            }

            function sub64(x, y) -> out {
                out := U64(mod(sub(U256(x), U256(y)), u64Mod()))
            }

            function mul64(x, y) -> out {
                out := u256ToU64(mul(U256(x), U256(y)))
            }

            function div64(x, y) -> out {
                out := u256ToU64(div(U256(x), U256(y)))
            }

            function sdiv64(x, y) -> out {
                // note: signed overflow semantics are the same between Go and EVM assembly
                out := u256ToU64(sdiv(signExtend64To256(x), signExtend64To256(y)))
            }

            function mod64(x, y) -> out {
                out := U64(mod(U256(x), U256(y)))
            }

            function smod64(x, y) -> out {
                out := u256ToU64(smod(signExtend64To256(x), signExtend64To256(y)))
            }

            function not64(x) -> out {
                out := u256ToU64(not(U256(x)))
            }

            function lt64(x, y) -> out {
                out := U64(lt(U256(x), U256(y)))
            }

            function gt64(x, y) -> out {
                out := U64(gt(U256(x), U256(y)))
            }

            function slt64(x, y) -> out {
                out := U64(slt(signExtend64To256(x), signExtend64To256(y)))
            }

            function sgt64(x, y) -> out {
                out := U64(sgt(signExtend64To256(x), signExtend64To256(y)))
            }

            function eq64(x, y) -> out {
                out := U64(eq(U256(x), U256(y)))
            }

            function iszero64(x) -> out {
                out := iszero(U256(x))
            }

            function and64(x, y) -> out {
                out := U64(and(U256(x), U256(y)))
            }

            function or64(x, y) -> out {
                out := U64(or(U256(x), U256(y)))
            }

            function xor64(x, y) -> out {
                out := U64(xor(U256(x), U256(y)))
            }

            function shl64(x, y) -> out {
                out := u256ToU64(shl(U256(x), U256(y)))
            }

            function shr64(x, y) -> out {
                out := U64(shr(U256(x), U256(y)))
            }

            function sar64(x, y) -> out {
                out := u256ToU64(sar(U256(x), signExtend64To256(y)))
            }

            // type casts, no-op in yul
            function b32asBEWord(v) -> out {
                out := v
            }
            function beWordAsB32(v) -> out {
                out := v
            }
            function U64(v) -> out {
                out := v
            }
            function U256(v) -> out {
                out := v
            }
            function toU256(v) -> out {
                out := v
            }
        } 
    } 
}