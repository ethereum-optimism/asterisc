# op-e2e

The end to end tests in this repo depend on genesis state that is
created with the `bedrock-devnet` package, and asterisc is deployed
and registered to dispute game. To create this state, run the
following commands from the root of the repository:

```bash
make devnet-allocs
```

This will leave artifacts in the `.devnet` directory that will be
read into `op-e2e` at runtime. The default deploy configuration
used for starting all `op-e2e` based tests can be found in
`packages/contracts-bedrock/deploy-config/devnetL1.json`. There
are some values that are safe to change in memory in `op-e2e` at
runtime, but others cannot be changed or else it will result in
broken tests. Any changes to `devnetL1.json` should result in
rebuilding the `.devnet` artifacts before the new values will
be present in the `op-e2e` tests.

The design of running op-e2e at this repo is identical with
[monorepo's op-e2e](https://github.com/ethereum-optimism/optimism/blob/develop/op-e2e/README.md).

You may rebuild artifacts by following commanads:

```bash
make devnet-clean
make devnet-allocs
```

## Running tests

Run

```
cd op-e2e
go test -v ./faultproofs -timeout 99999s
```
