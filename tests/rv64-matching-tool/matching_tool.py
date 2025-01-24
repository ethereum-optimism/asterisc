import sys
from elftools.elf.elffile import ELFFile
import re
import json

# Lookback 5 instructions before the ECALL
SYSCALL_INSTRUCTIONS_LOOKBACK = 5

def extract_text_section_instructions(elf_path):
    """
    Extract and print executable instructions from the .text section of a RISC-V ELF binary.
    
    Args:
    - elf_path (str): Path to the ELF binary file.
    
    Returns:
    - List of hexadecimal instructions from the .text section.
    """
    try:
        with open(elf_path, 'rb') as f:
            elffile = ELFFile(f)

            # Check if the ELF is for RISC-V architecture (EM_RISCV = 243)
            if elffile['e_machine'] != 'EM_RISCV':
                print(f"Error: ELF is not for RISC-V (detected: {elffile['e_machine']})")
                exit(1)

            # Get the .text section
            text_section = elffile.get_section_by_name('.text')
            if text_section is None:
                print(f"Error: Could not find the .text section in {elf_path}")
                exit(1)

            # Extract the raw bytes from the .text section
            text_data = text_section.data()

            # Divide the text section data into 32-bit (4-byte) RISC-V instructions
            instructions = []
            for i in range(0, len(text_data), 4):
                instruction_bytes = text_data[i:i + 4]
                if len(instruction_bytes) < 4:
                    break  # If the remaining bytes are less than 4, stop
                instruction = int.from_bytes(instruction_bytes, byteorder='little')
                instructions.append(instruction)

            return instructions

    except FileNotFoundError:
        print(f"Error: File '{elf_path}' not found.")
        exit(1)
    except Exception as e:
        print(f"Error: Unable to read the ELF file. Reason: {e}")
        exit(1)
    
def parse_rd(instr):
    return (instr >> 7) & 0x1F

def parse_imm_i(instr):
    return (instr >> 20) & 0xFFF

def parse_imm_u(instr):
    return instr & 0xFFFFF000

def parse_rs1(instr):
    return (instr >> 15) & 0x1F

def parse_funct3(instr):
    return (instr >> 12) & 0x7

def parse_funct7(instr):
    return (instr >> 25)

def parse_funct12(instr):
    return (instr >> 20) & 0xFFF

def parse_opcode(instr):
    return instr & 0x7F

def instruction_name(instruction, supported):
    opcode = parse_opcode(instruction)
    funct3 = parse_funct3(instruction)
    funct7 = parse_funct7(instruction)
    funct12 = parse_funct12(instruction)
    
    opcode_hex = f"{opcode:02X}"
    funct3_hex = f"{funct3:02X}"
    funct7_hex = f"{funct7:02X}"
    funct12_hex = f"{funct12:04X}"
    
    for opcode_entry in supported['opcodes']:
        if opcode_hex in opcode_entry:
            opcode_data = opcode_entry[opcode_hex]
            
            # Check if it's a direct instruction like LUI, JAL, etc.
            if isinstance(opcode_data, str):
                return opcode_data
            
            # Check for funct3-based instructions
            if 'funct3' in opcode_data:
                for funct3_entry in opcode_data['funct3']:
                    if funct3_hex in funct3_entry:
                        funct3_data = funct3_entry[funct3_hex]
                        
                        # Check for funct12 (for ECALL, EBREAK, etc.)
                        if 'funct12' in funct3_data:
                            for funct12_entry in funct3_data['funct12']:
                                if funct12_hex in funct12_entry:
                                    funct12_data = funct12_entry[funct12_hex]
                                    return funct12_data
                                    
                        return funct3_data

            # Check for funct7-based instructions
            if 'funct7' in opcode_data:
                for funct7_entry in opcode_data['funct7']:
                    if funct7_hex in funct7_entry:
                        funct7_data = funct7_entry[funct7_hex]
                        if 'funct3' in funct7_data:
                            for funct3_entry in funct7_data['funct3']:
                                if funct3_hex in funct3_entry:
                                    return funct3_entry[funct3_hex]
                    elif 'default' in funct7_entry:
                        funct7_data = funct7_entry['default']
                        if 'funct3' in funct7_data:
                            for funct3_entry in funct7_data['funct3']:
                                if funct3_hex in funct3_entry:
                                    return funct3_entry[funct3_hex]

    return "UNKNOWN"

def parse_instructions(instructions, json_path):
    last_bytes = {}
    unknown_syscalls = {}
    unknown_instructions = {}
    supported, syscall_map = dict_from_json(json_path)

    u32max = (2**32)-1
    for index, instruction in enumerate(instructions):
        if instruction < u32max:
            ins_name = instruction_name(instruction, supported)
            if ins_name == "ECALL":
                ins_name = parse_syscall(instructions, index, syscall_map)
                if "UNKNOWN" in ins_name:
                    unknown_syscalls[ins_name] = unknown_syscalls.get(ins_name,  0) +1
            if ins_name == "UNKNOWN":
                unknown_instructions[instruction] = unknown_instructions.get(instruction, 0) + 1
            last_bytes[ins_name] = last_bytes.get(ins_name, 0) + 1
        else:
            print(f"Error: Unexpected instruction: {instruction}.")
            exit(1)
    return last_bytes, unknown_instructions, unknown_syscalls

def find_a7_value(instructions, index):
    # parse the 5 previous instructions, looking for A7 value
    for i in range(max(0,index-SYSCALL_INSTRUCTIONS_LOOKBACK), index):
        instr = instructions[i]
        rd = parse_rd(instr)
        if rd == 17:  # a7 = x17
            opcode = parse_opcode(instr)
            if opcode == 0x13:  # ADDI
                imm = parse_imm_i(instr)
                return imm
            elif opcode == 0x37:  # LUI
                imm = parse_imm_u(instr) >> 12
                return imm
            elif opcode == 0x13 and parse_rs1(instr) == 0:  # LI (ADDI x17, x0, imm)
                imm = parse_imm_i(instr)
                return imm
    return None

def parse_syscall(instructions, index, syscall_map):
    a7 = find_a7_value(instructions, index)
    if a7 == None:
        return "UNKNOWN_SYSCALL (a7 = UNKNOWN)"
    syscall_name = syscall_map.get(f"{a7:02X}")
    if syscall_map.get(f"{a7:02X}") is None:
        return f"UNKNOWN_SYSCALL (a7 = 0x{a7:X})"
    return f"ECALL.{syscall_name}"
    

def dict_from_json(json_path):
    try:
        with open(json_path, 'r') as f:
            data = json.load(f)
            syscalls = {list(s.keys())[0]: list(s.values())[0] for s in data.get('syscalls', [])}
            return data, syscalls
    except Exception as e:
        print(f"Error: Unable to read the JSON file. Reason: {e}")
        exit(1)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 parse_riscv_elf.py <path_to_elf_file> <path_to_json_file>")
        sys.exit(1)
    
    elf_path = sys.argv[1]
    json_path = sys.argv[2]
    instructions = extract_text_section_instructions(elf_path)
    
    instruction_counts, unknown_instr, unknown_syscalls = parse_instructions(instructions, json_path)
    
    # SYSCALL results
    for key in unknown_syscalls.keys():
        print(f"There were {unknown_syscalls[key]} {key}.")

    if instruction_counts.get("UNKNOWN", 0) != 0:
        nb_unknown = instruction_counts["UNKNOWN"]
        print(f"There were {nb_unknown} unknown instructions.\n")
        for instru, count in sorted(unknown_instr.items()):
            print(f"Unknown instruction: {instru:08X}: {count} times")
        exit(1)
    else:
        print("All instructions known.")
