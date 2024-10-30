# Docs

The Asterisc repository contains documentation files in different directories. Refer to this docs on where to find each documentation. 

## Golang 
Golang implementation of Asterisc resides in [rvgo](../rvgo).
There are two separate implementation:
- [Fast-mode](../rvgo/fast): Emulate 1 instruction per step, directly on a VM state
- [Slow-mode](../rvgo/slow): Emulate 1 instruction per step, on a VM state-oracle.

The Golang VM's memory layout is implemented as a radix-trie. Refer to [this docs](radix-memory.md) on the radix-memory layout.

Relevant information about the Go runtime / compiler can be found in [golang.md](golang.md)

## Solidity
Solidity implementation of Asteirsc resides in [rvsol](../rvsol).

## RISC-V
You can find more information regarding RISC-V resources, instruction set, spec, toolchains in the following documents.
- [riscv.md](riscv.md)
- [toolchain.md](toolchain.md)

## Deployment
For documentation and scripts regarding deploying Asterisc to a devnet or actual network, refer to:
- [deployments](../deployments)

## Testing
There are number of different tests in Asterisc.
1. [Go unit tests](../rvgo/fast)
2. [Go VM tests](../rvgo/test)
3. [Solidity VM tests](../rvsol/test)
4. [RISC-V implementation tests](../tests/riscv-tests)
5. [End to End tests with op-program](../op-e2e)

Refer to the `README.md` in the each directories for more details and prerequisites on running the individual tests.

## Running Asterisc with op-program
To actually run a Fault Proof Program with Asterisc on op-program, refer to [running-fpvm](./running-fpvm.md) guide. 

## Maintenance

Asterisc has a dependency on the [Optimism monorepo](https://github.com/ethereum-optimism/optimism). [This documentation](monorepo-sync.md) contains information on how to update the monorepo dependency.  