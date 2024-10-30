# Asterisc VM contract

## Deployment
Currently, Asterisc only supports the local devnet launched from the Optimism monorepo.

### Build
1. Update the git submodule
   - Run `git submodule update --init --remote` in the project root. 
2. Build the Asterisc binary and generate prestate proof
   - Run `make prestate` in the project root.   

### Testing
To run tests on [rvsol/src/RISCV.sol](../rvsol/src/RISCV.sol) implementation, run the following in the project root:
```
(cd rvgo/scripts/go-ffi && go build)

cd rvsol
forge test -vvv --ffi
```

### Prerequisites
1. Running local devnet launched from the Optimism monorepo
   - Run ```make devnet-up``` in the monorepo root.
   - Set the env var `TARGET_L1_RPC_URL` to the L1 RPC endpoint (Local devnet L1: `http://localhost:8545`)
   - Set the env var `TARGET_L2_DEPLOYMENT_FILE` to the path of devnet deployment file (`optimism/packages/contracts-bedrock/deployments/devnetL1/.deploy`)
   - Set the env var `TARGET_L2_DEPLOY_CONFIG` to the path of devnet deploy config file (`optimism/packages/contracts-bedrock/deploy-config/devnetL1.json`)
2. Asterisc absolute prestate of op-program
   - Run ```make op-program-client-riscv``` in `optimism/op-program
   - Set the built elf file path to the env var `OP_PROGRAM_PATH` (`optimism/op-program/bin/op-program-client-riscv.elf`)

### How to deploy
1. Build Asterisc binary and contracts
   - Run ```make build``` in the project root
2. Generate prestate proof
   - Run ```make prestate``` in the project root
3. Run deploy script
   - Run ```./scripts/deploy.sh``` in `rvsol`

### Notes
- There are few issues with Foundry.
  - Run script directly without manual build does not work with the current version of Foundry (2024-03-15 `3fa0270`). 
    You **must run** `make build` **before** running the deploy script. ([issue](https://github.com/foundry-rs/foundry/issues/6572))
  - Some older version(2024-02-01 `2f4b5db`) of Foundry makes a dependency error reproted above issue. 
    Use the **latest version** of Foundry!
- The deploy script can be run only once on the devnet because of the `create2` salt. 
  To rerun the script for dev purpose, you must restart the devnet with `make devnet-clean && make devnet-up` command on the monorepo.