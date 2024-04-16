#!/bin/bash
set -e

if [ -z "${PYTHON_PATH}" ]; then
    echo "PYTHON_PATH is not set. Must point Python binary to run capture.py"
    exit 1
fi
absolute_python_path="$(cd "$(dirname "$PYTHON_PATH")"; pwd)/$(basename "$PYTHON_PATH")"
script_dir=$(cd "$(dirname $0)"; pwd)
root_dir=$(dirname $(dirname $script_dir))
optimism_dir=$root_dir/rvsol/lib/optimism

# Build asterisc binary
cd $root_dir
make build-rvgo
cp ./rvgo/bin/asterisc $script_dir/

# Build op-program
cd $optimism_dir
git_commit_hash=$(git rev-parse HEAD)
make -C op-program op-program-client-riscv
make -C op-program op-program-host
cp op-program/bin/op-program-client-riscv.elf $script_dir/
cp op-program/bin/op-program $script_dir/

# Launch devnet
make devnet-up

# Copy devnet artifacts
cp .devnet/rollup.json $script_dir/chain-artifacts/
cp .devnet/genesis-l2.json $script_dir/chain-artifacts/

# Load op-program RISCV binary
cd $script_dir
./asterisc load-elf --path=./op-program-client-riscv.elf

# Make op-program scripts
$absolute_python_path capture.py

# Capture preimages
rm ./preimages/*
./capture_cmd.sh

# Clean up
rm ./capture_cmd.sh ./asterisc ./op-program ./op-program-client-riscv.elf ./out.json

# Write optimism version
echo $git_commit_hash > VERSION
