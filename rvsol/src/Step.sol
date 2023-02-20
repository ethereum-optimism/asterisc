// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;


contract Step {


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
    uint256 immutable destRead     = 0;
    uint256 immutable destWrite       = 1;
    uint256 immutable destHeapIncr    = 2;
    uint256 immutable destCSRRW = 3;
    uint256 immutable destCSRRS = 4;
    uint256 immutable destCSRRC  = 5;

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
    function step(bytes32 calldata s, bytes calldata soData) public pure view returns (bytes32 stateRoot) {
        stateRoot = s;

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

            function u64Mod() -> out { // 1 << 64
                out := shl(toU256(1), toU256(64))
            }

            function u64TopBit() -> out { // 1 << 63
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

            function signExtend64To256(v) -> out {
                switch and(U256(v), u64TopBit())
                case 0 {
                    out := v
                }
                default {
                    out := or(shl(not(0), toU256(64)), v)
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
                            shl64(and64(shr64(instr, toU64(25)), toU64(0x3F)), toU64(5))
                        ),
                        or64(
                            shl64(and64(shr64(instr, toU64(7)), toU64(1)), toU64(11)),
                            shl64(shr64(instr, toU64(31)), toU64(12))
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
                            shl64(and64(shr64(instr, toU64(20)), toU64(1)), toU64(10))
                        ),
                        or64(
                            shl64(and64(shr64(instr, toU64(12)), toU64(0xFF)), toU64(11)),
                            shl64(shr64(instr, toU64(31)), toU64(19))
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

            function read(stateStackGindex, stateGindex, stateStackDepth) -> stateValue, stateStackHash {
                // READING MODE: if the stack gindex is lower than target, then traverse to target
                for {} lt(stateStackGindex, stateGindex) {} {
                    if eq(stateStackGindex, 1) {
                        stateValue = 123  // TODO mload stateRoot
                        // stateValue = stateRoot  // TODO
                    }
                    stateStackGindex = shl(stateStackGindex, toU256(1))
                    a, b := so.Get(stateValue)
                    if not(iszero(and(shr(stateGindex, toU256(stateStackDepth)), toU256(1)))) {
                        stateStackGindex = or(stateStackGindex, toU256(1))
                        stateValue = b
                        // keep track of where we have been, to use the trail to go back up the stack when writing
                        stateStackHash = so.Remember(stateStackHash, a)
                    } else {
                        stateValue = a
                        // keep track of where we have been, to use the trail to go back up the stack when writing
                        stateStackHash = so.Remember(stateStackHash, b)
                    }
                    stateStackDepth -= 1
                }
            }

            function write(stateStackGindex, stateGindex, stateValue, stateStackHash) {
                // WRITING MODE: if the stack gindex is higher than the target, then traverse back to root and update along the way
                for {} gt(stateStackGindex, stateGindex) {} {
                    prevStackHash, prevSibling := so.Get(stateStackHash)
                    stateStackHash = prevStackHash
                    if not(iszero(eq(and(stateStackGindex, toU256(1)), toU256(1)))) {
                        stateValue = so.Remember(prevSibling, stateValue)
                    } else {
                        stateValue = so.Remember(stateValue, prevSibling)
                    }
                    stateStackGindex = shr(stateStackGindex, toU256(1))
                    if eq(stateStackGindex, toU256(1)) {
                        //if d, ok := so.(oracle.Differ); ok {
                        //	fmt.Println("state change")
                        //	d.Diff(stateRoot, stateValue, 1)
                        //}
                        stateRoot = stateValue
                    }
                }
            }

            function mutate(gindex1, gindex2, offset, size, dest, value) -> out {
                // if we have not reached the gindex yet, then we need to start traversal to it
                rootGindex := toU256(1)
                stateStackDepth := uint8(gindex1.BitLen()) - 2
                targetGindex := gindex1

                stateValue, stateStackHash := read(rootGindex, targetGindex, stateStackDepth)

                switch dest {
                case destCSRRW:
                    // special case: CSRRW - read and write bits
                    out = decodeU64(stateValue[:8])
                    dest = destWrite
                case destCSRRS:
                    // special case: CSRRS - read and set bits
                    out = decodeU64(stateValue[:8])
                    value = or64(out, value) // set bits
                    dest = destWrite
                case destCSRRC:
                    // special case: CSRRC - read and clear bits
                    out = decodeU64(stateValue[:8])
                    value = and64(out, not64(value)) // clear bits
                    dest = destWrite
                case destHeapIncr:
                    // special case: increment before writing, and output result
                    value = add64(value, decodeU64(stateValue[:8]))
                    out = value
                    dest = destWrite
                }

                firstChunkBytes := sub64(toU64(32), toU64(offset))
                if gt64(firstChunkBytes, size) != (U64{}) {
                    firstChunkBytes = size
                }

                // we reached the value, now load/write it
                switch dest {
                case destWrite:
                    // note: stateValue holds the old 32 bytes, some of which may stay the same
                    v := encodePacked(value)
                    copy(stateValue[offset:], v[:size.val()])
                    write(targetGindex, rootGindex, stateValue, stateStackHash)
                case destRead:
                    out = decodeU64(stateValue[offset : uint64(offset)+firstChunkBytes.val()])
                }

                if gindex2 == (U256{}) {
                    return
                }

                stateStackDepth = uint8(gindex2.BitLen()) - 2
                targetGindex = gindex2

                stateValue, stateStackHash = read(rootGindex, targetGindex, stateStackDepth)

                secondChunkBytes := sub64(size, firstChunkBytes)

                // we reached the value, now load/write it
                switch dest {
                case destWrite:
                    // note: StateValue holds the old 32 bytes, some of which may stay the same
                    v := encodePacked(value)
                    copy(stateValue[:secondChunkBytes.val()], v[firstChunkBytes.val():size.val()])
                    write(targetGindex, rootGindex, stateValue, stateStackHash)
                case destRead:
                    a := decodeU64(stateValue[0:secondChunkBytes.val()])
                    out = or64(shl64(a, shl64(firstChunkBytes, toU64(3))), out)
                }
                return
            }

            loadMem := func(addr U64, size U64, signed bool) (out U64) {
                offset, gindex1, gindex2 := memToStateOp(addr, size)
                out = mutate(gindex1, gindex2, offset, size, destRead, U64{})
                if signed {
                    topBitIndex := sub64(shl64(size, toU64(3)), toU64(1))
                    out = signExtend64(out, topBitIndex)
                }
                return
            }

            storeMem := func(addr U64, size U64, value U64) {
                offset, gindex1, gindex2 := memToStateOp(addr, size)
                mutate(gindex1, gindex2, offset, size, destWrite, value)
            }

            loadRegister := func(num U64) (out U64) {
                out = mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destRead, U64{})
                return
            }

            writeRegister := func(num U64, val U64) {
                if iszero64(num) { // reg 0 must stay 0
                    // v is a HINT, but no hints are specified by standard spec, or used by us.
                    return
                }
                mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destWrite, val)
            }

            getPC := func() U64 {
                return mutate(pcGindex, toU256(0), 0, toU64(8), destRead, U64{})
            }

            setPC := func(pc U64) {
                mutate(pcGindex, toU256(0), 0, toU64(8), destWrite, pc)
            }

            readCSR := func(num U64) (out U64) {
                out = mutate(makeCSRGindex(num), toU256(0), 0, toU64(8), destRead, U64{})
                return
            }

            writeCSR := func(num U64, v U64) {
                mutate(makeCSRGindex(num), toU256(0), 0, toU64(8), destWrite, v)
            }

            sysCall := func() {
                a7 := loadRegister(toU64(17))
                switch a7.val() {
                case 93: // exit
                    a0 := loadRegister(toU64(0))
                    mutate(exitGindex, toU256(0), 0, toU64(8), destWrite, a0)
                case 214: // brk
                    // Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

                    // brk(0) changes nothing about the memory, and returns the current page break
                    v := shl64(toU64(1), toU64(30)) // set program break at 1 GiB
                    writeRegister(toU64(10), v)
                case 222: // mmap
                    // A0 = addr (hint)
                    addr := loadRegister(toU64(10))
                    // A1 = n (length)
                    length := loadRegister(toU64(11))
                    // A2 = prot (memory protection type, can ignore)
                    // A3 = flags (shared with other process and or written back to file, can ignore)  // TODO maybe assert the MAP_ANONYMOUS flag is set
                    // A4 = fd (file descriptor, can ignore because we support anon memory only)
                    // A5 = offset (offset in file, we don't support any non-anon memory, so we can ignore this)

                    // ignore: prot, flags, fd, offset
                    switch addr.val() {
                    case 0:
                        // no hint, allocate it ourselves, by as much as the requested length
                        heap := mutate(heapGindex, toU256(0), 0, toU64(8), destHeapIncr, length)
                        writeRegister(toU64(10), heap)
                    default:
                        // allow hinted memory address (leave it in A0 as return argument)
                    }
                    writeRegister(toU64(11), toU64(0)) // no error
                default:
                    // TODO maybe revert if the syscall is unrecognized?
                }
            }

            pc := getPC()
            instr := loadMem(pc, toU64(4), false)

            // these fields are ignored if not applicable to the instruction type / opcode
            opcode := parseOpcode(instr)
            rd := parseRd(instr) // destination register index
            funct3 := parseFunct3(instr)
            rs1 := parseRs1(instr) // source register 1 index
            rs2 := parseRs2(instr) // source register 2 index
            funct7 := parseFunct7(instr)
            rs1Value := loadRegister(rs1)
            rs2Value := loadRegister(rs2)

            //fmt.Printf("slow PC: %x\n", pc)
            //fmt.Printf("slow INSTR: %x\n", instr)
            //fmt.Printf("slow OPCODE: %x\n", opcode)
            //fmt.Printf("slow rs1 value: %x\n", rs1Value)
            //fmt.Printf("slow rs2 value: %x\n", rs2Value)

            switch opcode.val() {
            case 0b0000011: // memory loading
                // LB, LH, LW, LD, LBU, LHU, LWU
                imm := parseImmTypeI(instr)
                signed := iszero64(and64(funct3, toU64(4)))      // 4 = 100 -> bitflag
                size := shl64(toU64(1), and64(funct3, toU64(3))) // 3 = 11 -> 1, 2, 4, 8 bytes size
                memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                rdValue := loadMem(memIndex, size, signed)
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0100011: // memory storing
                // SB, SH, SW, SD
                imm := parseImmTypeS(instr)
                size := shl64(toU64(1), funct3)
                value := rs2Value
                memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                storeMem(memIndex, size, value)
                pc = add64(pc, toU64(4))
                setPC(pc)
            case 0b1100011: // branching
                branchHit := toU64(0)
                switch funct3.val() {
                case 0: // 000 = BEQ
                    branchHit = eq64(rs1Value, rs2Value)
                case 1: // 001 = BNE
                    branchHit = and64(not64(eq64(rs1Value, rs2Value)), toU64(1))
                case 4: // 100 = BLT
                    branchHit = slt64(rs1Value, rs2Value)
                case 5: // 101 = BGE
                    branchHit = and64(not64(slt64(rs1Value, rs2Value)), toU64(1))
                case 6: // 110 = BLTU
                    branchHit = lt64(rs1Value, rs2Value)
                case 7: // 111 = BGEU
                    branchHit = and64(not64(lt64(rs1Value, rs2Value)), toU64(1))
                }
                switch branchHit.val() {
                case 0:
                    pc = add64(pc, toU64(4))
                default:
                    imm := parseImmTypeB(instr)
                    // imm12 is a signed offset, in multiples of 2 bytes
                    pc = add64(pc, signExtend64(imm, toU64(11)))
                }
                // not like the other opcodes: nothing to write to rd register, and PC has already changed
                setPC(pc)
            case 0b0010011: // immediate arithmetic and logic
                imm := parseImmTypeI(instr)
                var rdValue U64
                switch funct3.val() {
                case 0: // 000 = ADDI
                    rdValue = add64(rs1Value, imm)
                case 1: // 001 = SLLI
                    rdValue = shl64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                case 2: // 010 = SLTI
                    rdValue = slt64(rs1Value, imm)
                case 3: // 011 = SLTIU
                    rdValue = lt64(rs1Value, imm)
                case 4: // 100 = XORI
                    rdValue = xor64(rs1Value, imm)
                case 5: // 101 = SR~
                    switch funct7.val() {
                    case 0x00: // 0000000 = SRLI
                        rdValue = shr64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                    case 0x20: // 0100000 = SRAI
                        rdValue = sar64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                    }
                case 6: // 110 = ORI
                    rdValue = or64(rs1Value, imm)
                case 7: // 111 = ANDI
                    rdValue = and64(rs1Value, imm)
                }
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0011011: // immediate arithmetic and logic signed 32 bit
                imm := parseImmTypeI(instr)
                var rdValue U64
                switch funct3.val() {
                case 0: // 000 = ADDIW
                    rdValue = mask32Signed64(add64(rs1Value, imm))
                case 1: // 001 = SLLIW
                    rdValue = mask32Signed64(shl64(rs1Value, and64(imm, toU64(0x1F))))
                case 5: // 101 = SR~
                    shamt := and64(imm, toU64(0x1F))
                    switch funct7.val() {
                    case 0x00: // 0000000 = SRLIW
                        rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), toU64(31))
                    case 0x20: // 0100000 = SRAIW
                        rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
                    }
                }
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0110011: // register arithmetic and logic
                var rdValue U64
                switch funct7.val() {
                case 1: // RV32M extension
                    switch funct3.val() {
                    case 0: // 000 = MUL: signed x signed
                        rdValue = mul64(rs1Value, rs2Value)
                    case 1: // 001 = MULH: upper bits of signed x signed
                        rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value)), toU256(64)))
                    case 2: // 010 = MULHSU: upper bits of signed x unsigned
                        rdValue = u256ToU64(shr(mul(signExtend64To256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
                    case 3: // 011 = MULHU: upper bits of unsigned x unsigned
                        rdValue = u256ToU64(shr(mul(u64ToU256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
                    case 4: // 100 = DIV
                        switch rs2Value.val() {
                        case 0:
                            rdValue = u64Mask()
                        default:
                            rdValue = sdiv64(rs1Value, rs2Value)
                        }
                    case 5: // 101 = DIVU
                        switch rs2Value.val() {
                        case 0:
                            rdValue = u64Mask()
                        default:
                            rdValue = div64(rs1Value, rs2Value)
                        }
                    case 6: // 110 = REM
                        switch rs2Value.val() {
                        case 0:
                            rdValue = rs1Value
                        default:
                            rdValue = smod64(rs1Value, rs2Value)
                        }
                    case 7: // 111 = REMU
                        switch rs2Value.val() {
                        case 0:
                            rdValue = rs1Value
                        default:
                            rdValue = mod64(rs1Value, rs2Value)
                        }
                    }
                default:
                    switch funct3.val() {
                    case 0: // 000 = ADD/SUB
                        switch funct7.val() {
                        case 0x00: // 0000000 = ADD
                            rdValue = add64(rs1Value, rs2Value)
                        case 0x20: // 0100000 = SUB
                            rdValue = sub64(rs1Value, rs2Value)
                        }
                    case 1: // 001 = SLL
                        rdValue = shl64(rs1Value, and64(rs2Value, toU64(0x3F))) // only the low 6 bits are consider in RV6VI
                    case 2: // 010 = SLT
                        rdValue = slt64(rs1Value, rs2Value)
                    case 3: // 011 = SLTU
                        rdValue = lt64(rs1Value, rs2Value)
                    case 4: // 100 = XOR
                        rdValue = xor64(rs1Value, rs2Value)
                    case 5: // 101 = SR~
                        switch funct7.val() {
                        case 0x00: // 0000000 = SRL
                            rdValue = shr64(rs1Value, and64(rs2Value, toU64(0x3F))) // logical: fill with zeroes
                        case 0x20: // 0100000 = SRA
                            rdValue = sar64(rs1Value, and64(rs2Value, toU64(0x3F))) // arithmetic: sign bit is extended
                        }
                    case 6: // 110 = OR
                        rdValue = or64(rs1Value, rs2Value)
                    case 7: // 111 = AND
                        rdValue = and64(rs1Value, rs2Value)
                    }
                }
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0111011: // register arithmetic and logic in 32 bits
                var rdValue U64
                switch funct7.val() {
                case 1: // RV64M extension
                    switch funct3.val() {
                    case 0: // 000 = MULW
                        rdValue = mask32Signed64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                    case 4: // 100 = DIVW
                        switch rs2Value.val() {
                        case 0:
                            rdValue = u64Mask()
                        default:
                            rdValue = mask32Signed64(sdiv64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
                        }
                    case 5: // 101 = DIVUW
                        switch rs2Value.val() {
                        case 0:
                            rdValue = u64Mask()
                        default:
                            rdValue = mask32Signed64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    case 6: // 110 = REMW
                        switch rs2Value.val() {
                        case 0:
                            rdValue = mask32Signed64(rs1Value)
                        default:
                            rdValue = mask32Signed64(smod64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
                        }
                    case 7: // 111 = REMUW
                        switch rs2Value.val() {
                        case 0:
                            rdValue = mask32Signed64(rs1Value)
                        default:
                            rdValue = mask32Signed64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    }
                default: // RV32M extension
                    switch funct3.val() {
                    case 0: // 000 = ADDW/SUBW
                        switch funct7.val() {
                        case 0x00: // 0000000 = ADDW
                            rdValue = mask32Signed64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        case 0x20: // 0100000 = SUBW
                            rdValue = mask32Signed64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    case 1: // 001 = SLLW
                        rdValue = mask32Signed64(shl64(rs1Value, and64(rs2Value, toU64(0x1F))))
                    case 5: // 101 = SR~
                        shamt := and64(rs2Value, toU64(0x1F))
                        switch funct7.val() {
                        case 0x00: // 0000000 = SRLW
                            rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), toU64(31))
                        case 0x20: // 0100000 = SRAW
                            rdValue = signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
                        }
                    }
                }
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0110111: // LUI = Load upper immediate
                imm := parseImmTypeU(instr)
                rdValue := shl64(imm, toU64(12))
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0010111: // AUIPC = Add upper immediate to PC
                imm := parseImmTypeU(instr)
                rdValue := add64(pc, signExtend64(shl64(imm, toU64(12)), toU64(31)))
                pc = add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b1101111: // JAL = Jump and link
                imm := parseImmTypeJ(instr)
                rdValue := add64(pc, toU64(4))
                pc = add64(pc, signExtend64(imm, toU64(21))) // signed offset in multiples of 2 bytes
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b1100111: // JALR = Jump and link register
                imm := parseImmTypeI(instr)
                rdValue := add64(pc, toU64(4))
                pc = and64(add64(rs1Value, signExtend64(imm, toU64(12))), xor64(u64Mask(), toU64(1))) // least significant bit is set to 0
                writeRegister(rd, rdValue)
                setPC(pc)
            case 0b1110011:
                switch funct3.val() {
                case 0: // 000 = ECALL/EBREAK
                    switch shr64(instr, toU64(20)).val() { // I-type, top 12 bits
                    case 0: // imm12 = 000000000000 ECALL
                        sysCall()
                        pc = add64(pc, toU64(4))
                        setPC(pc)
                    default: // imm12 = 000000000001 EBREAK
                        // ignore breakpoint
                        pc = add64(pc, toU64(4))
                        setPC(pc)
                    }
                default: // CSR instructions
                    imm := parseCSSR(instr)
                    rdValue := readCSR(imm)
                    value := rs1
                    if iszero64(and64(funct3, toU64(4))) {
                        value = rs1Value
                    }
                    switch and64(funct3, toU64(3)).val() {
                    case 1: // ?01 = CSRRW(I) = "atomic Read/Write bits in CSR"
                        writeCSR(imm, value)
                    case 2: // ?10 = CSRRS = "atomic Read and Set bits in CSR"
                        writeCSR(imm, or64(rdValue, value)) // v=0 will be no-op
                    case 3: // ?11 = CSRRC = "atomic Read and Clear Bits in CSR"
                        writeCSR(imm, and64(rdValue, not64(value))) // v=0 will be no-op
                    }
                    // TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
                    pc = add64(pc, toU64(4))
                    writeRegister(rd, rdValue)
                    setPC(pc)
                }
            case 0b0101111: // RV32A and RV32A atomic operations extension
                // TODO atomic operations
                // 0b010 == RV32A W variants
                // 0b011 == RV64A D variants
                //size := 1 << funct3
                switch shr64(and64(funct7, toU64(0x1F)), toU64(2)).val() {
                case 0x0: // 00000 = AMOADD
                case 0x1: // 00001 = AMOSWAP
                case 0x2: // 00010 = LR
                case 0x3: // 00011 = SC
                case 0x4: // 00100 = AMOXOR
                case 0x8: // 01000 = AMOOR
                case 0xc: // 01100 = AMOAND
                case 0x10: // 10000 = AMOMIN
                case 0x14: // 10100 = AMOMAX
                case 0x18: // 11000 = AMOMINU
                case 0x1c: // 11100 = AMOMAXU
                }
                pc = add64(pc, toU64(4))
                //writeRegister(rd, rdValue)
                setPC(pc)
            case 0b0001111:
                //// TODO: different layout of func data
                //// "fm pred succ"
                //switch funct3 {
                //case 0b000:
                //	switch funct7 {
                //	case 0b1000001: // FENCE.TSO
                //	default: // FENCE
                //	}
                //case 0b001: // FENCE.I
                //}
                // We can no-op FENCE, there's nothing to synchronize
                pc = add64(pc, toU64(4))
                //writeRegister(rd, rdValue)
                setPC(pc)
            default:
                panic(fmt.Errorf("unknown opcode: %b full instruction: %b", opcode, instr))
            }


        }

        return;
    }
}
