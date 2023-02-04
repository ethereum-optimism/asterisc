
Install toolchain from source (need `riscv64-unknown-elf`, not the one in package repositories):
```shell
# Arch linux
sudo pacman -Syyu autoconf automake curl python3 libmpc mpfr gmp gawk base-devel bison flex texinfo gperf libtool patchutils bc zlib expat
# Other:
# see https://github.com/riscv-collab/riscv-gnu-toolchain#prerequisites

git clone git@github.com:riscv-collab/riscv-gnu-toolchain.git
cd riscv-gnu-toolchain

# This defaults to building RV64GC with glibc
./configure --prefix=/opt/riscv

# makes riscv64-unknown-elf
sudo make

# Optional: makes riscv-linux-gnu toolchain (not necessary for test gen)
# sudo make linux
```

Optional: Install "Spike" (RISCV simulator):

```shell
# apt-get install device-tree-compiler
sudo pacman -S dtc

git clone git@github.com:riscv-software-src/riscv-isa-sim.git
cd riscv-isa-sim

export RISCV=/opt/riscv

mkdir build
cd build
../configure --prefix=$RISCV
make
sudo make install
```


Building unit tests (requires `riscv64-unknown-elf` toolchain):
```shell
git clone git@github.com:riscv-software-src/riscv-tests.git
cd riscv-tests

export RISCV=/opt/riscv
git submodule update --init --recursive
autoconf
./configure --prefix=$RISCV/target
RISCV_PREFIX="/opt/riscv/bin/riscv64-unknown-elf-" make

# files are output to the isa/ directory, e.g.
ls isa/rv64ui-p-*
```

- `riscv_test.h` defines test environment things
- The "TVM" (test virtual machine) is the feature set required by a test
- We're only interested in `rv64ui`: 64 bit user-level integer-only instructions. We don't need the floating point, 32 bit, supervisor variants.
- And there are different target environments too. But we only care about single-core.
  - `p` = single core, physical memory
  - `v` = virtual memory enabled, may be interesting (TODO)
  - ignore multi-core tests

Test structure is documented here: https://riscv.org/wp-content/uploads/2015/01/riscv-testing-frameworks-bootcamp-jan2015.pdf
But lacks test format definition.

Test format (reverse engineered, cannot find docs):
- ELF file for each test *suite* to execute
- a `.dump` file with the matching assembly to understand the test
- entry-point seems to always be `0x80_00_00_00`
- test *case* is small part of it, and test number is contained in register `gp` during test
  - `1137` is used to indicate exceptions
- cases jump to `fail` if bad, with single-argument syscall calling convention to signal result:
  - `a0` is set to `(test_case_num << 1) | 1`
  - `a7` is set to the function ID, 93 (`exit`)
  - `ecall`, indicating test num and fail status
- if not to `fail`, we end at `pass`, again a syscall:
  - `a0` is set to `0`
  - `a7` is set to the function ID, 93 (`exit`)
  - `ecall`


## Other

There's also ["RISCOF"](https://github.com/riscv-software-src/riscof) which seems to be a newer test framework
focused on the architecture as a whole rather than unit testing the instructions.
Probably more suited for hardware implementations that optimize instruction execution.




