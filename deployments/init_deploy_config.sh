#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <network-name>"
  exit 1
fi

l2=$(echo "$1" | cut -d"-" -f1)
l1=$(echo "$1" | cut -d"-" -f2)

input_file="../rvsol/lib/optimism/packages/contracts-bedrock/deploy-config/$l1.json"
output_file="deploy-config/$1.json"

if [ ! -f "$input_file" ]; then
  echo "Deploy config: $l1.json does not exist at monorepo."
  exit 1
fi

fault_game_genesis_output_root=$(jq -r .faultGameGenesisOutputRoot anchor-state/$1.json)
fault_game_genesis_block=$(jq -r .faultGameGenesisBlock anchor-state/$1.json)
fault_game_absolute_prestate=$(jq -r .faultGameAbsolutePrestate anchor-state/$1.json)

set_key_value() {
  jq --arg key "$1" --arg value "$2" '.[$key] = (try ($value | tonumber) catch $value)'
}

output_file_content=$(cat "$input_file")
# proxyAdminOwner will be the deployer of asterisc. Remove original one
output_file_content=$(echo "$output_file_content" | jq "del(.proxyAdminOwner)" )

output_file_content=$(echo "$output_file_content" | set_key_value "faultGameGenesisOutputRoot" $fault_game_genesis_output_root )
output_file_content=$(echo "$output_file_content" | set_key_value "faultGameAbsolutePrestate" $fault_game_absolute_prestate )
output_file_content=$(echo "$output_file_content" | set_key_value "faultGameGenesisBlock" $fault_game_genesis_block )

echo "$output_file_content" > "$output_file"
echo "Asterisc deployment config for $1 saved to $output_file"
