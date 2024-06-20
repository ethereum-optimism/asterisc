#!/bin/bash
set -e

TARGET_L1_RPC_URL="${TARGET_L1_RPC_URL:-"http://localhost:8545"}"

if [ -z "$1" ]; then
  echo "Usage: $0 <network-name>"
  exit 1
fi

output_root=$(cast call $(jq -r .AnchorStateRegistryProxy addresses/$1.json) --rpc-url $TARGET_L1_RPC_URL "anchors(uint32)(bytes32,uint256)" 0 --block finalized)

fault_game_genesis_output_root=$(echo "$output_root" | sed -n '1p')
fault_game_genesis_block=$(echo "$output_root" | sed -n '2p' | awk '{print $1}')

echo "network: $1"
echo "faultGameGenesisBlock: $fault_game_genesis_block"
echo "faultGameGenesisOutputRoot: $fault_game_genesis_output_root"

fault_proof_program=op-program
fault_proof_program_version=$(jq -r --arg program $fault_proof_program --arg network $1 '.[$network][$program]' fault-proof-program-version.json)
fault_game_absolute_prestate=$(jq -r --arg program $fault_proof_program --arg version $fault_proof_program_version '.[$program][$version]' ../prestates.json)

if [ -z "$fault_game_absolute_prestate" ] || [ "$fault_game_absolute_prestate" == "null" ]; then
    echo "Error: faultGameAbsolutePrestate is null or not found."
    exit 1
fi

echo "fault proof program with version: $fault_proof_program $fault_proof_program_version"
echo "faultGameAbsolutePrestate: $fault_game_absolute_prestate"

jq -n --arg fault_game_genesis_block $fault_game_genesis_block \
    --arg fault_game_genesis_output_root $fault_game_genesis_output_root \
    --arg fault_game_absolute_prestate $fault_game_absolute_prestate \
    '{faultGameGenesisBlock: ($fault_game_genesis_block | tonumber), faultGameGenesisOutputRoot: $fault_game_genesis_output_root, faultGameAbsolutePrestate: $fault_game_absolute_prestate}' \
    > anchor-state/$1.json
