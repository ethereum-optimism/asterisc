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
.PHONY: test
