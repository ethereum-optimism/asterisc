FROM ubuntu:22.04

RUN apt-get update && apt-get install --assume-yes --no-install-recommends \
    ca-certificates \
    build-essential \
    curl \
    g++-riscv64-linux-gnu \
    libc6-dev-riscv64-cross \
    binutils-riscv64-linux-gnu \
    llvm \
    clang \
    make \
    cmake \
    git 

COPY ./riscv64g-unknown-linux-gnu.json .

ENV SHELL=/bin/bash
RUN curl https://sh.rustup.rs -sSf | bash -s -- -y --default-toolchain nightly --component rust-src
ENV PATH="/root/.cargo/bin:${PATH}"

ENV CC_riscv64g_unknown_linux_gnu=riscv64-linux-gnu-gcc \
    CXX_riscv64g_unknown_linux_gnu=riscv64-linux-gnu-g++ \
    CARGO_TARGET_RISCV64G_UNKNOWN_LINUX_GNU_LINKER=riscv64-linux-gnu-gcc \
    CARGO_TARGET_RISCV64G_UNKNOWN_LINUX_GNU_RUSTFLAGS="-C target-feature=+crt-static,+a,+m,+f,+d" \
    CARGO_BUILD_TARGET="/riscv64g-unknown-linux-gnu.json"
