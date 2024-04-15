#!/bin/bash

module_name="github.com/ethereum-optimism/optimism"
submodule_path="rvsol/lib/optimism"

# Fetch the version from go.mod
go_list_version=$(go list -m -f '{{.Version}}' $module_name)
echo "Version in go.mod: $go_list_version"

# Go to the submodule directory and get the full commit hash
cd $submodule_path
submodule_version=$(git rev-parse HEAD)
echo "Submodule commit: $submodule_version"
cd ..

# Extract the commit hash from the go_list_version
# This regex assumes the hash is always after the last hyphen, which is typical for pseudo-versions
go_list_commit=$(echo $go_list_version | sed 's/.*-//')

# Check if go_list_commit is empty or not
if [ -z "$go_list_commit" ]; then
    echo "Error: Extracted commit hash is empty."
    exit 1
fi

# Ensure that the submodule version is compared properly by extracting only the necessary part
# Adjust the length of submodule_version to match the length of go_list_commit for comparison
length=${#go_list_commit}
if [ $length -eq 0 ]; then
    echo "Error: Length of the extracted commit hash is zero."
    exit 1
fi
submodule_commit_part=$(echo $submodule_version | cut -c 1-$length)

# Compare the two commit parts
if [ "$go_list_commit" == "$submodule_commit_part" ]; then
    echo "Versions match."
else
    echo "Versions do not match."
    exit 1
fi