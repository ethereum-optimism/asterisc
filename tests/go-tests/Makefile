all: bin bin/simple bin/simple.dump bin/minimal bin/minimal.dump

bin:
	mkdir bin

bin/simple:
	cd simple && GOOS=linux GOARCH=riscv64 GOROOT=$(LATEST_GOROOT) go build -gcflags="all=-d=softfloat" -o ../bin/simple .

bin/simple.dump: bin/simple
	riscv64-linux-gnu-objdump -D --disassemble --disassembler-options=no-aliases --wide --source -m riscv:rv64 -EL bin/simple > bin/simple.dump

bin/minimal:
	cd minimal && GOOS=linux GOARCH=riscv64 GOROOT=$(LATEST_GOROOT) go build -gcflags="all=-d=softfloat" -o ../bin/minimal .

bin/minimal.dump: bin/minimal
	riscv64-linux-gnu-objdump -D --disassemble --disassembler-options=no-aliases --wide --source -m riscv:rv64 -EL bin/minimal > bin/minimal.dump

