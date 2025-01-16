build:
	forge build
.PHONY: build

clean:
	forge clean
.PHONY: clean

test:
	forge test -vvv --ffi
.PHONY: test

lint-fix:
	forge fmt ./**/*.sol
.PHONY: lint-fix

lint-check:
	forge fmt ./**/*.sol && git diff --exit-code
.PHONY: lint-check

fuzz-ffi:
	forge test --match-path ./test/Yul64Test.sol
.PHONY: fuzz-ffi
