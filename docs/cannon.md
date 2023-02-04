# Cannon

Notes about [Cannon](https://github.com/ethereum-optimism/cannon/), and how it compares to Asterisc.

Cannon instructions:
```
addi, addiu, addu, sub, subu,   -- binary math
div, divu,                      -- division
mul, mult, multu, mfhi, mflo,   -- multiplication
and, andi, xor, xori, or, ori, nor  -- binary bit ops
clz, clo,                       -- unary bit ops
beq, bgez, blez, bltz, bne,     -- branching
jalr, jr,                       -- jumps
movn, movz, mtlo, mthi,         -- moves
slt, slti, sltiu, sltu,         -- set on less than
sll, sllv, sra, srav, srl, srlv,-- shifting
lb, lbu, lui, lh, lhu, lw,      -- loads
sb, sh, sw,                     -- storing
lwl, lwr, swl, swr,             -- unaligned load/store
syscall,                        -- syscalls
ll, sc,                         -- atomic

sync,                           -- sync load/stores
nop,                            -- no-op instruction
```


Disassembling `startup.s`:
```
mipsel-linux-gnu-objdump -D -m mips:isa32 -EB -b binary startup.bin
```

Compiling `startup.s`:
```
mipsel-linux-gnu-gcc -nostdlib -EL -c startup.s -o startup.o
mipsel-linux-gnu-objdump -d startup.o
# TODO: objcopy the section to a binary file

```