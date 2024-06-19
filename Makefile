MONOREPO_ROOT=./rvsol/lib/optimism

build-rvgo:
	make -C ./rvgo build
.PHONY: build-rvgo

build-rvsol:
	make -C ./rvsol build
.PHONY: build-rvsol

build-test:
	make -C ./tests/go-tests all-test
	make -C ./rvgo build-ffi
.PHONY: build-test

build: build-test build-rvsol build-rvgo
.PHONY: build

clean:
	make -C ./rvgo clean
	make -C ./rvsol clean
.PHONY: clean

test: build
	make -C ./rvgo test
	make -C ./rvsol test
	make fuzz
.PHONY: test

fuzz: build
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallExit ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallBrk ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallMmap ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallFcntl ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallOpenat ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallClockGettime ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallClone ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallGetrlimit ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallNoop ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateRead ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintRead ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageRead ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateWrite ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintWrite ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageWrite ./rvgo/test

fuzz-mac:
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallExit ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallBrk ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallMmap ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallFcntl ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallOpenat ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallClockGettime ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallClone ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallGetrlimit ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallNoop ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateRead ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintRead ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageRead ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateWrite ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintWrite ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageWrite ./rvgo/test

.PHONY: \
  fuzz \
  fuzz-mac

OP_PROGRAM_PATH ?= $(MONOREPO_ROOT)/op-program/bin-riscv/op-program-client-riscv.elf

prestate: build-rvgo op-program-riscv
	./rvgo/bin/asterisc load-elf --path $(OP_PROGRAM_PATH) --out ./rvgo/bin/prestate.json --meta ./rvgo/bin/meta.json
	./rvgo/bin/asterisc run --proof-at '=0' --stop-at '=1' --input ./rvgo/bin/prestate.json --meta ./rvgo/bin/meta.json --proof-fmt './rvgo/bin/%d.json' --output ""
	mv ./rvgo/bin/0.json ./rvgo/bin/prestate-proof.json
.PHONY: prestate

op-program-test-capture:
	./tests/op-program-test/capture.sh
.PHONY: op-program-test-capture

op-program:
	make -C $(MONOREPO_ROOT)/op-program
.PHONY: op-program

op-program-riscv:
	rm -rf $(MONOREPO_ROOT)/op-program/bin-riscv $(MONOREPO_ROOT)/op-program/bin
	make -C $(MONOREPO_ROOT)/op-program op-program-client-riscv
	# clear $(MONOREPO_ROOT)/op-program/bin to trigger `make cannon-prestate` at monorepo
	mv $(MONOREPO_ROOT)/op-program/bin $(MONOREPO_ROOT)/op-program/bin-riscv
.PHONY: op-program-riscv

devnet-allocs-monorepo:
	make -C $(MONOREPO_ROOT) devnet-allocs
.PHONY: devnet-allocs-monorepo

devnet-allocs: devnet-allocs-monorepo prestate
	./rvsol/scripts/devnet_allocs.sh
.PHONY: devnet-allocs

devnet-clean:
	rm -rf .devnet
	rm -rf packages/contracts-bedrock/deployments
	rm -rf packages/contracts-bedrock/deploy-config
.PHONY: devnet-clean

reproducible-prestate:
	@docker build --output ./bin/ --progress plain -f Dockerfile.repro .
	@echo "Absolute prestate hash:"
	@cat ./bin/prestate-proof.json | jq -r .pre
.PHONY: reproducible-prestate
