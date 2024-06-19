#!/usr/bin/env bash
set -eo pipefail

TARGET_L1_RPC_URL="${TARGET_L1_RPC_URL:-"http://localhost:8545"}"
# Use fresh account to with enough funds to initialize FDG with bonds
# Default: Foundry Test Account #6 (address 0x976ea74026e726554db657fa54763abd0c3a0aa9)
DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e"}"

if [ -z "${TARGET_L2_DEPLOYMENT_FILE}" ]; then
    echo "TARGET_L2_DEPLOYMENT_FILE is not set. Must point target chain deployment file (optimism/packages/contracts-bedrock/deployments/devnetL1/.deploy)"
    exit 1
fi

if [ -z "${TARGET_L2_DEPLOY_CONFIG}" ]; then
    echo "TARGET_L2_DEPLOY_CONFIG is not set. Must point target chain deploy config file (optimism/packages/contracts-bedrock/deploy-config/devnetL1.json)"
    exit 1
fi
SCRIPTS_DIR="$(dirname "$(realpath "$0")")"
LOCAL_DEPLOY_CONFIG_PATH="$(dirname "${SCRIPTS_DIR}")/deploy-config.json"
cp "$TARGET_L2_DEPLOY_CONFIG" "$LOCAL_DEPLOY_CONFIG_PATH"


echo "> Deploying contracts"
TARGET_L2_DEPLOY_CONFIG=$LOCAL_DEPLOY_CONFIG_PATH forge script -vvv "${SCRIPTS_DIR}"/Deploy_Stage_1_4.sol:Deploy --rpc-url "$TARGET_L1_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
