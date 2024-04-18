# op-program-test

This directory has tools to run op-program-test in the CI workflow.

The purpose of the test is to confirm that Asterisc fast VM can run the current version of op-program on the local devnet.

To run op-program in CI, we need following requirements:
- chain-artifacts: rollup config and L2 genesis of devnet
- captured preimages: preimage key and values which are used by op-program

These requirements should be generated in the local machine when `rvsol/lib/optimism` submodule is updated.

To generate requirements, simply run `PYTHON_PATH={PYTHON_BIN_PATH} capture.sh` in the local machine, 
and commit all generated files!

## Prerequisites
- All prerequisites to build the optimism monorepo and launch devnet. See https://github.com/ethereum-optimism/optimism/blob/develop/CONTRIBUTING.md#development-quick-start.
- Python environment(>3.10) which has all dependencies in `requirements.txt`.
