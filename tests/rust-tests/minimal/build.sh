#!/bin/bash

# GNU target (we want this one to work 🤞)
CROSS_CONTAINER_OPTS="--platform linux/amd64" \
  CARGO_TARGET_RISCV64GC_UNKNOWN_LINUX_GNU_RUSTFLAGS="-C target-feature=-c,+crt-static" \
  cross build --release --target riscv64gc-unknown-linux-gnu -Z build-std

# Custom target
# docker build -t riscv-builder -f rust-riscv-builder.dockerfile .
# docker run --rm -v `pwd`/.:/code -w="/code" --platform linux/amd64 riscv-builder cargo build --release -Z build-std

# Bare Metal
# CARGO_TARGET_RISCV64GC_UNKNOWN_NONE_ELF_RUSTFLAGS="-C target-feature=+crt-static,-c" \
#   cross build --release --target riscv64gc-unknown-none-elf -Z build-std=panic_abort,std
