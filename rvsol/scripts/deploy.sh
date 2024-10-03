#!/usr/bin/env bash
set -eo pipefail

TARGET_L1_RPC_URL="${TARGET_L1_RPC_URL:-"http://localhost:8545"}"
# foundry pre-funded account #10: finalSystemOwner(0xa0Ee7A142d267C1f36714E4a8F75612F20a79720)
DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"}"
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
TARGET_L2_DEPLOY_CONFIG=$LOCAL_DEPLOY_CONFIG_PATH forge script -vvv "${SCRIPTS_DIR}"/Deploy.s.sol:Deploy --rpc-url "$TARGET_L1_RPC_URL" --broadcast --private-key "$DEPLOY_PRIVATE_KEY"
