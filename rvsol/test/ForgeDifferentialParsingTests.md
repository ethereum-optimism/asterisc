# Rvsol Diff Tests 

To run diff tests directly: 

```bash
make fuzz-ffi
```

or 

```bash
cd rvsol
forge test --match-path ./test/Yul64Test.sol -vvvvv
```

`slow` should be built into the `./rvgo/scripts/parse-diff/ffi` directory. If changes here are necessary, ensure that `diff.go` print with no new lines, else Foundry will interpret the ffi result differently. A `/// @dev` note regarding this has been added to the `generateCommand` function in `Yul64Test.sol`. 

## Setup

solc 0.8.15 is required to run forge diff parsing tests.

You can use solc-select to download solc with the following script.
```bash
    # Create a python virtual environment
    python3 -m venv path/to/venv
    source path/to/venv/bin/activate
    # Install solc-select 
    python3 -m pip install solc-select 
    # Install solc 0.8.15 
    solc-select install 0.8.15
    # Set 0.8.15 as version in use 
    solc-select use 0.8.15
```

# Structure 

## `Yul64Test.sol` 

This is the entrypoint to implementing differential testing between golang and EVM. This contract is in charge of: 
- setting up the contract calls
- defining the fuzz functions
- calling EVM implementations
- calling golang implementations 
- asserting that the two implementations behave similarly 

If one implementation reverts and the other is successful, or if the result from golang does not match that of Solidity, these assertions will throw an error.

## `Yul64.yul`

This is the `yul` contract defined to be able to invoke parsing functions directly. The yul contract has the following properties: 
- The selectors match those defined in `RISCV.sol`
- Functions consistently take `uint64` as input, and return `uint256`
- Encoding function is used to return `uint256`
- Decoding function is used to decode input as `uint64`
- If selector does not match expected parse function, Solidity reverts 

## `slow`

This is the output from `rvgo/scripts/parse-diff-ffi` - which also allows us to call the golang functions directly for parsing functions. 

One notable difference between this and Solidity is the input – golang **only** provides a `toU64` function to convert a number into the representation required to invoke the `ParseX` functions - but this needs to truncate, unlike Solidity. This is likely the cause of a few differences that can be seen while running the tests. 
