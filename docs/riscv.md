# RISC-V resources

## Helpful learning resources

- rv32 instruction set cheat sheet: http://blog.translusion.com/images/posts/RISC-V-cheatsheet-RV32I-4-3.pdf
- rv32: reference card: https://github.com/jameslzhu/riscv-card/blob/master/riscv-card.pdf
- online riscv32 interpreter: https://www.cs.cornell.edu/courses/cs3410/2019sp/riscv/interpreter/#
- specs: https://riscv.org/technical/specifications/
- Berkely riscv card: https://inst.eecs.berkeley.edu/~cs61c/fa18/img/riscvcard.pdf
- "riscv bytes" blog posts: https://danielmangum.com/categories/risc-v-bytes/

## Emulators

The official fully-featured emulator "Spike": https://github.com/riscv-software-src/riscv-isa-sim

Go emulators to diff-fuzz against maybe:

- Doom emulator (RV32I): https://github.com/racerxdl/riscv-emulator   (Apache 2.0 licensed)
- grvemu (RV32I + Zicsr): https://github.com/kinpoko/grvemu  (MIT licensed)
- rv64 (RV64IMAFDC): https://github.com/mohanson/rv64        (WTFPL licensed)
- riscv-emu (RV64IMAC): https://github.com/LMMilewski/riscv-emu                    (Apache 2.0 licensed)

## Summary

- 32 uint64 registers (widened from 32 to 64 bits for 64 bit mode):
  - 0: zero
  - 1: ra - return address
  - 2: sp - stack pointer
  - 3: gp - global pointer
  - 4: tp - thread pointer
  - 5: t0 - temporary register
  - 6: t1
  - 7: t2
  - 8: s0/fp - saved by callee
  - 9: s1
  - 10-17: a0 - a7 - function arguments (and a0,a1 return values)
  - 18-27: s2 - s11
  - 28-31: t3 - t6

- "CSR": Control and Status Register
  - 0x001: fflags - floating point accrued exceptions  (read/write)
  - 0x002: frm - floating point dynamic rounding mode  (read/write)
  - 0x003: fcsr - floating point control and status register (frm+fflags)  (read/write)
  - 0xC00: cycle
  - 0xC01: time
  - 0xC80: instret
  - 0xC81: cycleh
  - 0xC82: timeh
- instructions:
  - "abbreviation G for the IMAFDZicsr Zifencei combination of instruction-set extensions."
  - "C" is the "compressed instruction set for performance / code size / energy efficiency"
    - Unprivileged:
      Atomics A
      Single-Precision Floating-Point F
      Double-Precision Floating-Point D
      General G
      Quad-Precision Floating-Point Q
      Decimal Floating-Point L
      16-bit Compressed Instructions C
      Bit Manipulation B
      Dynamic Languages J
      Transactional Memory T
      Packed-SIMD Extensions P
      Vector Extensions V
      User-Level Interrupts N
      Control and Status Register Access
      Instruction-Fetch Fence Zifencei
      Misaligned Atomics Zam
      Total Store Ordering Ztso

  - RV32I Base Instruction Set:
    - addi, slti, sltiu, xori, ori, andi, slli, srli, srai # immediate
    - add, sub, sll, slt, sltu, xor, srl, sra, or, and # on register
    - beq, bne, blt, bge, bltu, bgeu # branching
    - auipc # add upper imm to pc
    - lui # load upper imm
    - jal # jump and link
    - jalr # jump and link register
    - lb, lh, lw, lbu, lhu # load mem ops
    - sw, sh, sb # store mem ops
    - ecall / ebreak # syscall related
    - fence # ???
  - RV64I Base Instruction Set:
    - lwu, ld
    - sd
    - slli, srli, srai
    - addiw
    - slliw, srliw, sraiw
    - addw, subw
    - sllw, srlw, sraw
  - RV32/RV64 Zifencei: FENCE.I
  - RV32/RV64 Zicsr: CSRR{W,S,C,WI,SI,CI} # to interact with control and status registers
  - RVM:  # Integer Multiplication and Division
    - RV32M:
      - MUL,MULH,MULHSU,MULHU,
      - DIV,DIVU,
      - REM,REMU
    - RV64M:
      - MULW
      - DIVW,DIVUW,
      - REMW,REMUW
  - RVA:  # Atomics
    - RV32A:
      - LR.W
      - SC.W
      - AMOSWAP.W
      - AMOADD.W
      - AMOXOR.W
      - AMOOR.W
      - AMOMIN.W
      - AMOMAX.W
      - AMOMINU.W
      - AMOMAXU.W
    - RV64A:
      - LR.D
      - SC.D
      - AMOSWAP.D
      - AMOADD.D
      - AMOXOR.D
      - AMOOR.D
      - AMOMIN.D
      - AMOMAX.D
      - AMOMINU.D
      - AMOMAXU.D
  - RVF:  # Single-Precision Floating-Point
    - RV32F
    - RV64F
  - RVD:  # Double-Precision Floating-Point
    - RV32D
    - RV64D
  - RVQ:  # Quad-Precision Floating-Point
    - RV32Q
    - RV64Q

- instruction types: (different instruction formats, all spanning 32 bits)
  - R: register-register ALU instructions, no immediate
  - I: immediate ALU instructions, and load instructions
  - S: store instructions
  - B: branching
  - U: upper-immediate
  - J: jumps

## RiscV Supervisor Binary Interface

https://github.com/riscv-non-isa/riscv-sbi-doc/blob/master/riscv-sbi.adoc

The SBI is like a RiscV specific version of an EEI (Execution Environment Interface),
describing the encoding of syscalls.

Calling convention:

- An ECALL is used as the control transfer instruction between the supervisor and the SEE.
- a7 encodes the SBI extension ID (EID),
- a6 encodes the SBI function ID (FID) for a given extension ID encoded in a7 for any SBI extension defined in or after
  SBI v0.2.
- All registers except a0 & a1 must be preserved across an SBI call by the callee.
- SBI functions must return a pair of values in a0 and a1, with a0 returning an error code. This is analogous to
  returning the C structure
  ```c
  struct sbiret {
      long error;
      long value;
  };
  ```

In the name of compatibility, SBI extension IDs (EIDs) and SBI function IDs (FIDs) are encoded as signed 32-bit
integers. When passed in registers these follow the standard above calling convention rules.

The Table 1 below provides a list of Standard SBI error codes.
Table 1. Standard SBI Errors

| Error Type                  |  Value |
|-----------------------------|--------|
| `SBI_SUCCESS`               | 0      |
| `SBI_ERR_FAILED`            | -1     |
| `SBI_ERR_NOT_SUPPORTED`     | -2     |
| `SBI_ERR_INVALID_PARAM`     | -3     |
| `SBI_ERR_DENIED`            | -4     |
| `SBI_ERR_INVALID_ADDRESS`   | -5     |
| `SBI_ERR_ALREADY_AVAILABLE` | -6     |
| `SBI_ERR_ALREADY_STARTED`   | -7     |
| `SBI_ERR_ALREADY_STOPPED`   | -8     |

An ECALL with an unsupported SBI extension ID (EID) or an unsupported SBI function ID (FID) must return the error code
`SBI_ERR_NOT_SUPPORTED`.

Every SBI function should prefer unsigned long as the data type. It keeps the specification simple and easily adaptable
for all RISC-V ISA types. In case the data is defined as 32bit wide, higher privilege software must ensure that it only
uses 32 bit data only.

The Linux kernel implements this here:
[`sbi.c`](https://github.com/torvalds/linux/blob/7c698440524117dca7534592db0e7f465ae4d0bb/arch/riscv/kernel/sbi.c#L25)

Linux doesn't work quite like the SBI though.
The SBI says ECALL returns `sbiret` which encodes a0 value encodes an error code, and a1 a value.
But Linux seems to be more minimal/general, and (at least for the relevant syscalls in the Go runtime) 
uses only the first `sbiret` field for both errors (special magic range) and valid values.

For example, with `brk`:
Returned addresses in a0 in range `(-4096, -1]` (a linux kernel reserved page that may never actually be exposed to unprivileged programs)
are considered error codes, including by the Go runtime, see [here](https://github.com/golang/go/blob/ebb572d82f97d19d0016a49956eb1fddc658eb76/src/runtime/sys_linux_riscv64.s#L446)

To be nice with the SBI we should still write back to both `a0` and `a1`, but simply write a `0` in `a1`.
