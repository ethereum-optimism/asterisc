#!/bin/bash
set -e

if [ -z "$ETHERSCAN_API_KEY" ]; then
  echo "ETHERSCAN_API_KEY is not set"
  exit
fi

echo "This script must be executed at {PROJECT_ROOT}/rvsol"
cd ../../../rvsol

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(uint32, bytes32, uint256, uint256, uint64, uint64, address, address, address, uint256)" 2 0x0338dc64405def7e3b9ce8f5076b422a846d831832617d227f13baf219cb5406 73 30 10800 302400 0xfD70c0391eC0308445c2BE83F85689aF72946332 0xF3D833949133e4E4D3551343494b34079598EA5a 0x1a31b297355a30752D5E2A4AC3e9DC47D02485c2 11155420) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x71838a0d3b3226842166619338ab3218c62bb4a1 \
    lib/optimism/packages/contracts-bedrock/src/dispute/FaultDisputeGame.sol:FaultDisputeGame

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor()" ) \
    --etherscan-api-key $ETHERSCAN_API_KEY  \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x872507195C86Da21A9db7b9eD2589d21f20CB61E \
    lib/optimism/packages/contracts-bedrock/src/dispute/DisputeGameFactory.sol:DisputeGameFactory

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x1307f44beDCdCaeA0716abb2016c1792ed310f46) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0xB93884E2C21a78d036f213eA4f77Eb7e94147065 \
    lib/optimism/packages/contracts-bedrock/src/dispute/AnchorStateRegistry.sol:AnchorStateRegistry

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x1a31b297355a30752D5E2A4AC3e9DC47D02485c2 \
    lib/optimism/packages/contracts-bedrock/src/universal/Proxy.sol:Proxy

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x1307f44beDCdCaeA0716abb2016c1792ed310f46 \
    lib/optimism/packages/contracts-bedrock/src/universal/Proxy.sol:Proxy

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x7DaE43a953C40d371C4bE9963FdD250398a4915A) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1 \
    lib/optimism/packages/contracts-bedrock/src/universal/ProxyAdmin.sol:ProxyAdmin 

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x7DaE43a953C40d371C4bE9963FdD250398a4915A) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0x2F4613Aa09634CD07f39b3ee91FB9b1ca57b94C1 \
    lib/optimism/packages/contracts-bedrock/src/universal/ProxyAdmin.sol:ProxyAdmin 

forge verify-contract \
    --chain-id 11155111 \
    --num-of-optimizations 999999 \
    --watch \
    --constructor-args $(cast abi-encode "constructor(address)" 0x627F825CBd48c4102d36f287be71f4234426b9e4) \
    --etherscan-api-key $ETHERSCAN_API_KEY \
    --compiler-version 0.8.15+commit.e14f2714 \
    0xfD70c0391eC0308445c2BE83F85689aF72946332 \
    src/RISCV.sol:RISCV 
