#!/usr/bin/env bash
set -eo pipefail

MONOREPO_ROOT=./rvsol/lib/optimism

cp -r ${MONOREPO_ROOT}/.devnet .devnet
mkdir -p packages/contracts-bedrock
cp -r ${MONOREPO_ROOT}/packages/contracts-bedrock/deploy-config packages/contracts-bedrock
mkdir -p packages/contracts-bedrock/deployments/devnetL1
cp -r ${MONOREPO_ROOT}/packages/contracts-bedrock/deployments/devnetL1 packages/contracts-bedrock/deployments

# Generate L1 Allocs including asterisc
# copy everything locally due to foundry permission issues
cp ./rvgo/bin/prestate-proof.json ./rvsol/prestate-proof.json
cp -r packages/contracts-bedrock/deployments/devnetL1 ./rvsol/devnetL1
cp packages/contracts-bedrock/deploy-config/devnetL1.json ./rvsol/devnetL1.json
cp .devnet/allocs-l1.json ./rvsol/allocs-l1.json
cd ./rvsol && ASTERISC_PRESTATE=./prestate-proof.json \
TARGET_L2_DEPLOYMENT_FILE=./devnetL1/.deploy \
TARGET_L2_DEPLOY_CONFIG=./devnetL1.json \
TARGET_L1_ALLOC=./allocs-l1.json \
DEPLOYMENT_OUTFILE=./deployments/devnetL1/.deploy \
STATE_DUMP_PATH=./allocs-l1-asterisc.json \
./scripts/create_poststate_after_deployment.sh
cd ..
# Create address.json
jq -s '.[0] * .[1]' ./rvsol/devnetL1/.deploy ./rvsol/deployments/devnetL1/.deploy | tee .devnet/addresses.json
cp ./rvsol/allocs-l1-asterisc.json .devnet/allocs-l1.json
# Patch .deploy
cp .devnet/addresses.json packages/contracts-bedrock/deployments/devnetL1/.deploy
# Remove tmps
cd rvsol && rm -rf prestate-proof.json devnetL1 devnetL1.json allocs-l1.json deployments ./allocs-l1-asterisc.json
