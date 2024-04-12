set -eo pipefail

# Run at rvsol/

DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"}" # foundry pre-funded account #1
ASTERISC_PRESTATE="${ASTERISC_PRESTATE:-"/../rvgo/bin/prestate-proof.json"}"

if [ -z "${TARGET_L2_DEPLOYMENT_FILE}" ]; then
    echo "TARGET_L2_DEPLOYMENT_FILE is not set. Must point target chain deployment file (optimism/packages/contracts-bedrock/deployments/devnetL1/.deploy)"
    exit 1
fi

if [ -z "${TARGET_L2_DEPLOY_CONFIG}" ]; then
    echo "TARGET_L2_DEPLOY_CONFIG is not set. Must point target chain deploy config file (optimism/packages/contracts-bedrock/deploy-config/devnetL1.json)"
    exit 1
fi

if [ -z "${TARGET_L1_ALLOC}" ]; then
    echo "TARGET_L1_ALLOC is not set. Must point target chain l1 alloc file."
    exit 1
fi

forge script --chain-id 900 scripts/Deploy.s.sol --sig "runForDevnetAlloc()" --private-key "$DEPLOY_PRIVATE_KEY"
