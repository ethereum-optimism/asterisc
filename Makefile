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

op-program-riscv:
	rm -rf $(MONOREPO_ROOT)/op-program/bin-riscv $(MONOREPO_ROOT)/op-program/bin
	make -C $(MONOREPO_ROOT)/op-program op-program-client-riscv
	# clear $(MONOREPO_ROOT)/op-program/bin to trigger `make cannon-prestate` at monorepo
	mv $(MONOREPO_ROOT)/op-program/bin $(MONOREPO_ROOT)/op-program/bin-riscv
.PHONY: op-program

devnet-allocs-monorepo:
	make -C $(MONOREPO_ROOT) devnet-allocs
.PHONY: devnet-allocs-monorepo

devnet-allocs: devnet-allocs-monorepo
	cp -r $(MONOREPO_ROOT)/.devnet .devnet
	mkdir -p packages/contracts-bedrock
	cp -r $(MONOREPO_ROOT)/packages/contracts-bedrock/deploy-config packages/contracts-bedrock
	mkdir -p packages/contracts-bedrock/deployments/devnetL1
	cp -r $(MONOREPO_ROOT)/packages/contracts-bedrock/deployments/devnetL1 packages/contracts-bedrock/deployments
	# Patch L1 Allocs
	jq .accounts .devnet/allocs-l1.json > /tmp/allocs-l1-patched.json
	# Generate L1 Allocs including asterisc
	# copy everything locally due to foundry permission issues
	cp ./rvgo/bin/prestate-proof.json ./rvsol/prestate-proof.json
	cp -r packages/contracts-bedrock/deployments/devnetL1 ./rvsol/devnetL1
	cp packages/contracts-bedrock/deploy-config/devnetL1.json ./rvsol/devnetL1.json
	cp /tmp/allocs-l1-patched.json ./rvsol/allocs-l1-patched.json
	cd ./rvsol && ASTERISC_PRESTATE=./prestate-proof.json \
	TARGET_L2_DEPLOYMENT_FILE=./devnetL1/.deploy \
	TARGET_L2_DEPLOY_CONFIG=./devnetL1.json \
	TARGET_L1_ALLOC=./allocs-l1-patched.json \
	DEPLOYMENT_OUTFILE=./deployments/devnetL1/.deploy \
	STATE_DUMP_PATH=./allocs-l1-asterisc.json \
	./scripts/create_poststate_after_deployment.sh
	# Create address.json
	jq -s '.[0] * .[1]' ./rvsol/devnetL1/.deploy ./rvsol/deployments/devnetL1/.deploy | tee .devnet/addresses.json
	# Patch L1 Allocs: we need json as the form {"accounts": ... } for op-e2e
	jq '{accounts: .}' ./rvsol/allocs-l1-asterisc.json > .devnet/allocs-l1.json
	# Patch .deploy
	cp .devnet/addresses.json packages/contracts-bedrock/deployments/devnetL1/.deploy
	# Remove tmps
	cd rvsol && rm -rf prestate-proof.json devnetL1 devnetL1.json allocs-l1-patched.json deployments ./allocs-l1-asterisc.json 
.PHONY: devnet-allocs

devnet-clean:
	rm -rf .devnet
	rm -rf packages/contracts-bedrock/deployments
	rm -rf packages/contracts-bedrock/deploy-config
.PHONY: devnet-clean
