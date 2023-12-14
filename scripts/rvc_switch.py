import argparse

parser = argparse.ArgumentParser(description="C Extension Switch Generator")
parser.add_argument('--yul', default=False, action=argparse.BooleanOptionalAction)
parser.add_argument('--go', default=False, action=argparse.BooleanOptionalAction)
parsed_args = parser.parse_args()

# Instructions from the RVC extension of RISC-V, sorted by opcode -> funct3
# Tuple format: name - type - opcode - funct3
C_INSTRS = [
    # C0
    ('C.ADDI4SPN', 'CIW', 0, 0),
    ('C.FLD', 'CL', 0, 1),
    ('C.LW', 'CL', 0, 2),
    ('C.LD', 'CL', 0, 3),
    ('Reserved', '~', 0, 4),
    ('C.FSD (Unsupported)', 'CS', 0, 5),
    ('C.SW', 'CS', 0, 6),
    ('C.SD', 'CS', 0, 7),
    # C1
    ('C.NOP, C.ADDI', 'CI', 1, 0),
    ('C.ADDIW', 'CI', 1, 1),
    ('C.LI', 'CI', 1, 2),
    ('C.ADDI16SP, C.LUI', 'CI', 1, 3),
    ('C.SRLI, S.SRLI64, C.SRAI64, C.ANDI, C.SUB, C.XOR, C.OR, C.AND, C.SUBW, C.ADDW', '?', 1, 4),
    ('C.J', 'CR', 1, 5),
    ('C.BEQZ', 'CB', 1, 6),
    ('C.BNEZ', 'CB', 1, 7),
    # C2
    ('C.SLLI64', 'CI', 2, 0),
    ('C.FLDSP (Unsupported)', 'CI', 2, 1),
    ('C.LWSP', 'CI', 2, 2),
    ('C.LDSP', 'CI', 2, 3),
    ('C.JR, C.MV, C.EBREAK, C.JALR, C.ADD', '?', 2, 4),
    ('C.FSDSP (Unsupported)', 'CSS', 2, 5),
    ('C.SWSP', 'CSS', 2, 6),
    ('C.SDSP', 'CSS', 2, 7)
]

def compute_switch_selector(op: int, funct3: int) -> int:    
    """
    Computes a unique 5 bit switch statement selector for an opcode / funct3 combination from a RVC instruction.

    Format:

    Funct3      | Op
    vvvvvvvvvvvv|vvvvvvvv
    ┌───┬───┬───┬───┬───┐
    │ 0 │ 1 │ 2 │ 3 │ 4 │
    └───┴───┴───┴───┴───┘
    """
    return ((funct3 << 2) & 0x1C) | (op & 0x3)

# Precompute switch selectors for all `C_INSTRS` and sort by the selectors.
C_INSTRS = [(instr[0], instr[1], instr[2], instr[3], compute_switch_selector(instr[2], instr[3])) for instr in C_INSTRS]
C_INSTRS.sort(key=lambda x: x[4])

def print_switch():
    """
    Prints out a Golang switch statement for all supported RVC extension instructions, sorted by the computed
    switch selectors.
    """

    if (not (parsed_args.go or parsed_args.yul)) or (parsed_args.go and parsed_args.yul):
        print('[ERROR]: Must pass `--go` *or* `--yul`')
        return

    print('switch instr', end=(' {\n' if parsed_args.go else '\n'))

    for instr in C_INSTRS:
        if parsed_args.go:
            print('    // %s [OP: C%d | Funct3: %s | Format: %s]' % (instr[0], instr[2], '{:03b}'.format(instr[3]), instr[1]))
            print('    case 0x%X:' % instr[4])
            print('        // TODO: Perform translation to 32 bit analogue.')
        else:
            print('case 0x%X { // %s [OP: C%d | Funct3: %s | Format: %s]' % (instr[4], instr[0], instr[2], '{:03b}'.format(instr[3]), instr[1]))
            print('   // TODO: Perform translation to 32 bit analogue.')
            print('}')

    # Close out the switch
    if parsed_args.go:
        print('    default:')
        print('        panic("Unknown instruction")')
        print('}')
    else:
        print('default { revert(0, 0) }')

if __name__ == '__main__':
    print_switch()
