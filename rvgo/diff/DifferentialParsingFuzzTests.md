# Running 

```bash
make fuzz-parsing
```

or 

```bash
fuzz-parsing: build
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseTypeI ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseTypeS ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseTypeB ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseTypeU ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseTypeJ ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseOpcode ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime $(fuzztime) -fuzz=FuzzParseRd ./rvgo/test
```

## How it works 

This test is relatively simple in comparison to others – it implements each Parse function, and calls slow and fast to ensure that the output from both are identical. This pattern is meant to show the recommended way to test these functions when they are publicly accessible, such that you can compare to ensure that fast and slow behave exactly as expected. The EVM implementation of these differential fuzzing tests do not use go fuzz, and use foundry ffi to call out to a `slow` executable instead. 