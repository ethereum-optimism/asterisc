# RISCV64 matching tool

This project aims to analyze a RISCV64 ELF binary to ensure that all its RISCV64 executable instructions
are supported by a virtual machine implementation.

The tool takes as input:
- a RISCV64 binary: the ELF file to analyze
- a JSON file: the definition of the opcodes in the virtual machine implementatiob

The tool parses the opcode of an instruction and then some of the subfields (such as `funct3`) based on the given JSON file.

When an instruction is found in the binary but is not in the JSON file, the number of `UNKNOWN` instruction is incremented
and the instruction is collected.

Moreover, the tool now supports detecting Linux syscalls.

Finally, the tool prints out the number of `UNKNOWN` instruction and the number of occurences for each of them.


## Limits

### Instructions parsing

This tool is an instruction parser. As RISCV64 instructions can be closed one to the other,
the tool may not be able to detect small discrepancies.

For example, based on the `supported_targets/asterisc-v1.1.2.json`, the tool is not yet able to differentiate
`FLW` and `FLD` instructions.

### Machine definition

The machine definition is defined in the JSON file. It details all the instructions that the machine is able to support.

The unsupported instructions must not be included in the machine definition.

Note that this definition requires detailing instruction fields such as opcode, funct3, ...

### ELF section

The tool only parses the `.text` section of the ELF RISCV64 binary. Potential executable instructions in other sections are not parsed.

## Syscalls detection

The tool supports detecting some syscalls. When an `ECALL` instruction is detected, the tool navigates through the **5 previous instructions** to find if the value of the `a7` register was set to an immediate.
The `a7` register holds the syscall identifier.

Why **5 previous instructions**? This is an arbitrary value, determined by reverse-engineering a RISCV-64 binary.

If no value for `a7` register is found or if the `a7` value is not recognized as supported, the tool will trigger an alert.

## Install

Clone the repository:

```bash
git clone https://github.com/zigtur/rv64-matching-tool
cd rv64-matching-tool
```

Install the local environment and its dependencies:

```bash
python3 -m venv localenv
source localenv/bin/activate
pip3 install -r requirements.txt
```

## Usage

Execute the Python script to analyze the RV64 binary:

```bash
python3 matching_tool.py ./path_to_binary ./supported_targets/your_VM_target.json
```




