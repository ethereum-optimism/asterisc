# Asterisc VM contract

## Deployment
Currently, Asterisc only supports the local devnet launched from the Optimism monorepo.

### Prerequisites
1. Running local devnet launched from the Optimism monorepo
   - Run ```make devnet-up``` in the monorepo root.
   - Set the env var `TARGET_L1_RPC_URL` to the L1 RPC endpoint
   - Set the env var `TARGET_L2_DEPLOYMENT_FILE` to the path of devnet deployment file (`optimism/packages/contracts-bedrock/deployments/devnetL1/.deploy`)
   - Set the env var `TARGET_L2_DEPLOY_CONFIG` to the path of devnet deploy config file (`optimism/packages/contracts-bedrock/deploy-config/devnetL1.json`)
2. Asterisc absolute prestate of op-program
   - Run ```make op-program-client-riscv``` in `optimism/op-program`
   - Set the built elf file path to the env var `OP_PROGRAM_PATH`

### How to deploy
1. Build Asterisc binary and contracts
   - Run ```make build``` in the project root
2. Generate prestate proof
   - Run ```make prestate``` in the project root
3. Run deploy script
   - Run ```./rvsol/scripts/deploy.sh``` in the project root
