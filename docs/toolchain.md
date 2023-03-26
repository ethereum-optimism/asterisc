
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
ba
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

Looking at an ELF file:
```
readelf -a myfile
```


