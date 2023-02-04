package fast

import (
	"debug/elf"
	"fmt"
)

func LoadELF(f *elf.File) (*VMState, error) {
	out := &VMState{
		PC:        0,
		Memory:    make(map[uint64]*[pageSize]byte),
		Registers: [32]uint64{},
		CSR:       [4096]uint64{},
		Exit:      0,
		Exited:    false,
		Heap:      0,
	}

	// statically prepare VM state:
	out.PC = f.Entry
	// TODO support argc/argv/aux etc. by prepending (higher address) those to entrypoint in expected stack space

	for i, prog := range f.Progs {
		if prog.Type == 0x70000003 {
			// RISC-V reuses the MIPS_ABIFLAGS program type to type its segment with the `.riscv.attributes` section.
			// See: https://github.com/riscv-non-isa/riscv-elf-psabi-doc/blob/master/riscv-elf.adoc#attributes
			// This section has 0 mem size because it is not loaded into memory.
			// TODO: maybe try to parse this section to see what metadata Go outputs? (if any?)
			continue
		}
		if prog.Filesz != prog.Memsz {
			return nil, fmt.Errorf("program segment %d has different file size (%d) than mem size (%d): filling not supported", i, prog.Filesz, prog.Memsz)
		}

		// copy the segment into its assigned virtual memory, page by page
		end := prog.Vaddr + prog.Memsz
		offset := int64(0)
		for addr := prog.Vaddr; addr < end; {
			// map address to page index, and start within page
			page := out.loadOrCreatePage(addr >> pageAddrSize)
			pageStart := addr & pageAddrMask
			// copy till end of page
			pageEnd := uint64(pageSize)
			// unless we reached the end
			if (addr&^pageAddrMask)+pageSize > end {
				pageEnd = end & pageAddrMask
			}
			if _, err := prog.ReadAt(page[pageStart:pageEnd], offset); err != nil {
				return nil, fmt.Errorf("failed to read program segment %d at offset %d into memory %d: %w", i, offset, pageStart, err)
			}
			addr += pageEnd - pageStart
			offset += int64(pageEnd - pageStart)
		}
	}
	return out, nil
}
