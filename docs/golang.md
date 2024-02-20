# Go support

To run Go, we must support a RISC-V runtime in Go.

[TinyGo](https://tinygo.org/) has nice minimal riscv [targets](https://github.com/tinygo-org/tinygo/blob/release/targets/riscv64.json) with small largely-optional runtime for embedded systems,
but is not mature enough.

Another path forward in the future may be to implement an alternative `riscv64_ethereum` runtime in Go,
largely based on the javascript "OS" from the `wasm_js` runtime, to make a lot of system interaction no-ops,
and to fake concurrency with the weird parking/recovering that is done there.

So instead of forking the Go runtime, or using a fork like TinyGo,
for now we can try to support the `riscv64_linux` runtime of the official Go compiler.


## Initialization

Steps:
1. Compile Go with `riscv64_linux` target
2. Take ELF binary output, and concatenate all sections, with filling to mem-size etc. where necessary: i.e. pre-process the ELF-loader steps.
3. If concurrency is not supported, we must replicate the hack from geohotz in Cannon to make the GC start function in the Go runtime a no-op, 
   during the ELF processing this can be done based on inspection of program symbols and patching `runtime.gcenable` to immediately jump to the address in the return-address register (`ra`).
4. Prepare the stack:
   - After `0x7f_ff_d0_00` in memory, lay out the stack:
     - `argc = 0`
     - `argv[n] = 0`
     - `envp[term] = 0`
     - `auxv[0] = _AT_PAGESZ = 6`
     - `auxv[1] = 4 KiB = 4096`
     - `auxv[2] = _AT_RANDOM = 25`
     - `auxv[3] = address to 16 bytes of randomness`
     - `auxv[term] = 0`
     - `16 bytes of randomness`
   - Initialize stack pointer to top of stack
5. Merkleize the binary, this will be the cartridge we slot into the VM

## AUX vectors used by Go

See [`os_linux.go`](https://github.com/golang/go/blob/1a9893a969e0a73dc4f1e48ed40ccaf29ec238a6/src/runtime/os_linux.go#L215):
- `_AT_NULL`: End of vector
- `_AT_PAGESZ = 6`: System physical page size. Go requires [4KB minimum physical page size](https://github.com/golang/go/blob/1a09d57de58a90987789ef637083aac21533eeb7/src/runtime/mheap.go#L23).
- `_AT_HWCAP  = 16`, `_AT_HWCAP2 = 26`: hardware capability bit vector, and vector 2:
  The `internal/cpu` package in Go uses these to determine abilities like AES, SHA2, etc.
  However, the RISC-V target doesn't support any hardware accellerated functions,
  [the `doinit()` is empty](https://github.com/golang/go/blob/feb984c96b10900daade4b47c4d308d7dd4ed5c3/src/internal/cpu/cpu_riscv64.go#L9).
- `_AT_RANDOM = 25`: The `AT_RANDOM` vector is a pointer to 16 bytes of data, initialized by the Go runtime
  [here](https://github.com/golang/go/blob/db36eca33c389871b132ffb1a84fd534a349e8d8/src/runtime/os_linux.go#L284)
  to [initialize](https://github.com/golang/go/blob/0b323a3c1690050340fc8e39730a07bb01373f0a/src/runtime/proc.go#L867)
  the "fast random" with, i.e. the randomness used by maps iteration, as well as the `hashkey` used by `hash{32,64}.go`


## Linux syscalls used by Go

The Go runtime defines the following linux syscalls to be used by Go:
[`sys_linux_riscv64.s`](https://github.com/golang/go/blob/master/src/runtime/sys_linux_riscv64.s)

And table of all syscalls:
https://github.com/golang/go/blob/master/src/syscall/zsysnum_linux_riscv64.go
Note that some are available to std-lib, but not used by the runtime, and thus present here but not in the runtime bindings.

To read the Go assembler, see [this Go asm syntax doc](https://go.dev/doc/asm) (it's not as complete, but one of few resources).

By supporting a minimal subset of these, most Go programs can be proven.
The GC won't have to be disabled if concurrency is supported, and will then avoid growing the memory indefinitely.

Note that hardware-accelerated AES hashing is not supported by the riscv64 runtime,
fallback functions [are used instead](https://github.com/golang/go/blob/0b323a3c1690050340fc8e39730a07bb01373f0a/src/runtime/asm_riscv64.s#L222). 

Errors and file descriptor flags used by the syscalls can be found in https://github.com/golang/go/blob/master/src/syscall/zerrors_linux_riscv64.go

System calls used by `linux_riscv64` Go runtime:

```text
# Memory: must-have
SYS_brk			214
SYS_mmap		222

# exits: must-have
SYS_exit		93
SYS_exit_group		94

# Concurrency
# ------------------------------------
# Threads: necessary for GC to work, and concurrency support!
SYS_clone		220

# locking: to handle state of threads
SYS_futex		98

# sleeping: maybe? Can be simple
SYS_nanosleep		101

# to put the thread at end of queue, maybe useful, could be no-op
SYS_sched_yield		124

# thread id: if threads support
SYS_gettid		178

# sending signals to threads (to close them)
SYS_tgkill		131
SYS_tkill		130


# File reads/writes may be useful for preimage oracle interaction
# Otherwise fine to not support.
# ------------------------------------
#
# file reading, can return small number to limit reads
SYS_read		63
# file writing, can also return smaller number, but may cause errors
SYS_write		64


# Time: maybe useful to fake based on instruction counter
# ------------------------------------

# Timers   (does the GC use these?)
SYS_setitimer		103
SYS_timer_create	107
SYS_timer_delete	111
SYS_timer_settime	110

# Clocks
SYS_clock_gettime	113
SYS_gettimeofday	169


# To simplify/hardcode
# -----------------------
# input/output readiness. Input never ready, output always ready
SYS_pselect6		72
# Hardcode the process ID
SYS_getpid		172
# "in core" = if page is in memory and not in disk, can be hardcoded
SYS_mincore		232
# set/get CPU affinity mask: keep this simple
SYS_sched_getaffinity	123

# NOOP
# -----------------------
# program advises kernel how to use memory, can be no-op
SYS_madvise		233

# NOOP - To not support, but needed to run op-program
# May need more investigation
# -----------------------
# file memory mapping
SYS_munmap		215

# interprocess communication
SYS_pipe2		59

SYS_epoll_create1       20
SYS_epoll_ctl           21
SYS_readlinkat          78
SYS_newfstatat          79
SYS_newuname            160
SYS_getrandom           278

# To not support
# -----------------------
# sockets
SYS_connect		203
SYS_socket		198

# files closing/opening/stats
SYS_close		57
SYS_openat		56
SYS_faccessat		48

# send a signal to another process
SYS_kill		129

# change action taken on signal
SYS_rt_sigaction	134
# fetch or change signal mask
SYS_rt_sigprocmask	135
# signal trampoline
SYS_rt_sigreturn	139
# to specify an alternate signal stack
SYS_sigaltstack		132

```

Additionally, syscalls to support for Go std-lib:

```text
# file descriptor manipulation with flags - support flag lookups
# 
SYS_fcntl 25
```


## RISC-V Instructions used by Go

Instructions used by Go compiler:[`internal/obj/riscv/cpu.go`](https://github.com/golang/go/blob/38cfb3be9d486833456276777155980d1ec0823e/src/cmd/internal/obj/riscv/cpu.go#L278)

TLDR: extensions: A and G from unprivileged spec,
and some things from the privileged instruction set (separate spec)

```text
2.4: Integer Computational Instructions
ADDI, SLTI, SLTIU, ANDI, ORI, XORI, SLLI, SRLI, SRAI, LUI, AUIPC, ADD, SLT, SLTU, AND, OR, XOR, SLL, SRL, SUB, SRA

"The SLL/SRL/SRA instructions differ slightly between RV32 and RV64"
SLLIRV32, SRLIRV32, SRAIRV32

2.5: Control Transfer Instructions
JAL, JALR, BEQ, BNE, BLT, BLTU, BGE, BGEU

2.6: Load and Store Instructions
LW, LWU, LH, LHU, LB, LBU, SW, SH, SB

2.7: Memory Ordering Instructions
FENCE, FENCEI, FENCETSO

5.2: Integer Computational Instructions (RV64I)
ADDIW, SLLIW, SRLIW, SRAIW, ADDW, SLLW, SRLW, SUBW, SRAW

5.3: Load and Store Instructions (RV64I)
LD, SD

7.1: Multiplication Operations
MUL, MULH, MULHU, MULHSU, MULW, DIV, DIVU, REM, REMU, DIVW, DIVUW, REMW, REMUW

8.2: Load-Reserved/Store-Conditional Instructions
LRD, SCD, LRW, SCW

8.3: Atomic Memory Operations
AMOSWAPD, AMOADDD, AMOANDD, AMOORD, AMOXORD, AMOMAXD, AMOMAXUD, AMOMIND, AMOMINUD, AMOSWAPW, AMOADDW, AMOANDW, AMOORW, AMOXORW, AMOMAXW, AMOMAXUW, AMOMINW, AMOMINUW

10.1: Base Counters and Timers
RDCYCLE, RDCYCLEH, RDTIME, RDTIMEH, RDINSTRET, RDINSTRETH

Floating point ops, no need to support in fraud prover:
11.2: Floating-Point Control and Status Register
11.5: Single-Precision Load and Store Instructions
11.6: Single-Precision Floating-Point Computational Instructions
11.7: Single-Precision Floating-Point Conversion and Move Instructions
11.8: Single-Precision Floating-Point Compare Instructions
11.9: Single-Precision Floating-Point Classify Instruction
12.3: Double-Precision Load and Store Instructions
12.4: Double-Precision Floating-Point Computational Instructions
12.5: Double-Precision Floating-Point Conversion and Move Instructions
12.6: Double-Precision Floating-Point Compare Instructions
12.7: Double-Precision Floating-Point Classify Instruction
13.1 Quad-Precision Load and Store Instructions
13.2: Quad-Precision Computational Instructions
13.3 Quad-Precision Convert and Move Instructions
13.4 Quad-Precision Floating-Point Compare Instructions
13.5 Quad-Precision Floating-Point Classify Instruction

Privileged ISA (Version 20190608-Priv-MSU-Ratified)
3.1.9: Instructions to Access CSRs
CSRRW, CSRRS, CSRRC, CSRRWI, CSRRSI, CSRRCI

3.2.1: Environment Call and Breakpoint
ECALL, SCALL, EBREAK, SBREAK

3.2.2: Trap-Return Instructions
MRET, SRET, URET, DRET

3.2.3: Wait for Interrupt
WFI

4.2.1: Supervisor Memory-Management Fence Instruction
SFENCEVMA

Hypervisor Memory-Management Instructions
HFENCEGVMA, HFENCEVVMA

The escape hatch. Inserts a single 32-bit word.
WORD
```

fresh reg reads:
runtime.rt0_go: 1

0xC000000000 - called by mmap

