#!/bin/bash

# GNU target
CROSS_CONTAINER_OPTS="--platform linux/amd64" \
  CARGO_TARGET_RISCV64GC_UNKNOWN_LINUX_GNU_RUSTFLAGS="-C target-feature=-c,+crt-static -C opt-level=3" \
  cross build --release --target riscv64gc-unknown-linux-gnu

# Bare Metal
# CARGO_TARGET_RISCV64GC_UNKNOWN_NONE_ELF_RUSTFLAGS="-C target-feature=+crt-static,-c" \
#   cross build --release --target riscv64gc-unknown-none-elf -Z build-std=panic_abort,std
