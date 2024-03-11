package main

import "fmt"

// checkErr checks if err is not nil, and throws if so.
// Shorthand to ease go's god awful error handling
// from https://github.com/ethereum-optimism/optimism/blob/5bd1f4633195411b93d292f7b34e2565da7b773e/packages/contracts-bedrock/scripts/differential-testing/utils.go
func checkErr(err error, failReason string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", failReason, err))
	}
}
