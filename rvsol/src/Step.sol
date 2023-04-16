// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;


contract Step {

    address public preimageOracle;

    constructor(address _preimageOracle) public {
        preimageOracle = _preimageOracle;
    }

    // Executes a single RISC-V instruction, starting from the given state-root s,
    // using state-oracle witness data soData, and outputs a new state-root.
    function step(bytes32 s, bytes calldata soData) public returns (bytes32 stateRootOut) {
        assembly {
            function preimageOraclePos() -> out {
                out := 0
            }
            // 0x00 and 0x20 are scratch.
            function stateRootMemAddr() -> out {
                out := 0x40
            }
            function soDataIndexMemAddr() -> out {
                out := 0x60
            }
            mstore(stateRootMemAddr(), calldataload(4))
            mstore(soDataIndexMemAddr(), 100) // 4 + 32 + 32 + 32 = selector, stateroot, offset, length

            function revertWithCode(code) {
                mstore(0, code)
                revert(0, 0x20)
            }

            function hash(a, b) -> h {
                mstore(0, a)
                mstore(0x20, b)
                h := keccak256(0, 0x40)
            }

            function soGet(key) -> a, b {
                let i := mload(soDataIndexMemAddr())
                a := calldataload(i)
                i := add(i, 0x20)
                b := calldataload(i)
                i := add(i, 0x20)
                let h := hash(a, b)
                // TODO: we can check i is after offset and within length, but it buys us nothing
                if iszero(eq(h, key)) {
                    revertWithCode(0x8badf00d)
                }
                mstore(soDataIndexMemAddr(), i)
            }

            function soRemember(a, b) -> h {
                // Use the memory scratchpad for hashing input
                h := hash(a, b)
                // TODO: we can event-log the (a,b) so we can fill the state-oracle with rvsol like with rvgo
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
            function loadResGindex() -> out { out := 14 }
            function preimageGindex() -> out { out := 15 }

            // Writing destinations. Note: also update the switch-case entries (no constant support...)
            function destRead() -> out { out := 0 }
            function destWrite() -> out { out := 1 }
            function destHeapIncr() -> out { out := 2 }
            function destCSRRW() -> out { out := 3 }
            function destCSRRS() -> out { out := 4 }
            function destCSRRC() -> out { out := 5 }
            function destADD() -> out { out   := 6 }
            function destSWAP() -> out { out  := 7 }
            function destXOR() -> out { out   := 8 }
            function destOR() -> out { out    := 9 }
            function destAND() -> out { out   := 10 }
            function destMIN() -> out { out   := 11 }
            function destMAX() -> out { out   := 12 }
            function destMINU() -> out { out  := 13 }
            function destMAXU() -> out { out  := 14 }

            function b32asBEWord(v) -> out {
                out := v
            }
            function beWordAsB32(v) -> out {
                out := v
            }

            // type casts, no-op in yul
            function U64(v) -> out {
                out := v
            }
            function U256(v) -> out {
                out := v
            }

            function toU256(v) -> out {
                out := v
            }

            // 1 11 0110 00000101
            function bitlen(x) -> n {
                if gt(x, sub(shl(128, 1), 1)) {
                    x := shr(128, x)
                    n := add(n, 128)
                }
                if gt(x, sub(shl(64, 1), 1)) {
                    x := shr(64, x)
                    n := add(n, 64)
                }
                if gt(x, sub(shl(32, 1), 1)) {
                    x := shr(32, x)
                    n := add(n, 32)
                }
                if gt(x, sub(shl(16, 1), 1)) {
                    x := shr(16, x)
                    n := add(n, 16)
                }
                if gt(x, sub(shl(8, 1), 1)) {
                    x := shr(8, x)
                    n := add(n, 8)
                }
                if gt(x, sub(shl(4, 1), 1)) {
                    x := shr(4, x)
                    n := add(n, 4)
                }
                if gt(x, sub(shl(2, 1), 1)) {
                    x := shr(2, x)
                    n := add(n, 2)
                }
                if gt(x, sub(shl(1, 1), 1)) {
                    x := shr(1, x)
                    n := add(n, 1)
                }
                if gt(x, 0) {
                    n := add(n, 1)
                }
            }

            function endianSwap(x) -> out {
                for { let i := 0 } lt(i, 32) { i := add(i, 1) } {
                    out := or(shl(8, out), and(x, 0xff))
                    x := shr(8, x)
                }
            }

            //
            // Yul64 - functions to do 64 bit math - see yul64.go
            //
            function u64Mask() -> out { // max uint64
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
                out := shl(toU256(64), toU256(1))
            }

            function u64TopBit() -> out { // 1 << 63
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
                case 0 {
                    out := v
                }
                default {
                    out := or(shl(toU256(64), not(0)), v)
                }
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

            function sdiv64(x, y) -> out { // note: signed overflow semantics are the same between Go and EVM assembly
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


            //
            // Parse - functions to parse RISC-V instructions - see parse.go
            //
            function parseImmTypeI(instr) -> out {
                out := signExtend64(shr64(toU64(20), instr), toU64(11))
            }

            function parseImmTypeS(instr) -> out {
                out := signExtend64(
                    or64(
                        shl64(toU64(5), shr64(toU64(25), instr)),
                        and64(shr64(toU64(7), instr), toU64(0x1F))
                    ),
                    toU64(11))
            }

            function parseImmTypeB(instr) -> out {
                out := signExtend64(
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
                out := signExtend64(
                    or64(
                        or64(
                            and64(shr64(toU64(21), instr), shortToU64(0x3FF)),          // 10 bits for index 0:9
                            shl64(toU64(10), and64(shr64(toU64(20), instr), toU64(1)))  // 1 bit for index 10
                        ),
                        or64(
                            shl64(toU64(11), and64(shr64(toU64(12), instr), toU64(0xFF))), // 8 bits for index 11:18
                            shl64(toU64(19), shr64(toU64(31), instr))                      // 1 bit for index 19
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

            function parseCSSR(instr) -> out {
                out := shr64(toU64(20), instr)
            }

            function read(stateStackGindex, stateGindex, stateStackDepth) -> stateValue, stateStackHash {
                // READING MODE: if the stack gindex is lower than target, then traverse to target
                for {} lt(stateStackGindex, stateGindex) {} {
                    if eq(stateStackGindex, 1) {
                        stateValue := mload(stateRootMemAddr())
                    }
                    stateStackGindex := shl(toU256(1), stateStackGindex)
                    let a, b := soGet(stateValue)
                    switch and(shr(toU256(stateStackDepth), stateGindex), toU256(1))
                    case 1 {
                        stateStackGindex := or(stateStackGindex, toU256(1))
                        stateValue := b
                        // keep track of where we have been, to use the trail to go back up the stack when writing
                        stateStackHash := soRemember(stateStackHash, a)
                    }
                    case 0 {
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
                    let prevStackHash, prevSibling := soGet(stateStackHash)
                    stateStackHash := prevStackHash
                    switch eq(and(stateStackGindex, toU256(1)), toU256(1))
                    case 1 {
                        stateValue := soRemember(prevSibling, stateValue)
                    }
                    case 0 {
                        stateValue := soRemember(stateValue, prevSibling)
                    }
                    stateStackGindex := shr(toU256(1), stateStackGindex)
                    if eq(stateStackGindex, toU256(1)) {
                        mstore(stateRootMemAddr(), stateValue)
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
                // TODO: RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH
                case 3 { // destCSRRW: atomic Read/Write bits in CSR
                    out := endianSwap(stateValue)
                    dest := destWrite()
                }
                case 4 { // destCSRRS: atomic Read and Set bits in CSR
                    out := endianSwap(stateValue)
                    value := or64(out, value) // set bits, v=0 will be no-op
                    dest := destWrite()
                }
                case 5 { // destCSRRC: atomic Read and Clear Bits in CSR
                    out := endianSwap(stateValue)
                    value := and64(out, not64(value)) // clear bits, v=0 will be no-op
                    dest := destWrite()
                }
                case 2 { // destHeapIncr
                    // we want the heap value before we increase it
                    out := endianSwap(stateValue)
                    value := add64(out, value)
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
                        let shamt := shl(toU256(3), sub(sub(toU256(31), toU256(i)), toU256(offset)))
                        let valByte := shl(shamt, and(u64ToU256(value), toU256(0xff)))
                        let maskByte := shl(shamt, toU256(0xff))
                        value := shr64(toU64(8), value)
                        base := or(and(base, not(maskByte)), valByte)
                    }
                    write(targetGindex, rootGindex, beWordAsB32(base), stateStackHash)
                }
                case 0 { // destRead
                    for { let i := 0 } lt(i, firstChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(toU256(3), sub(sub(toU256(31), toU256(i)), toU256(offset)))
                        let valByte := U64(and(shr(shamt, base), toU256(0xff)))
                        out := or64(out, shl64(shl64(toU64(3), toU64(i)), valByte))
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
                        let shamt := shl(toU256(3), toU256(sub(31, i)))
                        let valByte := shl(shamt, and(u64ToU256(value), toU256(0xff)))
                        let maskByte := shl(shamt, toU256(0xff))
                        value := shr64(toU64(8), value)
                        base := or(and(base, not(maskByte)), valByte)
                    }
                    write(targetGindex, rootGindex, beWordAsB32(base), stateStackHash)
                }
                case 0 { // destRead
                    for { let i := 0 } lt(i, secondChunkBytes) { i := add(i, 1) } {
                        let shamt := shl(toU256(3), sub(toU256(31), toU256(i)))
                        let valByte := U64(and(shr(shamt, base), toU256(0xff)))
                        out := or64(out, shl64(shl64(toU64(3), add64(toU64(i), firstChunkBytes)), valByte))
                    }
                }
            }

            function makeMemGindex(byteIndex) -> out {
                // memory is packed in 32 byte leaf values. = 5 bits, thus 64-5=59 bit path
                out := or(shl(toU256(59), memoryGindex()), shr(toU256(5), U256(byteIndex)))
            }

            function makeRegisterGindex(register) -> out {
                if gt(register, 31) { // there are only 32 valid registers
                    revertWithCode(0xbadacce550)
                }
                out := or(shl(toU256(5), registersGindex()), U256(register))
            }

            function makeCSRGindex(num) -> out {
                if gt(num, 4095) { // there are only 4096 valid CSR registers
                    revertWithCode(0xbadacce551)
                }
                out := or(shl(toU256(12), csrGindex()), U256(num))
            }

            function memToStateOp(memIndex, size) -> offset, gindex1, gindex2 {
                gindex1 := makeMemGindex(memIndex)
                offset := and64(memIndex, toU64(31))
                gindex2 := 0
                if iszero(lt(add(toU256(offset), U256(size)), toU256(32))) { // if offset+size >= 32, then it spans into the next memory chunk
                    // note: intentional overflow, circular 64 bit memory is part of riscv5 spec (chapter 1.4)
                    gindex2 := makeMemGindex(add64(memIndex, sub64(size, toU64(1))))
                }
            }

            function loadMem(addr, size, signed) -> out {
                let offset, gindex1, gindex2 := memToStateOp(addr, size)
                out := mutate(gindex1, gindex2, offset, size, destRead(), 0)
                if signed {
                    let topBitIndex := sub64(shl64(toU64(3), size), toU64(1))
                    out := signExtend64(out, topBitIndex)
                }
            }

            function storeMem(addr, size, value) {
                let offset, gindex1, gindex2 := memToStateOp(addr, size)
                pop(mutate(gindex1, gindex2, offset, size, destWrite(), value))
            }

            function loadRegister(num) -> out {
                out := mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destRead(), 0)
            }

            function writeRegister(num, val) {
                if iszero64(num) { // reg 0 must stay 0
                    // v is a HINT, but no hints are specified by standard spec, or used by us.
                    leave
                }
                pop(mutate(makeRegisterGindex(num), toU256(0), 0, toU64(8), destWrite(), val))
            }

            function setLoadReservation(addr) {
                pop(mutate(loadResGindex(), toU256(0), 0, toU64(8), destWrite(), addr))
            }

            function getLoadReservation() -> out {
                out := mutate(loadResGindex(), toU256(0), 0, toU64(8), destRead(), 0)
            }

            function getPC() -> out {
                out := mutate(pcGindex(), toU256(0), 0, toU64(8), destRead(), 0)
            }

            function setPC(v) {
                pop(mutate(pcGindex(), toU256(0), 0, toU64(8), destWrite(), v))
            }

            function opMem(op, addr, size, value) -> out {
                let v := loadMem(addr, size, true)
                out := v
                switch op
                case 6 { // destADD
                    v := add64(v, value)
                } case 7 { // destSWAP
                    v := value
                } case 8 { // destXOR
                    v := xor64(v, value)
                } case 9 { // destOR
                    v := or64(v, value)
                } case 10 { // destAND
                    v := and64(v, value)
                } case 11 { // destMIN
                    if slt64(value, v) {
                        v := value
                    }
                } case 12 { // destMAX
                    if sgt64(value, v) {
                        v := value
                    }
                } case 13 { // destMINU
                    if lt64(value, v) {
                        v := value
                    }
                } case 14 { // destMAXU
                    if gt64(value, v) {
                        v := value
                    }
                } default {
                    revertWithCode(0xbadc0de1) // unrecognized mem op
                }
                storeMem(addr, size, v)
            }

            function updateCSR(num, v, mode) -> out {
                let dest := 0
                switch mode
                case 1 {
                    dest := destCSRRW() // ?01 = CSRRW(I)
                } case 2 {
                    dest := destCSRRS() // ?10 = CSRRS(I)
                } case 3 {
                    dest := destCSRRC() // ?11 = CSRRC(I)
                } default {
                    revertWithCode(0xbadc0de0)
                }
                out := mutate(makeCSRGindex(num), toU256(0), 0, toU64(8), dest, v)
            }

            function writePreimageKey(addr, count) -> out {
                // adjust count down, so we only have to read a single 32 byte leaf of memory
                let alignment := and64(addr, toU64(31))
                let maxData := sub64(toU64(32), alignment)
                if gt64(count, maxData) {
                    count := maxData
                }

                let memGindex := makeMemGindex(addr)
                let node, stateStackHash := read(toU256(1), memGindex, 61) // top tree + mem tree - root bit - inspect bit = 4 + (64-5) - 1 - 1 = 61
                // mask the part of the data we are shifting in
                let bits := shl(toU256(3), u64ToU256(count))
                let mask := sub(shl(bits, toU256(1)), toU256(1))
                let dat := and(b32asBEWord(node), mask)

                node, stateStackHash := read(toU256(1), preimageGindex(), 2)
                let preImageKey, __ := soGet(node)

                // Append to key content by bit-shifting
                let key := b32asBEWord(preImageKey)
                key := shl(bits, key)
                key := or(key, dat)

                // We reset the pre-image value offset back to 0 (the right part of the merkle pair)
                let newPreImageRoot := soRemember(beWordAsB32(key), 0)
                write(preimageGindex(), toU256(1), newPreImageRoot, stateStackHash)
                out := count
            }

            function readPreimagePart(key, offset) -> dat, datlen {
                let addr := sload(preimageOraclePos()) // calling Oracle.readPreimage(bytes32,uint256)
                mstore(0x80, shl(224, 0xe03110e1)) // (32-4)*8=224: right-pad the function selector, and then store it as prefix
                mstore(0x84, key)
                mstore(0xa4, offset)
                let cgas := 100000 // TODO change call gas
                let res := call(cgas, addr, 0, 0x80, 0x44, 0x00, 0x40) // output into scratch space
                if res { // 1 on success
                    dat := mload(0x00)
                    datlen := mload(0x20)
                    leave
                }
                revertWithCode(0xbadf00d)
            }

            function readPreimageValue(addr, count) -> out {
                let node, stateStackHash := read(toU256(1), preimageGindex(), 2)
                let preImageKey, preImageValueOffset := soGet(node)

                let offset := u256ToU64(b32asBEWord(preImageValueOffset))

                // make call to pre-image oracle contract
                let pdatB32, pdatlen := readPreimagePart(preImageKey, offset)
                if iszero64(toU64(pdatlen)) { // EOF
                    out := toU64(0)
                    leave
                }
                mstore(1009, pdatB32)
                mstore(1010, pdatlen)

                // align with memory
                let alignment := and64(addr, toU64(31))    // how many bytes addr is offset from being left-aligned
                let maxData := sub64(toU64(32), alignment) // higher alignment leaves less room for data this step
                if gt64(count, maxData) {
                    count := maxData
                }
                if gt64(count, toU64(pdatlen)) { // cannot read more than pdatlen
                    count := toU64(pdatlen)
                }

                let bits := shl64(toU64(3), sub64(toU64(32), count))             // 32-count, in bits
                let mask := not(sub(shl(u64ToU256(bits), toU256(1)), toU256(1))) // left-aligned mask for count bytes
                let alignmentBits := u64ToU256(shl64(toU64(3), alignment))
                mask := shr(alignmentBits, mask)                  // mask of count bytes, shifted by alignment
                let pdat := shr(alignmentBits, b32asBEWord(pdatB32)) // pdat, shifted by alignment

                // update pre-image reader with updated offset
                let newOffset := add64(offset, count)
                let newPreImageRoot := soRemember(preImageKey, beWordAsB32(u64ToU256(newOffset)))
                write(preimageGindex(), toU256(1), newPreImageRoot, stateStackHash)

                // put data into memory
                let memGindex := makeMemGindex(addr)
                node, stateStackHash := read(toU256(1), memGindex, 61)
                let dat := and(b32asBEWord(node), not(mask)) // keep old bytes outside of mask
                dat := or(dat, and(pdat, mask))           // fill with bytes from pdat

                write(memGindex, toU256(1), beWordAsB32(dat), stateStackHash)
                out := count
            }

            function sysCall() {
                let a7 := loadRegister(toU64(17))
                switch a7
                case 93 { // exit the calling thread. No multi-thread support yet, so just exit.
                    let a0 := loadRegister(toU64(10))
                    pop(mutate(exitGindex(), toU256(0), 0, toU64(8), destWrite(), a0))
                    // program stops here, no need to change registers.
                } case 94 { // exit-group
                    let a0 := loadRegister(toU64(10))
                    pop(mutate(exitGindex(), toU256(0), 0, toU64(8), destWrite(), a0))
                } case 214 { // brk
                    // Go sys_linux_riscv64 runtime will only ever call brk(NULL), i.e. first argument (register a0) set to 0.

                    // brk(0) changes nothing about the memory, and returns the current page break
                    let v := shl64(toU64(30), toU64(1)) // set program break at 1 GiB
                    writeRegister(toU64(10), v)
                    writeRegister(toU64(11), toU64(0)) // no error
                } case 222 { // mmap
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
                        // No hint, allocate it ourselves, by as much as the requested length.
                        // Increase the length to align it with desired page size if necessary.
                        let align := and64(length, shortToU64(4095))
                        if align {
                            length := add64(length, sub64(shortToU64(4096), align))
                        }
                        let heap := mutate(heapGindex(), toU256(0), 0, toU64(8), destHeapIncr(), length) // increment heap with length
                        writeRegister(toU64(10), heap)
                    }
                    default {
                        // allow hinted memory address (leave it in A0 as return argument)
                    }
                    writeRegister(toU64(11), toU64(0)) // no error
                } case 63 { // read
                    let fd := loadRegister(toU64(10))    // A0 = fd
                    let addr := loadRegister(toU64(11))  // A1 = *buf addr
                    let count := loadRegister(toU64(12)) // A2 = count
                    let n := 0
                    let errCode := 0
                    switch fd
                    case 0 { // stdin
                        n := toU64(0) // never read anything from stdin
                        errCode := toU64(0)
                    } case 3 { // pre-image oracle
                        n := readPreimageValue(addr, count)
                        errCode := toU64(0)
                    } default {
                        n := u64Mask()         //  -1 (reading error)
                        errCode := toU64(0x4d) // EBADF
                    }
                    writeRegister(toU64(10), n)
                    writeRegister(toU64(11), errCode)
                } case 64 { // write
                    let fd := loadRegister(toU64(10))    // A0 = fd
                    let addr := loadRegister(toU64(11))  // A1 = *buf addr
                    let count := loadRegister(toU64(12)) // A2 = count
                    let n := 0
                    let errCode := 0
                    switch fd
                    case 1 { // stdout
                        //_, err := io.Copy(stdOut, s.GetMemRange(addr, count)) // TODO stdout
                        //if err != nil {
                        //	panic(fmt.Errorf("stdout writing err: %w", err))
                        //}
                        n := count // write completes fully in single instruction step
                        errCode := toU64(0)
                    } case 2 { // stderr
                        //_, err := io.Copy(stdErr, s.GetMemRange(addr, count)) // TODO stderr
                        //if err != nil {
                        //	panic(fmt.Errorf("stderr writing err: %w", err))
                        //}
                        n := count // write completes fully in single instruction step
                        errCode := toU64(0)
                    } case 3 { // pre-image oracle
                        n := writePreimageKey(addr, count)
                        errCode := toU64(0) // no error
                    } default { // any other file, including (4) pre-image hinter
                        n := u64Mask()         //  -1 (writing error)
                        errCode := toU64(0x4d) // EBADF
                    }
                    writeRegister(toU64(10), n)
                    writeRegister(toU64(11), errCode)
                } case 25 { // fcntl - file descriptor manipulation / info lookup
                    let fd := loadRegister(toU64(10))  // A0 = fd
                    let cmd := loadRegister(toU64(11)) // A1 = cmd
                    let out := 0
                    let errCode := 0
                    switch cmd
                    case 0x3 { // F_GETFL: get file descriptor flags
                        switch fd
                        case 0 { // stdin
                            out := toU64(0) // O_RDONLY
                        } case 1 { // stdout
                            out := toU64(1) // O_WRONLY
                        } case 2 { // stderr
                            out := toU64(1) // O_WRONLY
                        } case 3 { // pre-image oracle
                            out := toU64(2) // O_RDWR
                        } default {
                            out := u64Mask()
                            errCode := toU64(0x4d) // EBADF
                        }
                    } default { // no other commands: don't allow changing flags, duplicating FDs, etc.
                        out := u64Mask()
                        errCode := toU64(0x16) // EINVAL (cmd not recognized by this kernel)
                    }
                    writeRegister(toU64(10), out)
                    writeRegister(toU64(11), errCode) // EBADF
                } case 56 { // openat - the Go linux runtime will try to open optional /sys/kernel files for performance hints
                    writeRegister(toU64(10), u64Mask())
                    writeRegister(toU64(11), toU64(0xd)) // EACCES - no access allowed
                } case 123 { // sched_getaffinity - hardcode to indicate affinity with any cpu-set mask
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                } case 113 { // clock_gettime
                    let addr := loadRegister(toU64(11))                  // addr of timespec struct
                    storeMem(addr, toU64(8), shortToU64(1337))           // seconds
                    storeMem(add64(addr, toU64(8)), toU64(8), toU64(42)) // nanoseconds: must be nonzero to pass Go runtimeInitTime check
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                } case 135 { // rt_sigprocmask - ignore any sigset changes
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                } case 132 { // sigaltstack - ignore any hints of an alternative signal receiving stack addr
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                } case 178 { // gettid - hardcode to 0
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                } case 134 { // rt_sigaction - no-op, we never send signals, and thus need no sig handler info
                    writeRegister(toU64(10), toU64(0))
                    writeRegister(toU64(11), toU64(0))
                //case 220 // clone - not supported
                } case 163 { // getrlimit
                    let res := loadRegister(toU64(10))
                    let addr := loadRegister(toU64(11))
                    switch res
                    case 0x7 {  // RLIMIT_NOFILE
                        storeMem(addr, toU64(8), shortToU64(1024))                  // soft limit. 1024 file handles max open
                        storeMem(add64(addr, toU64(8)), toU64(8), shortToU64(1024)) // hard limit
                    } default {
                        revertWithCode(0xf0012) // unrecognized resource limit lookup
                    }
                } default {
                    revertWithCode(0xf001ca11) // unrecognized system call
                }
            }

            let _pc := getPC()
            let instr := loadMem(_pc, toU64(4), false)

            // these fields are ignored if not applicable to the instruction type / opcode
            let opcode := parseOpcode(instr)
            let rd := parseRd(instr) // destination register index
            let funct3 := parseFunct3(instr)
            let rs1 := parseRs1(instr) // source register 1 index
            let rs2 := parseRs2(instr) // source register 2 index
            let funct7 := parseFunct7(instr)

            switch opcode
            case 0x03 { // 000_0011: memory loading
                // LB, LH, LW, LD, LBU, LHU, LWU
                let imm := parseImmTypeI(instr)
                let signed := iszero64(and64(funct3, toU64(4)))      // 4 = 100 -> bitflag
                let size := shl64(and64(funct3, toU64(3)), toU64(1)) // 3 = 11 -> 1, 2, 4, 8 bytes size
                let rs1Value := loadRegister(rs1)
                let memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                let rdValue := loadMem(memIndex, size, signed)
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x23 { // 010_0011: memory storing
                // SB, SH, SW, SD
                let imm := parseImmTypeS(instr)
                let size := shl64(funct3, toU64(1))
                let value := loadRegister(rs2)
                let rs1Value := loadRegister(rs1)
                let memIndex := add64(rs1Value, signExtend64(imm, toU64(11)))
                storeMem(memIndex, size, value)
                setPC(add64(_pc, toU64(4)))
            } case 0x63 { // 110_0011: branching
                let rs1Value := loadRegister(rs1)
                let rs2Value := loadRegister(rs2)
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
                    _pc := add64(_pc, toU64(4))
                } default {
                    let imm := parseImmTypeB(instr)
                    // imm12 is a signed offset, in multiples of 2 bytes.
			        // So it's really 13 bits with a hardcoded 0 bit.
                    _pc := add64(_pc, imm)
                }
                // not like the other opcodes: nothing to write to rd register, and PC has already changed
                setPC(_pc)
            } case 0x13 { // 001_0011: immediate arithmetic and logic
		        let rs1Value := loadRegister(rs1)
                let imm := parseImmTypeI(instr)
                let rdValue := 0
                switch funct3
                case 0 { // 000 = ADDI
                    rdValue := add64(rs1Value, imm)
                } case 1 { // 001 = SLLI
                    rdValue := shl64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
                } case 2 { // 010 = SLTI
                    rdValue := slt64(rs1Value, imm)
                } case 3 { // 011 = SLTIU
                    rdValue := lt64(rs1Value, imm)
                } case 4 { // 100 = XORI
                    rdValue := xor64(rs1Value, imm)
                } case 5 { // 101 = SR~
                    switch shr64(toU64(6), imm) // in rv64i the top 6 bits select the shift type
                    case 0x00 { // 000000 = SRLI
                        rdValue := shr64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
                    } case 0x10 { // 010000 = SRAI
                        rdValue := sar64(and64(imm, toU64(0x3F)), rs1Value) // lower 6 bits in 64 bit mode
                    }
                } case 6 { // 110 = ORI
                    rdValue := or64(rs1Value, imm)
                } case 7 { // 111 = ANDI
                    rdValue := and64(rs1Value, imm)
                }
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x1B { // 001_1011: immediate arithmetic and logic signed 32 bit
		        let rs1Value := loadRegister(rs1)
                let imm := parseImmTypeI(instr)
                let rdValue := 0
                switch funct3
                case 0 { // 000 = ADDIW
                    rdValue := mask32Signed64(add64(rs1Value, imm))
                } case 1 { // 001 = SLLIW
                    rdValue := mask32Signed64(shl64(and64(imm, toU64(0x1F)), rs1Value))
                } case 5 { // 101 = SR~
                    let shamt := and64(imm, toU64(0x1F))
                    switch shr64(toU64(6), imm) // in rv64i the top 6 bits select the shift type
                    case 0x00 { // 000000 = SRLIW
                        rdValue := signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
                    } case 0x10 { // 010000 = SRAIW
                        rdValue := signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x33 { // 011_0011: register arithmetic and logic
		        let rs1Value := loadRegister(rs1)
		        let rs2Value := loadRegister(rs2)
                let rdValue := 0
                switch funct7
                case 1 { // RV M extension
                    switch funct3
                    case 0 { // 000 = MUL: signed x signed
                        rdValue := mul64(rs1Value, rs2Value)
                    } case 1 { // 001 = MULH: upper bits of signed x signed
                        rdValue := u256ToU64(shr(toU256(64), mul(signExtend64To256(rs1Value), signExtend64To256(rs2Value))))
                    } case 2 { // 010 = MULHSU: upper bits of signed x unsigned
                        rdValue := u256ToU64(shr(toU256(64), mul(signExtend64To256(rs1Value), u64ToU256(rs2Value))))
                    } case 3 { // 011 = MULHU: upper bits of unsigned x unsigned
                        rdValue := u256ToU64(shr(toU256(64), mul(u64ToU256(rs1Value), u64ToU256(rs2Value))))
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
                        rdValue := shl64(and64(rs2Value, toU64(0x3F)), rs1Value) // only the low 6 bits are consider in RV6VI
                    } case 2 { // 010 = SLT
                        rdValue := slt64(rs1Value, rs2Value)
                    } case 3 { // 011 = SLTU
                        rdValue := lt64(rs1Value, rs2Value)
                    } case 4 { // 100 = XOR
                        rdValue := xor64(rs1Value, rs2Value)
                    } case 5 { // 101 = SR~
                        switch funct7
                        case 0x00 { // 0000000 = SRL
                            rdValue := shr64(and64(rs2Value, toU64(0x3F)), rs1Value) // logical: fill with zeroes
                        } case 0x20 { // 0100000 = SRA
                            rdValue := sar64(and64(rs2Value, toU64(0x3F)), rs1Value) // arithmetic: sign bit is extended
                        }
                    } case 6 { // 110 = OR
                        rdValue := or64(rs1Value, rs2Value)
                    } case 7 { // 111 = AND
                        rdValue := and64(rs1Value, rs2Value)
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x3B { // 011_1011: register arithmetic and logic in 32 bits
                let rs1Value := loadRegister(rs1)
                let rs2Value := loadRegister(rs2)
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
                } default {
                    switch funct3
                    case 0 { // 000 = ADDW/SUBW
                        switch funct7
                        case 0x00 { // 0000000 = ADDW
                            rdValue := mask32Signed64(add64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        } case 0x20 { // 0100000 = SUBW
                            rdValue := mask32Signed64(sub64(and64(rs1Value, u32Mask()), and64(rs2Value, u32Mask())))
                        }
                    } case 1 { // 001 = SLLW
                        rdValue := mask32Signed64(shl64(and64(rs2Value, toU64(0x1F)), rs1Value))
                    } case 5 { // 101 = SR~
                        let shamt := and64(rs2Value, toU64(0x1F))
                        switch funct7
                        case 0x00 { // 0000000 = SRLW
                            rdValue := signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), toU64(31))
                        } case 0x20 { // 0100000 = SRAW
                            rdValue := signExtend64(shr64(shamt, and64(rs1Value, u32Mask())), sub64(toU64(31), shamt))
                        }
                    }
                }
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x37 { // 011_0111: LUI = Load upper immediate
                let imm := parseImmTypeU(instr)
                let rdValue := shl64(toU64(12), imm)
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x17 { // 001_0111: AUIPC = Add upper immediate to PC
                let imm := parseImmTypeU(instr)
                let rdValue := add64(_pc, signExtend64(shl64(toU64(12), imm), toU64(31)))
                writeRegister(rd, rdValue)
                setPC(add64(_pc, toU64(4)))
            } case 0x6F { // 110_1111: JAL = Jump and link
                let imm := parseImmTypeJ(instr)
                let rdValue := add64(_pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(add64(_pc, signExtend64(shl64(toU64(1), imm), toU64(20)))) // signed offset in multiples of 2 bytes (last bit is there, but ignored)
            } case 0x67 { // 110_0111: JALR = Jump and link register
		        let rs1Value := loadRegister(rs1)
                let imm := parseImmTypeI(instr)
                let rdValue := add64(_pc, toU64(4))
                writeRegister(rd, rdValue)
                setPC(and64(add64(rs1Value, signExtend64(imm, toU64(11))), xor64(u64Mask(), toU64(1)))) // least significant bit is set to 0
            } case 0x73 { // 111_0011: environment things
                switch funct3
                case 0 { // 000 = ECALL/EBREAK
                    switch shr64(toU64(20), instr) // I-type, top 12 bits
                    case 0 { // imm12 = 000000000000 ECALL
                        sysCall()
                        setPC(add64(_pc, toU64(4)))
                    } default { // imm12 = 000000000001 EBREAK
                        setPC(add64(_pc, toU64(4))) // ignore breakpoint
                    }
                } default { // CSR instructions
                    let imm := parseCSSR(instr)
                    let value := rs1
                    if iszero64(and64(funct3, toU64(4))) {
                        value := loadRegister(rs1)
                    }
                    let mode := and64(funct3, toU64(3))
                    let rdValue := updateCSR(imm, value, mode)
                    writeRegister(rd, rdValue)
                    setPC(add64(_pc, toU64(4)))
                }
            } case 0x2F { // 010_1111: RV32A and RV32A atomic operations extension
                // acquire and release bits:
                //   aq := and64(shr64(toU64(1), funct7), toU64(1))
                //   rl := and64(funct7, toU64(1))
                // if none set: unordered
                // if aq is set: no following mem ops observed before acquire mem op
                // if rl is set: release mem op not observed before earlier mem ops
                // if both set: sequentially consistent
                // These are no-op here because there is no pipeline of mem ops to acquire/release.

                // 0b010 == RV32A W variants
                // 0b011 == RV64A D variants
                let size := shl64(funct3, toU64(1))
                if lt64(size, toU64(4)) {
                    revertWithCode(0xbada70) // bad AMO size
                }
                let addr := loadRegister(rs1)

                let op := shr64(toU64(2), funct7)
                switch op
                case 0x2 { // 00010 = LR = Load Reserved
                    let v := loadMem(addr, size, true)
                    writeRegister(rd, v)
                    setLoadReservation(addr)
                } case 0x3 { // 00011 = SC = Store Conditional
                    let rdValue := toU64(1)
                    if eq64(addr, getLoadReservation()) {
                        let rs2Value := loadRegister(rs2)
                        storeMem(addr, size, rs2Value)
                        rdValue := toU64(0)
                    }
                    writeRegister(rd, rdValue)
                    setLoadReservation(toU64(0))
                } default { // AMO: Atomic Memory Operation
                    let rs2Value := loadRegister(rs2)
                    if eq64(size, toU64(4)) {
                        rs2Value := mask32Signed64(rs2Value)
                    }
                    // Specifying the operation allows us to implement it closer to the memory for smaller witness data.
                    // And that too can be optimized: only one 32 bytes leaf is affected,
                    // since AMOs are always 4 or 8 byte aligned (Zam extension not supported here).
                    let dest := 0
                    switch op
                    case 0x0 { // 00000 = AMOADD = add
                        dest := destADD()
                    } case 0x1 { // 00001 = AMOSWAP
                        dest := destSWAP()
                    } case 0x4 { // 00100 = AMOXOR = xor
                        dest := destXOR()
                    } case 0x8 { // 01000 = AMOOR = or
                        dest := destOR()
                    } case 0xc { // 01100 = AMOAND = and
                        dest := destAND()
                    } case 0x10 { // 10000 = AMOMIN = min signed
                        dest := destMIN()
                    } case 0x14 { // 10100 = AMOMAX = max signed
                        dest := destMAX()
                    } case 0x18 { // 11000 = AMOMINU = min unsigned
                        dest := destMINU()
                    } case 0x1c { // 11100 = AMOMAXU = max unsigned
                        dest := destMAXU()
                    } default {
                        revertWithCode(0xf001a70) // unknown atomic operation
                    }
                    let rdValue := opMem(dest, addr, size, rs2Value)
                    writeRegister(rd, rdValue)
                }
                setPC(add64(_pc, toU64(4)))
            } case 0x0F { // 000_1111: fence
                // Used to impose additional ordering constraints; flushing the mem operation pipeline.
                // This VM doesn't have a pipeline, nor additional harts, so this is a no-op.
                // FENCE / FENCE.TSO / FENCE.I all no-op: there's nothing to synchronize.
                setPC(add64(_pc, toU64(4)))
            } case 0x07 { // FLW/FLD: floating point load word/double
		        setPC(add64(_pc, toU64(4))) // no-op this.
	        } case 0x27 { // FSW/FSD: floating point store word/double
		        setPC(add64(_pc, toU64(4))) // no-op this.
	        } case 0x53 { // FADD etc. no-op is enough to pass Go runtime check
		        setPC(add64(_pc, toU64(4))) // no-op this.
            } default {
                revertWithCode(0xf001c0de)
            }

            return(stateRootMemAddr(), 0x20)
        }
    }
}
