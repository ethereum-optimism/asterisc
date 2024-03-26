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
	forge fmt
.PHONY: lint-fix

lint-check:
	forge fmt && git diff --exit-code
.PHONY: lint-check
