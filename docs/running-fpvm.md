## Running the FPVM - Asterisc

This guide provides a comprehensive walkthrough on how to run the Fault Proof Program (FPP) using the Fault Proof Virtual Machine (FPVM), specifically Asterisc. 
The goal is to validate an output root by executing the FPP from a known L2 block head to a target block, ensuring that the output root matches the claimed value.

First, read more about Fault Proof programs from the following resources:
- https://github.com/ethereum-optimism/specs/blob/main/specs/fault-proof/index.md#fault-proof
- https://www.youtube.com/watch?v=RGts3PSg4F8

Use this guide as a way to actually run the fault proof system in a more granular manner.

## Prerequisite
Before you begin, make sure you have the following:
- `L1 Execution Client Endpoint`: Access to an L1 execution client (e.g., Geth, Erigon).
- `L1 Beacon Client Endpoint`: Access to an L1 beacon client (e.g., Lighthouse, Prysm).
- `L2 Execution Client Endpoint`: Access to an L2 execution client (e.g., op-geth).
- `OP Node` Running on Your Selected Chain: An operational op-node connected to your chosen network.

## Gathering the necessary hashes

A fault proof program starts from a specific l2 head, and runs the fault proof program until another l2 head in the future. 
Then, we want to validate that the l2 output root the output root generated is equal to the claim we want to verify against.

Take a look at the following command:
```bash
./bin/op-program \
  --datadir=/data2/op-mainnet-preimage/123993796_123993889_20522562 \
  --network=op-mainnet \
  --l2.blocknumber=123993889 \
  --l2.claim=0xa0a0b59f6ef9658d3f0614a97e05958975ed5c3e1e2c80cf0ed674a0bce3b467 \
  --l2.outputroot=0x15ea3bef4bfe4fafbb832431f53c74852229aca16b741a52e17d47ba4271ef54 \
  --l2.head=0xe21cf8b2d24b8f9ffd3fc4748ff618192bfdee03b1ca6fc8f4a7ec3cfe70f903 \
  --l1.head=0x86f4b04b8fc9c614a51423f6e521160ca7214da5608a458e6928fb7442c613f1 \
  --l1=http://L1_EXECUTION_ENDPOINT \
  --l1.beacon=http://L1_BEACON_ENDPOINT \
  --l1.trustrpc=true \
  --l2=http://L2_EXECUTION_ENDPOINT
```
The following command…

- inputs L1 data and L2 data by `l1, l1.beacon, l2` …
- starts at L2 block with block hash `l2.head` with output root `l2.outputroot` …
- reads L1 blocks until `l1.head` …
- derives until `l2.blocknumber` …
- checks the final output root is equal to `l2.claim`.

### Gathering the l2 head and output root
Pick a l2 block number to start derivation from. 

For example, if we start from block `123,993,796`, call the following rpc on your op-node RPC: 
```bash
cast rpc optimism_outputAtBlock 0x763fec4 --rpc-url http://OP_NODE_ENDPOINT | jq
```

This will return something like (but not exactly): 
```json
{
  "version": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "outputRoot": "0x15ea3bef4bfe4fafbb832431f53c74852229aca16b741a52e17d47ba4271ef54",
  "blockRef": {
    "hash": "0xc97424185ea025078785bc037d0931dd0b7173ffcda1289ba82fc8c4bd29882a",
    "number": 127275901,
    "parentHash": "0xbdf74f76b015d6ffcf73798128227feaf1525484f5b5b22e2b7adec14172a3b7",
    "timestamp": 1730150579,
    "l1origin": {
      "hash": "0x4c8ea4740900532e1fb5389ecc411d312ca43d9d513bde7314c183b0e8f02e20",
      "number": 21066777
    },
    "sequenceNumber": 4
  },
  "withdrawalStorageRoot": "0xf34e1185b11494a8887490ec951a1dc814327fd54a36bd6e39434393beee05ea",
  "stateRoot": "0x6812727536d8cb39d86fdfb4518b3a7c9e1e5bea34da3560209f11b019d35484"
  // ... 
}
```

The `outputRoot` and `blockRef.hash` are what we're interested in. These values correspond to the `l2.head` and `l2.outputroot` in the previous command.

### Gathering the target l2 head and output root
Now, pick a l2 block number to end derivation at. 

For example, if we start from block `123,993,889`, call the following rpc on your op-node RPC:
```bash
cast rpc optimism_outputAtBlock 0x763FF21 --rpc-url http://OP_NODE_ENDPOINT | jq
```

From the result, get the `outputRoot`, `blockRef.number`, and `blockRef.l1origin.hash`. These values correspond to the `l2.claim`, `l2.blocknumber`, `l1.head`.

### Gathering the preimage

Now, let's return to the original command we were working on:
```bash
./bin/op-program \
  --datadir=/data2/op-mainnet-preimage/123993796_123993889_20522562 \
  --network=op-mainnet \
  --l2.blocknumber=123993889 \
  --l2.claim=0xa0a0b59f6ef9658d3f0614a97e05958975ed5c3e1e2c80cf0ed674a0bce3b467 \
  --l2.outputroot=0x15ea3bef4bfe4fafbb832431f53c74852229aca16b741a52e17d47ba4271ef54 \
  --l2.head=0xe21cf8b2d24b8f9ffd3fc4748ff618192bfdee03b1ca6fc8f4a7ec3cfe70f903 \
  --l1.head=0x86f4b04b8fc9c614a51423f6e521160ca7214da5608a458e6928fb7442c613f1 \
  --l1=http://L1_EXECUTION_ENDPOINT \
  --l1.beacon=http://L1_BEACON_ENDPOINT \
  --l1.trustrpc=true \
  --l2=http://L2_EXECUTION_ENDPOINT
```

The above command uses op-program to [gather the pre-image](https://github.com/ethereum-optimism/specs/blob/main/specs/fault-proof/index.md#pre-image-oracle), which are necessary for the FPVM to work with.

The op-program will run from `l2.head` to `l2.blocknumber`, and save the preimage data to the `--datadir` we specified.

Once this is ready, we can run the Asterisc FPVM

## Building the Fault Proof VM 

What is a FPVM?
- First, the op-program is compiled down to a binary in RISC-V format.
- Then, this ELF binary is fed into the Asterisc, which produced a json format of the program (prestate)
- Finally, Asterisc can execute through this json

### Compiling op-program into RISC-V
Navigate to `rvsol/lib/optimism/op-program`, and run:
```bash
make op-program-client-riscv
```

### Generating the Prestate JSON
Use Asterisc to translate the ELF binary into a JSON format that the FPVM can execute:

```bash
./bin/asterisc load-elf \
  --path ./bin/op-program-client-riscv.elf \
  --out ./bin/prestate.json \
  --meta ./bin/meta.json
```

This generates:
- `prestate.json`: The translated program in JSON format.
- `meta.json`: Metadata about the program.

You can also run `make prestate` at the root directory to run the above steps. 

### Running Asterisc
To run Asterisc with the prestate generated in the above step, run:
```bash
./bin/asterisc run \
    --info-at=%100000000 \
    --proof-at=never \
    --input=./bin/prestate.json \
    --meta=./bin/meta.json
```

Note that the `prestate.json` is now provided as the `input`.

## Running the Fault Proof Program with Asterisc
Now, run the FPP within the FPVM, using the prestate JSON and the pre-images gathered earlier.

```bash
./bin/asterisc run \
  --info-at=%100000000 \
  --proof-at=never \
  --input=./bin/prestate.json \
  --meta=./bin/meta.json \
  -- ./bin/op-program \
    --datadir=/data2/op-mainnet-preimage/123993796_123993889_20522562 \
    --network=op-mainnet \
    --l2.blocknumber=123993889 \
    --l2.claim=0xa0a0b59f6ef9658d3f0614a97e05958975ed5c3e1e2c80cf0ed674a0bce3b467 \
    --l2.outputroot=0x15ea3bef4bfe4fafbb832431f53c74852229aca16b741a52e17d47ba4271ef54 \
    --l2.head=0xe21cf8b2d24b8f9ffd3fc4748ff618192bfdee03b1ca6fc8f4a7ec3cfe70f903 \
    --l1.head=0x86f4b04b8fc9c614a51423f6e521160ca7214da5608a458e6928fb7442c613f1 \
    --l1=http://144.76.43.227:8545 \
    --l1.beacon=https://restless-muddy-fog.quiknode.pro/7b1bbe64477e404159509e7e55db8a89aaba9269 \
    --l2=http://actually-central-quail.n0des.xyz \
    --l1.rpckind=erigon \
    --l1.trustrpc=true \
    --server
```
### Breaking Down the Command
- `./bin/asterisc run`: Executes the FPVM with the prestate.
- The arguments after `--` are passed to the server part of op-program.
- `./bin/op-program` runs the op-program.
  - The `--server` flag runs op-program in server mode, providing pre-images. 

## Validating the Output Root
After running the above command, the FPVM should output whether the final output root matches the claimed output root.
- Success: If the output root matches, the claim is valid.
- Failure: If it doesn't match, there may be a fault in the state transition.

At the end of the execution, you want to look for something like: 
```
Successfully validated L2 block #17484899 with output root 0xe9d2dddb6badcb5060a653c90657c2f95f0017febe284b7bfb97ff12adf8b0f1
```