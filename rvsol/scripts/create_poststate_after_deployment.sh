set -eo pipefail

# Run at rvsol/

# foundry pre-funded account #4: gnosis safe's owner(0x90F79bf6EB2c4f870365E785982E1f101E93b906)
DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6"}"
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
