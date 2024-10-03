set -eo pipefail

# Run at rvsol/

# foundry pre-funded account #10: finalSystemOwner(0xa0Ee7A142d267C1f36714E4a8F75612F20a79720)
DEPLOY_PRIVATE_KEY="${DEPLOY_PRIVATE_KEY:-"0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"}"
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
