// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;


contract Step {



    //
    function step(bytes32 calldata s, bytes calldata soData) public pure view returns (bytes32 stateRoot) {
        stateRoot = s;

        assembly {
            function soGet(key) -> a, b {
                // TODO increment memory counter, and load relevant length/data
                if not(eq(d.length, 64)) { revert(0, 0) }  // TODO fix length lookup
                if not(eq(keccak256(d), key)) { revert(0, 0) }  // TODO fix hashing
                a, b := unpack(d)  // TODO unpack data
            }

            function soRemember(a, b) -> h {
                // TODO: we can event-log the (a,b) so we can fill the state-oracle with rvsol like with rvgo
                h := keccak256(a, b)
            }

            // tree:
            // ```
            //
            //	         1
            //	    2          3
            //	 4    5     6     7
            //	8 9 10 11 12 13 14 15
            //
            // ```
            function pcGindex() -> out { out := 8 }
            function memoryGindex() -> out { out := 9 }
            function registersGindex() -> out { out := 10 }
            function csrGindex() -> out { out := 11 }
            function exitGindex() -> out { out := 12 }
            function heapGindex() -> out { out := 13 }

            // Writing destinations. Note: also update the switch-case entries (no constant support...)
            function destRead() -> out { out := 0 }
            function destWrite() -> out { out := 1 }
            function destHeapIncr() -> out { out := 2 }
            function destCSRRW() -> out { out := 3 }
            function destCSRRS() -> out { out := 4 }
            function destCSRRC() -> out { out := 5 }

            // type casts, no-op in yul
            function U64(v) -> out {
                out := v
            }
            function U256(v) -> out {
                out := v
            }

            function bitlen(x) -> n {
                if gt(x, sub(shl(1, 128), 1)) {
                    x := shr(x, 128)
                    n := add(n, 128)
                }
                if gt(x, sub(shl(1, 64), 1)) {
                    x := shr(x, 64)
                    n := add(n, 64)
                }
                if gt(x, sub(shl(1, 32), 1)) {
                    x := shr(x, 32)
                    n := add(n, 32)
                }
                if gt(x, sub(shl(1, 16), 1)) {
                    x := shr(x, 16)
                    n := add(n, 16)
                }
                if gt(x, sub(shl(1, 8), 1)) {
                    x := shr(x, 8)
                    n := add(n, 8)
                }
                if gt(x, sub(shl(1, 4), 1)) {
                    x := shr(x, 4)
                    n := add(n, 4)
                }
                if gt(x, sub(shl(1, 2), 1)) {
                    x := shr(x, 2)
                    n := add(n, 2)
                }
                if gt(x, sub(shl(1, 1), 1)) {
                    x := shr(x, 1)
                    n := add(n, 1)
                }
                if gt(x, sub(shl(1, 0), 1)) {
                    n := add(n, 0)
                }
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
                        stateValue := 123  // TODO mload stateRoot
                        // stateValue = stateRoot  // TODO
                    }
                    stateStackGindex := shl(stateStackGindex, toU256(1))
                    let a, b := soGet(stateValue)
                    switch and(shr(stateGindex, toU256(stateStackDepth)), toU256(1))
                    case 1 {
                        stateStackGindex := or(stateStackGindex, toU256(1))
                        stateValue := b
                        // keep track of where we have been, to use the trail to go back up the stack when writing
                        stateStackHash := soRemember(stateStackHash, a)
                    }
                    case 2 {
                        stateValue := a
                        // keep track of where we have been, to use the trail to go back up the stack when writing
                        stateStackHash := soRemember(stateStackHash, b)
                    }
                    stateStackDepth := sub(stateStackDepth, 1)
                }
            }

            function write(stateStackGindex, stateGindex, stateValue, stateStackHash) {
                // WRITING MODE: if the stack gindex is higher than the target, then traverse back to root and update along the way
                for {} gt(stateStackGindex, stateGindex) {} {
                    prevStackHash, prevSibling := soGet(stateStackHash)
                    stateStackHash := prevStackHash
                    switch eq(and(stateStackGindex, toU256(1)), toU256(1))
                    case 1 {
                        stateValue := soRemember(prevSibling, stateValue)
                    }
                    case 0 {
                        stateValue := soRemember(stateValue, prevSibling)
                    }
                    stateStackGindex := shr(stateStackGindex, toU256(1))
                    if eq(stateStackGindex, toU256(1)) {
                        //if d, ok := so.(oracle.Differ); ok {
                        //	fmt.Println("state change")
                        //	d.Diff(stateRoot, stateValue, 1)
                        //}
                        stateRoot := stateValue
                    }
                }
            }

            function mutate(gindex1, gindex2, offset, size, dest, value) -> out {
                // if we have not reached the gindex yet, then we need to start traversal to it
                let rootGindex := toU256(1)
                let stateStackDepth := sub(bitlen(gindex1), 2)
                let targetGindex := gindex1

                let stateValue, stateStackHash := read(rootGindex, targetGindex, stateStackDepth)

                switch dest
                case 3 { // destCSRRW
                    // special case: CSRRW - read and write bits
                    out := stateValue
                    dest := destWrite()
                }
                case 4 { // destCSRRS
                    // special case: CSRRS - read and set bits
                    out := stateValue
                    value := or64(out, value) // set bits
                    dest := destWrite()
                }
                case 5 { // destCSRRC
                    // special case: CSRRC - read and clear bits
                    out := stateValue
                    value := and64(out, not64(value)) // clear bits
                    dest := destWrite()
                }
                case 2 { // destHeapIncr
                    // special case: increment before writing, and output result
                    value := add64(value, stateValue)
                    out := value
                    dest := destWrite()
                }

                let firstChunkBytes := sub64(toU64(32), toU64(offset))
                if gt64(firstChunkBytes, size) {
                    firstChunkBytes := size
                }

                let base := b32asBEWord(stateValue)
                // we reached the value, now load/write it
                switch dest
                case 1 { // destWrite
                    for { let i := 0 } lt(i, firstChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(sub(sub(toU256(31), toU256(i)), toU256(offset)), toU256(3))
                        let valByte := shl(and(u64ToU256(value), toU256(0xff)), shamt)
                        let maskByte := shl(toU256(0xff), shamt)
                        value := shr64(value, toU64(8))
                        base := or(and(base, not(maskByte)), valByte)
                    }
                    write(targetGindex, rootGindex, beWordAsB32(base), stateStackHash)
                }
                case 0 { // destRead
                    for { let i := 0 } lt(i, firstChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(sub(sub(toU256(31), toU256(i)), toU256(offset)), toU256(3))
                        let valByte := U64(and(shr(base, shamt), toU256(0xff)))
                        out := or64(out, shl64(valByte, shl64(toU64(i), toU64(3))))
                    }
                }

                if iszero(gindex2) {
                    leave
                }

                stateStackDepth := sub(bitlen(gindex2), 2)
                targetGindex := gindex2

                stateValue, stateStackHash := read(rootGindex, targetGindex, stateStackDepth)

                let secondChunkBytes := sub64(size, firstChunkBytes)

                base := b32asBEWord(stateValue)
                // we reached the value, now load/write it
                switch dest
                case 1 { // destWrite
                    // note: StateValue holds the old 32 bytes, some of which may stay the same
                    for { let i := 0 } lt(i, secondChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(toU256(sub(31, i)), toU256(3))
                        let valByte := shl(and(u64ToU256(value), toU256(0xff)), shamt)
                        let maskByte := shl(toU256(0xff), shamt)
                        value := shr64(value, toU64(8))
                        base := or(and(base, not(maskByte)), valByte)
                    }
                    write(targetGindex, rootGindex, beWordAsB32(base), stateStackHash)
                }
                case 0 { // destRead
                    for { let i := 0 } lt(i, secondChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(sub(toU256(31), toU256(i)), toU256(3))
                        let valByte := U64(and(shr(base, shamt), toU256(0xff)))
                        out := or64(out, shl64(valByte, shl64(add64(toU64(i), firstChunkBytes), toU64(3))))
                    }
                }
            }

            function loadMem(addr, size, signed) -> out {
                let offset, gindex1, gindex2 := memToStateOp(addr, size)
                out := mutate(gindex1, gindex2, offset, size, destRead(), 0)
                if signed {
                    let topBitIndex := sub64(shl64(size, toU64(3)), toU64(1))
                    out := signExtend64(out, topBitIndex)
                }
            }

            function storeMem(addr, size, value) {
                offset, gindex1, gindex2 := memToStateOp(addr, size)
                mutate(gindex1, gindex2, offset, size, destWrite(), value)
            }

            function loadRegister(num) -> out {
                out := mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destRead(), 0)
            }

            function writeRegister(num, val) {
                if iszero64(num) { // reg 0 must stay 0
                    // v is a HINT, but no hints are specified by standard spec, or used by us.
                    leave
                }
                mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destWrite(), val)
            }

            function getPC() -> out {
                out := mutate(pcGindex, toU256(0), 0, toU64(8), destRead(), 0)
            }

            function setPC(pc) {
                mutate(pcGindex, toU256(0), 0, toU64(8), destWrite(), pc)
            }

            function readCSR(num) -> out {
                out := mutate(makeCSRGindex(num), toU256(0), 0, toU64(8), destRead(), 0)
            }

            function writeCSR(num, v) {
                mutate(makeCSRGindex(num), toU256(0), 0, toU64(8), destWrite(), v)
            }

            function sysCall() {
                let a7 := loadRegister(toU64(17))
                switch a7
                case 93 { // exit
                    let a0 := loadRegister(toU64(0))
                    mutate(exitGindex, toU256(0), 0, toU64(8), destWrite(), a0)
                }
                case 214 { // brk
                    // Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

                    // brk(0) changes nothing about the memory, and returns the current page break
                    let v := shl64(toU64(1), toU64(30)) // set program break at 1 GiB
                    writeRegister(toU64(10), v)
                }
                case 222 { // mmap
                    // A0 = addr (hint)
                    let addr := loadRegister(toU64(10))
                    // A1 = n (length)
                    let length := loadRegister(toU64(11))
                    // A2 = prot (memory protection type, can ignore)
                    // A3 = flags (shared with other process and or written back to file, can ignore)  // TODO maybe assert the MAP_ANONYMOUS flag is set
                    // A4 = fd (file descriptor, can ignore because we support anon memory only)
                    // A5 = offset (offset in file, we don't support any non-anon memory, so we can ignore this)

                    // ignore: prot, flags, fd, offset
                    switch addr
                    case 0 {
                        // no hint, allocate it ourselves, by as much as the requested length
                        heap := mutate(heapGindex, toU256(0), 0, toU64(8), destHeapIncr(), length)
                        writeRegister(toU64(10), heap)
                    }
                    default {
                        // allow hinted memory address (leave it in A0 as return argument)
                    }
                    writeRegister(toU64(11), toU64(0)) // no error
                }
                default {
                    // TODO maybe revert if the syscall is unrecognized?
                }
            }

            let pc := getPC()
            let instr := loadMem(pc, toU64(4), false)

            // these fields are ignored if not applicable to the instruction type / opcode
            let opcode := parseOpcode(instr)
            let rd := parseRd(instr) // destination register index
            let funct3 := parseFunct3(instr)
            let rs1 := parseRs1(instr) // source register 1 index
            let rs2 := parseRs2(instr) // source register 2 index
            let funct7 := parseFunct7(instr)
            let rs1Value := loadRegister(rs1)
            let rs2Value := loadRegister(rs2)

            //fmt.Printf("slow PC: %x\n", pc)
            //fmt.Printf("slow INSTR: %x\n", instr)
            //fmt.Printf("slow OPCODE: %x\n", opcode)
            //fmt.Printf("slow rs1 value: %x\n", rs1Value)
            //fmt.Printf("slow rs2 value: %x\n", rs2Value)

            switch opcode
            case 0x03 { // 000_0011: memory loading
                // LB, LH, LW, LD, LBU, LHU, LWU
                let imm := parseImmTypeI(instr)
                let signed := iszero64(and64(funct3, toU64(4)))      // 4 = 100 -> bitflag
                let size := shl64(toU64(1), and64(funct3, toU64(3))) // 3 = 11 -> 1, 2, 4, 8 bytes size
                let memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                let rdValue := loadMem(memIndex, size, signed)
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x23 { // 010_0011: memory storing
                // SB, SH, SW, SD
                let imm := parseImmTypeS(instr)
                let size := shl64(toU64(1), funct3)
                let value := rs2Value
                let memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                storeMem(memIndex, size, value)
                setPC(add64(pc, toU64(4)))
            } case 0x63 { // 110_0011: branching
                let branchHit := toU64(0)
                switch funct3
                case 0 { // 000 = BEQ
                    branchHit := eq64(rs1Value, rs2Value)
                } case 1 { // 001 = BNE
                    branchHit := and64(not64(eq64(rs1Value, rs2Value)), toU64(1))
                } case 4 { // 100 = BLT
                    branchHit := slt64(rs1Value, rs2Value)
                } case 5 { // 101 = BGE
                    branchHit := and64(not64(slt64(rs1Value, rs2Value)), toU64(1))
                } case 6 { // 110 = BLTU
                    branchHit := lt64(rs1Value, rs2Value)
                } case 7 { // 111 := BGEU
                    branchHit := and64(not64(lt64(rs1Value, rs2Value)), toU64(1))
                }
                switch branchHit
                case 0 {
                    pc := add64(pc, toU64(4))
                } default {
                    imm := parseImmTypeB(instr)
                    // imm12 is a signed offset, in multiples of 2 bytes
                    pc := add64(pc, signExtend64(imm, toU64(11)))
                }
                // not like the other opcodes: nothing to write to rd register, and PC has already changed
                setPC(pc)
            } case 0x13 { // 001_0011: immediate arithmetic and logic
                let imm := parseImmTypeI(instr)
                let rdValue := 0
                switch funct3
                case 0 { // 000 = ADDI
                    rdValue := add64(rs1Value, imm)
                } case 1 { // 001 = SLLI
                    rdValue := shl64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                } case 2 { // 010 = SLTI
                    rdValue := slt64(rs1Value, imm)
                } case 3 { // 011 = SLTIU
                    rdValue := lt64(rs1Value, imm)
                } case 4 { // 100 = XORI
                    rdValue := xor64(rs1Value, imm)
                } case 5 { // 101 = SR~
                    switch funct7
                    case 0x00 { // 0000000 = SRLI
                        rdValue := shr64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                    } case 0x20 { // 0100000 = SRAI
                        rdValue := sar64(rs1Value, and64(imm, toU64(0x3F))) // lower 6 bits in 64 bit mode
                    }
                } case 6 { // 110 = ORI
                    rdValue := or64(rs1Value, imm)
                } case 7 { // 111 = ANDI
                    rdValue := and64(rs1Value, imm)
                }
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x1B { // 001_1011: immediate arithmetic and logic signed 32 bit
                let imm := parseImmTypeI(instr)
                let rdValue := 0
                switch funct3
                case 0 { // 000 = ADDIW
                    rdValue := mask32Signed64(add64(rs1Value, imm))
                } case 1 { // 001 = SLLIW
                    rdValue := mask32Signed64(shl64(rs1Value, and64(imm, toU64(0x1F))))
                } case 5 { // 101 = SR~
                    let shamt := and64(imm, toU64(0x1F))
                    switch funct7
                    case 0x00 { // 0000000 = SRLIW
                        rdValue := signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), toU64(31))
                    } case 0x20 { // 0100000 = SRAIW
                        rdValue := signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x33 { // 011_0011: register arithmetic and logic
                let rdValue := 0
                switch funct7
                case 1 { // RV32M extension
                    switch funct3
                    case 0 { // 000 = MUL: signed x signed
                        rdValue := mul64(rs1Value, rs2Value)
                    } case 1 { // 001 = MULH: upper bits of signed x signed
                        rdValue := u256ToU64(shr(mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value)), toU256(64)))
                    } case 2 { // 010 = MULHSU: upper bits of signed x unsigned
                        rdValue := u256ToU64(shr(mul(signExtend64To256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
                    } case 3 { // 011 = MULHU: upper bits of unsigned x unsigned
                        rdValue := u256ToU64(shr(mul(u64ToU256(rs1Value), u64ToU256(rs2Value)), toU256(64)))
                    } case 4 { // 100 = DIV
                        switch rs2Value
                        case 0 {
                            rdValue := u64Mask()
                        } default {
                            rdValue := sdiv64(rs1Value, rs2Value)
                        }
                    } case 5 { // 101 = DIVU
                        switch rs2Value
                        case 0 {
                            rdValue := u64Mask()
                        } default {
                            rdValue := div64(rs1Value, rs2Value)
                        }
                    } case 6 { // 110 = REM
                        switch rs2Value
                        case 0 {
                            rdValue := rs1Value
                        } default {
                            rdValue := smod64(rs1Value, rs2Value)
                        }
                    } case 7 { // 111 = REMU
                        switch rs2Value
                        case 0 {
                            rdValue := rs1Value
                        } default {
                            rdValue := mod64(rs1Value, rs2Value)
                        }
                    }
                } default {
                    switch funct3
                    case 0 { // 000 = ADD/SUB
                        switch funct7
                        case 0x00 { // 0000000 = ADD
                            rdValue := add64(rs1Value, rs2Value)
                        } case 0x20 { // 0100000 = SUB
                            rdValue := sub64(rs1Value, rs2Value)
                        }
                    } case 1 { // 001 = SLL
                        rdValue := shl64(rs1Value, and64(rs2Value, toU64(0x3F))) // only the low 6 bits are consider in RV6VI
                    } case 2 { // 010 = SLT
                        rdValue := slt64(rs1Value, rs2Value)
                    } case 3 { // 011 = SLTU
                        rdValue := lt64(rs1Value, rs2Value)
                    } case 4 { // 100 = XOR
                        rdValue := xor64(rs1Value, rs2Value)
                    } case 5 { // 101 = SR~
                        switch funct7
                        case 0x00 { // 0000000 = SRL
                            rdValue := shr64(rs1Value, and64(rs2Value, toU64(0x3F))) // logical: fill with zeroes
                        } case 0x20 { // 0100000 = SRA
                            rdValue := sar64(rs1Value, and64(rs2Value, toU64(0x3F))) // arithmetic: sign bit is extended
                        }
                    } case 6 { // 110 = OR
                        rdValue := or64(rs1Value, rs2Value)
                    } case 7 { // 111 = AND
                        rdValue := and64(rs1Value, rs2Value)
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x3B { // 011_1011: register arithmetic and logic in 32 bits
                let rdValue := 0
                switch funct7
                case 1 { // RV64M extension
                    switch funct3
                    case 0 { // 000 = MULW
                        rdValue := mask32Signed64(mul64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                    } case 4 { // 100 = DIVW
                        switch rs2Value
                        case 0 {
                            rdValue := u64Mask()
                        } default {
                            rdValue := mask32Signed64(sdiv64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
                        }
                    } case 5 { // 101 = DIVUW
                        switch rs2Value
                        case 0 {
                            rdValue := u64Mask()
                        } default {
                            rdValue := mask32Signed64(div64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    } case 6 { // 110 = REMW
                        switch rs2Value
                        case 0 {
                            rdValue := mask32Signed64(rs1Value)
                        } default {
                            rdValue := mask32Signed64(smod64(mask32Signed64(rs1Value), mask32Signed64(rs2Value)))
                        }
                    } case 7 { // 111 = REMUW
                        switch rs2Value
                        case 0 {
                            rdValue := mask32Signed64(rs1Value)
                        } default {
                            rdValue := mask32Signed64(mod64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    }
                } default { // RV32M extension
                    switch funct3
                    case 0 { // 000 = ADDW/SUBW
                        switch funct7
                        case 0x00 { // 0000000 = ADDW
                            rdValue := mask32Signed64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        } case 0x20 { // 0100000 = SUBW
                            rdValue := mask32Signed64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    } case 1 { // 001 = SLLW
                        rdValue := mask32Signed64(shl64(rs1Value, and64(rs2Value, toU64(0x1F))))
                    } case 5 { // 101 = SR~
                        let shamt := and64(rs2Value, toU64(0x1F))
                        switch funct7
                        case 0x00 { // 0000000 = SRLW
                            rdValue := signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), toU64(31))
                        } case 0x20 { // 0100000 = SRAW
                            rdValue := signExtend64(shr64(and64(rs1Value, u32Mask()), shamt), sub64(toU64(31), shamt))
                        }
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x37 { // 011_0111: LUI = Load upper immediate
                let imm := parseImmTypeU(instr)
                let rdValue := shl64(imm, toU64(12))
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x17 { // 001_0111: AUIPC = Add upper immediate to PC
                let imm := parseImmTypeU(instr)
                let rdValue := add64(pc, signExtend64(shl64(imm, toU64(12)), toU64(31)))
                writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x6F { // 110_1111: JAL = Jump and link
                let imm := parseImmTypeJ(instr)
                let rdValue := add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(add64(pc, signExtend64(imm, toU64(21)))) // signed offset in multiples of 2 bytes
            } case 0x67 { // 110_0111: JALR = Jump and link register
                let imm := parseImmTypeI(instr)
                let rdValue := add64(pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(and64(add64(rs1Value, signExtend64(imm, toU64(12))), xor64(u64Mask(), toU64(1)))) // least significant bit is set to 0
            } case 0x73 { // 111_0011: environment things
                switch funct3
                case 0 { // 000 = ECALL/EBREAK
                    switch shr64(instr, toU64(20)) // I-type, top 12 bits
                    case 0 { // imm12 = 000000000000 ECALL
                        sysCall()
                        setPC(add64(pc, toU64(4)))
                    } default { // imm12 = 000000000001 EBREAK
                        // ignore breakpoint
                        setPC(add64(pc, toU64(4)))
                    }
                } default { // CSR instructions
                    let imm := parseCSSR(instr)
                    let rdValue := readCSR(imm)
                    let value := rs1
                    if iszero64(and64(funct3, toU64(4))) {
                        value := rs1Value
                    }
                    switch and64(funct3, toU64(3))
                    case 1 { // ?01 = CSRRW(I) = "atomic Read/Write bits in CSR"
                        writeCSR(imm, value)
                    } case 2 { // ?10 = CSRRS = "atomic Read and Set bits in CSR"
                        writeCSR(imm, or64(rdValue, value)) // v=0 will be no-op
                    } case 3 { // ?11 = CSRRC = "atomic Read and Clear Bits in CSR"
                        writeCSR(imm, and64(rdValue, not64(value))) // v=0 will be no-op
                    }
                    // TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
                    writeRegister(rd, rdValue)
                    setPC(add64(pc, toU64(4)))
                }
            } case 0x2F { // 010_1111: RV32A and RV32A atomic operations extension
                // TODO atomic operations
                // 0b010 == RV32A W variants
                // 0b011 == RV64A D variants
                //size := 1 << funct3
                switch shr64(and64(funct7, toU64(0x1F)), toU64(2))
                case 0x0 { // 00000 = AMOADD
                } case 0x1 { // 00001 = AMOSWAP
                } case 0x2 { // 00010 = LR
                } case 0x3 { // 00011 = SC
                } case 0x4 { // 00100 = AMOXOR
                } case 0x8 { // 01000 = AMOOR
                } case 0xc { // 01100 = AMOAND
                } case 0x10 { // 10000 = AMOMIN
                } case 0x14 { // 10100 = AMOMAX
                } case 0x18 { // 11000 = AMOMINU
                } case 0x1c { // 11100 = AMOMAXU
                }
                //writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } case 0x0F { // 000_1111: fence
                //// TODO: different layout of func data
                //// "fm pred succ"
                //switch funct3
                //case 0 {  // 000
                //	switch funct7
                //	case 0x41 { // 100_0001 = FENCE.TSO
                //	} default { // FENCE
                //	}
                //} case 1 { // 001: FENCE.I
                //}
                // We can no-op FENCE, there's nothing to synchronize
                //writeRegister(rd, rdValue)
                setPC(add64(pc, toU64(4)))
            } default {
                revert(0, 0) // TODO memory output: unknown opcode: %b full instruction: %b", opcode, instr
            }
        }

        return;
    }
}
