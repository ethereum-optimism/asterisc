# Introduction

Asterisc has a dependency to the optimism monorepo at `rvsol/lib/optimism`. Asterisc uses various components from the monorepo, and depends on testing utilities for op-e2e, etc testing. 

Periodically, optimism monorepo must be updated in order to support for newer features or utilities. 

This guide will walk you through what must be checked when updating the monorepo dependency. 

## Update Optimism commit
- Go to `rvsol/lib/optimism` and perform `git checkout` to a specific commit you want to check out. 

- Go to `go.mod` and look for `github.com/ethereum-optimism/optimism` entry. Then, update the commit hash and commit datetime like the following:
```
github.com/ethereum-optimism/optimism v1.9.2-0.20241008153126-117c9a427168
```
- Then, run `go mod tidy`.

- Lastly, update the `tests/op-program-test/VERSION` file by replacing the content with the full commit hash. 

## Update other dependencies
- Find op-geth dependency from optimism monorepo [like this](https://github.com/ethereum-optimism/optimism/blob/a05feb362b5209ab6a200874e9d45244f12240d1/go.mod#L254).
  - If op-geth version is different from the one in Asterisc's `go.mod`, update it.
- Update `rvsol/lib/forge-std` if necessary.

## Update go version
Update the go version in `go.mod` to match the monorepo's version. 

## Update CI
- Go to `.circleci/config.yml` and look `ci_builder_image` like the following:  
```
default: us-docker.pkg.dev/oplabs-tools-artifacts/images/ci-builder:v0.53.0
```
- Compare the ci-builder's version with `rvsol/lib/optimism/.circleci/config.yml`, and update it to match the monorepo's version
- Go to `.github/workflows/ci.yaml` and update `go-version` if necessary. 


## Update Geth ABIGen
If the op-geth's version changed in the above step, follow these:
- Go to `versions.json`, and update the `"abigen"` version to the new geth version. 
- Go to `.github/workflows/ci.yaml` and navigate to `rvgo-abigen` workflow [here](https://github.com/ethereum-optimism/asterisc/blob/019d4b9f95e9ac146fe2948d85638b30ead8d5f4/.github/workflows/ci.yaml#L120-L134). Update the geth alltools version from the [Geth's Download page](https://geth.ethereum.org/downloads).
- Go to `rvgo` and run `make gen-bindings`. Commit any changed abi bindings at `rvgo/bindings`.

## Update op-e2e tests
Asterisc's op-e2e tests are mirrored from the op-e2e tests in optimism monorepo. 
Navigate to [Optimism Monorepo's op-e2e directory](https://github.com/ethereum-optimism/optimism/tree/develop/op-e2e), and move any new changes to [Asterisc's counterpart directory](https://github.com/ethereum-optimism/asterisc/tree/master/op-e2e).

## Update op-program-test data
Asterisc runs tests that verify the behavior of Asterisc with op-program. In order to run this program, proper test data is required. 
- Run `./tests/op-program-test/capture.sh`.
- This will generate a file at `tests/op-program-test/test-data.tar.gz`, commit and upload it. 