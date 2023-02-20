// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;


contract Step {

    // sub index steps
    uint256 immutable StepLoadPC          = 0;
    uint256 immutable StepLoadInstr       = 1;
    uint256 immutable StepLoadRs1         = 2;
    uint256 immutable StepLoadRs2         = 3;
    uint256 immutable StepOpcode          = 4;
    uint256 immutable StepLoadSyscallArgs = 5;
    uint256 immutable StepRunSyscall      = 6;
    uint256 immutable StepWriteSyscallRet = 7;
    uint256 immutable StepWriteRd         = 8;
    uint256 immutable StepWritePC         = 9;
    uint256 immutable StepFinal           = 10;

    // tree:
    // ```
    //
    //	         1
    //	    2          3
    //	 4    5     6     7
    //	8 9 10 11 12 13 14 15
    //
    // ```
    uint256 immutable pcGindex        = 8;
    uint256 immutable memoryGindex    = 9;
    uint256 immutable registersGindex = 10;
    uint256 immutable csrGindex       = 11;
    uint256 immutable exitGindex      = 12;
    uint256 immutable heapGindex      = 13;

    // Writing destinations
    uint256 immutable destNone     = 0;
    uint256 immutable destPc       = 1;
    uint256 immutable destInstr    = 2;
    uint256 immutable destRs1Value = 3;
    uint256 immutable destRs2Value = 4;
    uint256 immutable destRdvalue  = 5;
    uint256 immutable destSysReg   = 6;
    uint256 immutable destHeapIncr = 7;
    uint256 immutable destCSRRW    = 8;
    uint256 immutable destCSRRS    = 9;
    uint256 immutable destCSRRC    = 10;


    struct VMSubState {
        uint64 SubIndex; // step in the instruction execution
        uint64 PC; // PC counter

        uint64 Instr; // raw instruction

        uint64 RdValue; // destination register value to write
        uint64 Rs1Value; // loaded source registers. Only load if rs1/rs2 are not zero.
        uint64 Rs2Value; //

        uint8 SyscallArgsI;
        uint64[8] SyscallRegs; // load up to 6 syscall register args: A0,A1,A2,A3,A4,A5 and A6 (FID), A7 (EID)

        bytes32 StateRoot; // commits to state as structured binary tree

        // State machine
        uint256 StateGindex; // target
        bytes32 StateStackHash; // hash of previous traversed stack and last stack element
        uint256 StateStackGindex; // to navigate the state loading/writing
        uint8 StateStackDepth;
        bytes32 StateValue;

        bytes8 Scratch;

        uint256 Gindex1; // first leaf to read from or write to
        uint256 Gindex2; // second leaf to read from or write to
        uint8 Offset; // offset: value might start anywhere in Gindex1 leaf, and overflow to Gindex2 leaf
        uint64 Value; // value to write
        bool Signed; // if value to read should be sign-extended
        uint8 Size; // size of value to read, may be 1, 2, 4 or 8 bytes
        uint64 Dest; // destination to load a value back into
    }

    function soGet(bytes32 key, bytes d) public pure view returns (bytes32 a, bytes32 b) {
        require(d.length == 64);
        require(keccak256(d) == key);
        return abi.decode(d, (bytes32, bytes32));
    }

    function soRemember(bytes32 a, bytes32 b) public pure view returns (bytes32 h) {
        // TODO: we can event-log the (a,b) so we can fill the state-oracle with rvsol like with rvgo
        return keccak256(a, b);
    }

    //
    function step(bytes calldata _s, bytes calldata soData) public pure view returns (VMSubState memory s) {
        s = abi.decode(_s, (VMSubState));

        if (s.StateStackGindex < s.StateGindex) {
            // READING MODE: if the stack gindex is lower than target, then traverse to target
            if (s.StateStackGindex == 1) {
                s.StateValue = s.StateRoot;
            }
            s.StateStackGindex = s.StateStackGindex << 1;
            a, b := soGet(s.StateValue, soData);
            if (((s.StateGindex >> s.StateStackDepth) & 1) != uint256(0)) {
                s.StateStackGindex = s.StateStackGindex | uint256(1);
                s.StateValue = b;
                // keep track of where we have been, to use the trail to go back up the stack when writing
                s.StateStackHash = soRemember(s.StateStackHash, a);
            } else {
                s.StateValue = a;
                // keep track of where we have been, to use the trail to go back up the stack when writing
                s.StateStackHash = soRemember(s.StateStackHash, b);
            }
            s.StateStackDepth -= 1;
            return;
        } else if (s.StateStackGindex > s.StateGindex) {
            // WRITING MODE: if the stack gindex is higher than the target, then traverse back to root and update along the way
            prevStackHash, prevSibling := soGet(s.StateStackHash, soData);
            s.StateStackHash = prevStackHash;
            if ((s.StateStackGindex & uint256(1)) != uint256(0)) {
                s.StateValue = soRemember(prevSibling, s.StateValue);
            } else {
                s.StateValue = soRemember(s.StateValue, prevSibling);
            }
            s.StateStackGindex = s.StateStackGindex >> uint256(1);
            if (s.StateStackGindex == uint256(1)) {
                s.StateRoot = s.StateValue;
            }
            return;
        }

        // if we want to read/write a value at the given gindex, then try.
        if (s.Gindex1 != uint256(0)) {
            if (s.Gindex1 != s.StateGindex) {
                // if we have not reached the gindex yet, then we need to start traversal to it
                s.StateStackGindex = uint256(1);
                s.StateStackDepth = uint8(s.Gindex1.BitLen()) - 2;
                s.StateGindex = s.Gindex1;
            } else {
                s.Gindex1 = uint256(0)

                if (s.Dest == destCSRRW) {
                    // special case: CSRRW - read and write bits
                    s.RdValue = decodeU64(s.StateValue[:8])
                    s.Dest = destNone
                } else if (s.Dest == destCSRRS) {
                    // special case: CSRRS - read and set bits
                    s.RdValue = decodeU64(s.StateValue[:8])
                    s.Value = or64(s.RdValue, s.Value) // set bits
                    s.Dest = destNone
                } else if (s.Dest == destCSRRC) {
                    // special case: CSRRC - read and clear bits
                    s.RdValue = decodeU64(s.StateValue[:8])
                    s.Value = and64(s.RdValue, not64(s.Value)) // clear bits
                    s.Dest = destNone
                } else if (s.Dest == destHeapIncr) {
                    // special case: increment before writing, and remember as first syscall reg
                    s.Value = add64(s.Value, decodeU64(s.StateValue[:8]))
                    s.SyscallRegs[0] = s.Value
                    s.Dest = destNone
                }

                // we reached the value, now load/write it
                if (s.Dest == destNone) { // writing
                    // note: StateValue holds the old 32 bytes, some of which may stay the same
                    v := encodePacked(s.Value);
                    copy(s.StateValue[s.Offset:], v[:s.Size])
                    s.StateGindex = uint256(1);
                } else { // reading
                    copy(s.Scratch[:], s.StateValue[s.Offset:])
                }
            }
            return;
        }

        if (s.Gindex2 != uint256(0)) {
            if (s.Gindex2 != s.StateGindex) {
                // if we have not reached the gindex yet, then we need to start traversal to it
                s.StateStackGindex = toU256(1);
                s.StateStackDepth = uint8(s.Gindex2.BitLen()) - 2;
                s.StateGindex = s.Gindex2;
            } else {
                s.Gindex2 = uint256(0);

                firstChunkBytes := 32 - s.Offset;
                if (firstChunkBytes > s.Size) {
                    firstChunkBytes = s.Size
                }
                secondChunkBytes := s.Size - firstChunkBytes

                // we reached the value, now load/write it
                if (s.Dest == destNone) { // writing
                    // note: StateValue holds the old 32 bytes, some of which may stay the same
                    v := encodePacked(s.Value)
                    copy(s.StateValue[:secondChunkBytes], v[firstChunkBytes:s.Size])
                    s.StateGindex = toU256(1)
                } else { // reading
                    copy(s.Scratch[firstChunkBytes:], s.StateValue[:secondChunkBytes])
                }
            }
            return;
        }

        if (s.Dest != destNone) { // complete reading work if any
            val := decodeU64(s.Scratch[:s.Size]);

            if (s.Signed) {
                topBitIndex := (s.Size << 3) - 1;
                val = signExtend64(val, toU64(topBitIndex));
            }

            switch s.Dest {
            case destPc:
                s.PC = val
            case destInstr:
                s.Instr = val
            case destRs1Value:
                s.Rs1Value = val
            case destRs2Value:
                s.Rs2Value = val
            case destRdvalue:
                s.RdValue = val
            case destSysReg:
                s.SyscallRegs[s.SyscallArgsI] = val
                s.SyscallArgsI += 1
            }
            s.Dest = destNone
            return;
        }

        assembly {
            // type casts, no-op in yul
            function U64(v) -> out {
                out := v
            }
            function U256(v) -> out {
                out := v
            }

            //
            // Yul64 - functions to do 64 bit math - see yul64.go
            //
            function u64Mask() -> out { // max uint64
                out := shr(not(0), 192) // 256-64 = 192
            }

            function u32Mask() -> out {
                out := U64(shr(not(0), toU256(224))) // 256-32 = 224
            }

            function toU64(v) -> out {
                out := v
            }

            function shortToU64(v) -> out {
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

            function u64Mod() U256 { // 1 << 64
                out := shl(toU256(1), toU256(64))
            }

            function u64TopBit() U256 { // 1 << 63
                out := shl(toU256(1), toU256(63))
            }

            function signExtend64(v, bit) -> out {
                switch and(v, shl(1, bit))
                case 0 {
                    // fill with zeroes, by masking
                    out := U64(and(U256(v), shr(U256(u64Mask()), sub(toU256(63), bit))))
                }
                default {
                    // fill with ones, by or-ing
                    out := U64(or(U256(v), shl(shr(U256(u64Mask()), bit), bit)))
                }
            }

            function signExtend64To256(v U64) -> out {
                switch and(U256(v), u64TopBit())
                case 0 {
                    out := v
                }
                default {
                    out := or(shl(not(U256{}), toU256(64)), v)
                }
            }

            function add64(x, y) -> out {
                out := U64(mod(add(U256(x), y), u64Mod()))
            }

            function sub64(x, y) -> out {
                out := U64(mod(sub(U256(x), y), u64Mod()))
            }

            function mul64(x, y) -> out {
                out := u256ToU64(mul(U256(x), y))
            }

            function div64(x, y) -> out {
                out := u256ToU64(div(U256(x), y))
            }

            function sdiv64(x, y) -> out { // note: signed overflow semantics are the same between Go and EVM assembly
                out := u256ToU64(sdiv(signExtend64To256(x), signExtend64To256(y)))
            }

            //
            // Parse - functions to parse RISC-V instructions - see parse.go
            //
            function parseImmTypeI(instr) -> out {
                out := signExtend64(shr64(instr, toU64(20)), toU64(11))
            }

            function parseImmTypeS(instr) -> out {
                out := signExtend64(or64(shl64(shr64(instr, toU64(25)), toU64(5)), and64(shr64(instr, toU64(7)), toU64(0x1F))), toU64(11))
            }

            function parseImmTypeB(instr) -> out {
                out := signExtend64(
                    or64(
                        or64(
                            shl64(and64(shr64(instr, toU64(8)), toU64(0xF)), toU64(1)),
                            shl64(and64(shr64(instr, toU64(25)), toU64(0x3F)), toU64(5)),
                        ),
                        or64(
                            shl64(and64(shr64(instr, toU64(7)), toU64(1)), toU64(11)),
                            shl64(shr64(instr, toU64(31)), toU64(12)),
                        )
                    ),
                    toU64(12)
                )
            }

            function parseImmTypeU(instr) -> out {
                out := signExtend64(shr64(instr, toU64(12)), toU64(19))
            }

            function parseImmTypeJ(instr) -> out {
                out := signExtend64(
                    or64(
                        or64(
                            shl64(and64(shr64(instr, toU64(21)), shortToU64(0x1FF)), toU64(1)),
                            shl64(and64(shr64(instr, toU64(20)), toU64(1)), toU64(10)),
                        ),
                        or64(
                            shl64(and64(shr64(instr, toU64(12)), toU64(0xFF)), toU64(11)),
                            shl64(shr64(instr, toU64(31)), toU64(19)),
                        )
                    ),
                    toU64(19)
                )
            }

            function parseOpcode(instr) -> out {
                out := and64(instr, toU64(0x7F))
            }

            function parseRd(instr) -> out {
                out := and64(shr64(instr, toU64(7)), toU64(0x1F))
            }

            function parseFunct3(instr) -> out {
                out := and64(shr64(instr, toU64(12)), toU64(0x7))
            }

            function parseRs1(instr) -> out {
                out := and64(shr64(instr, toU64(15)), toU64(0x1F))
            }

            function parseRs2(instr) -> out {
                out := and64(shr64(instr, toU64(20)), toU64(0x1F))
            }

            function parseFunct7(instr) -> out {
                out := shr64(instr, toU64(25))
            }

            function parseCSSR(instr) -> out {
                out := shr64(instr, toU64(20))
            }

            // unpacked sub-step state
            rdValue := s.RdValue
            pc := s.PC
            instr := s.Instr
            // these fields are ignored if not applicable to the instruction type / opcode
            opcode := parseOpcode(instr)
            rd := parseRd(instr) // destination register index
            funct3 := parseFunct3(instr)
            rs1 := parseRs1(instr) // source register 1 index
            rs2 := parseRs2(instr) // source register 2 index
            funct7 := parseFunct7(instr)
            rs1Value := s.Rs1Value
            rs2Value := s.Rs2Value

            syscallArgsI := s.SyscallArgsI
            syscallRegs := s.SyscallRegs
            subIndex := s.SubIndex

            // write-only sub-state. All reading/processing happens in state machinery above
            let gindex1
            let gindex2
            let size
            let signed
            let value
            let dest
            let offset

            // VM - the instruction transform
            // TODO


            // encode sub-step state
            s.SubIndex = subIndex
            s.PC = pc
            s.Instr = instr
            s.Rs1Value = rs1Value
            s.Rs2Value = rs2Value
            s.RdValue = rdValue
            s.Gindex1 = gindex1
            s.Gindex2 = gindex2
            s.Offset = offset
            s.Value = value
            s.Signed = signed
            s.Size = uint8(size.val())
            s.Dest = dest
            s.SyscallArgsI = syscallArgsI
            s.SyscallRegs = syscallRegs

        }

        return;
    }
}
