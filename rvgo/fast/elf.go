package fast

import (
	"bytes"
	"debug/elf"
	"fmt"
	"io"
	"sort"
)

func LoadELF(f *elf.File) (*VMState, error) {
	out := &VMState{
		PC:        0,
		Memory:    NewMemory(),
		Registers: [32]uint64{},
		ExitCode:  0,
		Exited:    false,
		// Note: Go heap arenas in 64 bit riscv start at 0xc0_00_00_00_00
		// (c0 << 32) and range to 7f_00_00_00_00_00  (7f << 40) and specifies these with mmap hints.
		// Go imposes no address space limits on riscv64 however (based on malloc.go heapAddrBits).
		// So we grow the heap starting from this address, to not overlap with any hinted data
		Heap: 0x7f_00_00_00_00_00,
	}

	// statically prepare VM state:
	out.PC = f.Entry

	for i, prog := range f.Progs {
		//fmt.Printf("prog %d: paddr: %x range %016x - %016x  (mem %016x)  type: %s\n", i, prog.Paddr, prog.Vaddr, prog.Vaddr+prog.Memsz, prog.Memsz, prog.Type.String())
		if prog.Type == 0x70000003 {
			// RISC-V reuses the MIPS_ABIFLAGS program type to type its segment with the `.riscv.attributes` section.
			// See: https://github.com/riscv-non-isa/riscv-elf-psabi-doc/blob/master/riscv-elf.adoc#attributes
			// This section has 0 mem size because it is not loaded into memory.
			// TODO: maybe try to parse this section to see what metadata Go outputs? (if any?)
			continue
		}

		r := io.Reader(io.NewSectionReader(prog, 0, int64(prog.Filesz)))
		if prog.Filesz != prog.Memsz {
			if prog.Type == elf.PT_LOAD {
				if prog.Filesz < prog.Memsz {
					r = io.MultiReader(r, bytes.NewReader(make([]byte, prog.Memsz-prog.Filesz)))
				} else {
					return nil, fmt.Errorf("invalid PT_LOAD program segment %d, file size (%d) > mem size (%d)", i, prog.Filesz, prog.Memsz)
				}
			} else {
				return nil, fmt.Errorf("program segment %d has different file size (%d) than mem size (%d): filling for non PT_LOAD segments is not supported", i, prog.Filesz, prog.Memsz)
			}
		}

		if err := out.Memory.SetMemoryRange(prog.Vaddr, r); err != nil {
			return nil, fmt.Errorf("failed to read program segment %d: %w", i, err)
		}
	}
	return out, nil
}

func PatchVM(f *elf.File, vmState *VMState) error {
	symbols, err := f.Symbols()
	if err != nil {
		return fmt.Errorf("failed to read symbols data, cannot patch program: %w", err)
	}
	for _, s := range symbols {
		// Disable Golang GC by patching the functions that enable the GC to a no-op function.
		switch s.Name {
		case "runtime.gcenable",
			"runtime.init.5",            // patch out: init() { go forcegchelper() }
			"runtime.main.func1",        // patch out: main.func() { newm(sysmon, ....) }
			"runtime.deductSweepCredit", // uses floating point nums and interacts with gc we disabled
			"runtime.(*gcControllerState).commit",
			// We need to patch this out, we don't pass float64nan because we don't support floats
			"runtime.check":
			// RISCV patch: ret (pseudo instruction)
			// 00008067 = jalr zero, ra, 0
			// Jump And Link Register, but rd=zero so no linking, and thus only jumping to the return address.
			// (return address is in register $ra based on RISCV call convention)
			if err := vmState.Memory.SetMemoryRange(s.Value, bytes.NewReader([]byte{
				0x67, 0x80, 0x00, 0x00,
			})); err != nil {
				return fmt.Errorf("failed to patch Go runtime.gcenable: %w", err)
			}
		case "runtime.MemProfileRate":
			if err := vmState.Memory.SetMemoryRange(s.Value, bytes.NewReader(make([]byte, 8))); err != nil { // disable mem profiling, to avoid a lot of unnecessary floating point ops
				return err
			}
		}
	}

	// To no-op an instruction:
	//vmState.SetMemRange(addr, 4, bytes.NewReader([]byte{0x13, 0x00, 0x00, 0x00}))

	// now insert the initial stack

	// setup stack pointer
	sp := uint64(0x10_00_00_00_00_00_00_00)
	vmState.writeRegister(2, sp)

	// init argc, argv, aux on stack
	vmState.storeMem(sp+8*1, 8, 0x42)                // argc = 0 (argument count)
	vmState.storeMem(sp+8*2, 8, 0x35)                // argv[n] = 0 (terminating argv)
	vmState.storeMem(sp+8*3, 8, 0)                   // envp[term] = 0 (no env vars)
	vmState.storeMem(sp+8*4, 8, 6)                   // auxv[0] = _AT_PAGESZ = 6 (key)
	vmState.storeMem(sp+8*5, 8, 4096)                // auxv[1] = page size of 4 KiB (value) - (== minPhysPageSize)
	vmState.storeMem(sp+8*6, 8, 25)                  // auxv[2] = AT_RANDOM
	vmState.storeMem(sp+8*7, 8, sp+8*9)              // auxv[3] = address of 16 bytes containing random value
	vmState.storeMem(sp+8*8, 8, 0)                   // auxv[term] = 0
	vmState.storeMem(sp+8*9, 8, 0x6f727020646e6172)  // randomness 8/16
	vmState.storeMem(sp+8*10, 8, 0x6164626d616c6f74) // randomness 16/16

	// entrypoint is set as part of elf load function
	return nil
}

type SortedSymbols []elf.Symbol

// FindSymbol finds the symbol that intersects with the given addr, or nil if none exists
func (s SortedSymbols) FindSymbol(addr uint64) elf.Symbol {
	// find first symbol with higher start. Or n if no such symbol exists
	i := sort.Search(len(s), func(i int) bool {
		return s[i].Value > addr
	})
	if i == 0 {
		return elf.Symbol{Name: "!start", Value: 0}
	}
	out := &s[i-1]
	if out.Value+out.Size < addr { // addr may be pointing to a gap between symbols
		return elf.Symbol{Name: "!gap", Value: addr}
	}
	return *out
}

func Symbols(f *elf.File) (SortedSymbols, error) {
	symbols, err := f.Symbols()
	if err != nil {
		return nil, fmt.Errorf("failed to read symbols data: %w", err)
	}
	// Go compiler supposedly already sorts them,
	// but it does not do so for some Go internals like internal/bytealg.IndexByteString),
	// and not every ELF has sorted symbols.
	out := make(SortedSymbols, len(symbols))
	for i := range out {
		out[i] = symbols[i]
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Value < out[j].Value
	})
	return out, nil
}
