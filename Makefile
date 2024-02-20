build-rvgo:
	make -C ./rvgo build
.PHONY: build-rvgo

build-rvsol:
	make -C ./rvsol build
.PHONY: build-rvsol

build-test:
	make -C ./tests/go-tests all
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

fuzz build:
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallExit ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallNoop ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintRead ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageRead ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintWrite ./rvgo/test
	go test -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageWrite ./rvgo/test

fuzz-mac:
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallExit ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateSyscallNoop ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintRead ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageRead ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStateHintWrite ./rvgo/test
	go test -ldflags=-extldflags=-Wl,-ld_classic -run NOTAREALTEST -v -fuzztime 10s -fuzz=FuzzStatePreimageWrite ./rvgo/test

.PHONY: \
  fuzz \
  fuzz-mac
