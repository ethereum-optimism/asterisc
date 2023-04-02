all: bin bin/simple bin/simple.dump bin/minimal bin/minimal.dump

# Use latest Go compiler, since riscv is actively changing, e.g. gentraceback refactor.
# Replace this with your custom Go root
LATEST_GOROOT=/home/proto/projects/go

bin:
	mkdir bin

bin/simple:
	cd simple && GOOS=linux GOARCH=riscv64 GOROOT=$(LATEST_GOROOT) $(LATEST_GOROOT)/bin/go build -o ../bin/simple .

bin/simple.dump: bin/simple
	riscv64-linux-gnu-objdump -D --disassemble --disassembler-options=no-aliases --wide --source -m riscv:rv64 -EL bin/simple > bin/simple.dump

bin/minimal:
	cd minimal && GOOS=linux GOARCH=riscv64 GOROOT=$(LATEST_GOROOT) $(LATEST_GOROOT)/bin/go build -o ../bin/minimal .

bin/minimal.dump: bin/minimal
	riscv64-linux-gnu-objdump -D --disassemble --disassembler-options=no-aliases --wide --source -m riscv:rv64 -EL bin/minimal > bin/minimal.dump

