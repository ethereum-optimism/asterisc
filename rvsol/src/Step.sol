pragma solidity >=0.8.10 <0.9.0;

contract Step {
    //
    function step(bytes s) public pure view returns (bytes32 out) {
        assembly {
            function signExtend(v, bit) -> vs {
                switch and(v, shl(1, bit))
                case 0 {
                    // fill with zeroes, by masking
                    out := and(v, shr(0xFFFF_FFFF_FFFF_FFFF, sub(63, bit)))
                }
                default {
                    // fill with ones, by or-ing
                    out := or(v, shl(shr(0xFFFF_FFFF_FFFF_FFFF, bit), bit))
                }
            }
            function parseImmTypeI(instr) -> imm {
                imm := signExtend(shr(instr, 20), 11)
            }
            function parseImmTypeS(instr) -> imm {
                imm := signExtend(or(shl(shr(instr, 25), 5), and(shr(instr, 7), 0x1F)), 11)
            }
            function parseImmTypeB(instr) -> imm {
                imm := signExtend(
                    or(
                        or(
                            shl(and(shr(instr, 8), 0xF), 1),
                            shl(and(shr(instr, 25), 0x3F), 5)
                        ),
                        or(
                            shl(and(shr(instr, 7), 1), 11),
                            shl(shr(instr, 31), 12)
                        )
                    ),
                    12
                )
            }
            function parseImmTypeU(instr) -> imm {
                imm := signExtend(shl(shr(instr, 20), 12), 19)
            }
            function parseImmTypeJ(instr) -> imm {
                imm := signExtend(
                    or(
                        or(
                            shl(and(shr(instr, 21), 0x1FF), 1),
                            shl(and(shr(instr, 20), 1), 10)
                        ),
                        or(
                            shl(and(shr(instr, 12), 0xFF), 11),
                            shl(shr(instr, 31), 19)
                        )
                    ),
                    19
                )
            }

            // TODO: port over from Go:
            // - VM scratchpad as struct encoded in calldata
            // - state machine
            // - instruction phases
            // - opcode execution
        }
    }
}
