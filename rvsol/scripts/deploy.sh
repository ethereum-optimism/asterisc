#!/usr/bin/env bash
set -eo pipefail

TARGET_L1_RPC_URL="${TARGET_L1_RPC_URL:-"http://localhost:8545"}"
DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"}" # foundry pre-funded account #1

if [ -z "${TARGET_L2_DEPLOYMENT_FILE}" ]; then
    echo "TARGET_L2_DEPLOYMENT_FILE is not set. Must point target chain deployment file (optimism/packages/contracts-bedrock/deployments/devnetL1/.deploy)"
    exit 1
fi

if [ -z "${TARGET_L2_DEPLOY_CONFIG}" ]; then
    echo "TARGET_L2_DEPLOY_CONFIG is not set. Must point target chain deploy config file (optimism/packages/contracts-bedrock/deploy-config/devnetL1.json)"
    exit 1
fi
LOCAL_DEPLOY_CONFIG_PATH="$(dirname "$(dirname "$(realpath "$0")")")/deploy-config.json"
cp "$TARGET_L2_DEPLOY_CONFIG" "$LOCAL_DEPLOY_CONFIG_PATH"


echo "> Deploying contracts"
TARGET_L2_DEPLOY_CONFIG=$LOCAL_DEPLOY_CONFIG_PATH forge script -vvv scripts/Deploy.s.sol:Deploy --rpc-url "$TARGET_L1_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
