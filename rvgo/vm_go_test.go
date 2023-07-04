package fast

import (
	"github.com/protolambda/asterisc/rvgo/fast"
)

type testOracle struct {
	hint        func(v []byte) error
	getPreimage func(k [32]byte) ([]byte, error)
}

func (t *testOracle) Hint(v []byte) error {
	return t.hint(v)
}

func (t *testOracle) GetPreimage(k [32]byte) ([]byte, error) {
	return t.getPreimage(k)
}

var _ fast.PreimageOracle = (*testOracle)(nil)

/*
func TestSimple(t *testing.T) {
	programELF, err := elf.Open("../tests/go-tests/bin/simple")
	require.NoError(t, err)
	defer programELF.Close()

	vmState, err := fast.LoadELF(programELF)
	require.NoError(t, err, "must load test suite ELF binary")

	err = fast.PatchVM(programELF, vmState)
	require.NoError(t, err, "must patch Go runtime.gcenable")

	symbols, err := fast.Symbols(programELF)
	require.NoError(t, err)

	preImages := make(map[[32]byte][]byte)
	addPreimage := func(img []byte) {
		var dat [64]byte
		copy(dat[:32], crypto.Keccak256(img))
		dat[32] = 1
		preImages[crypto.Keccak256Hash(dat[:])] = img
	}
	addInput := func(i uint64, val []byte) {
		var dat [64]byte
		binary.BigEndian.PutUint64(dat[24:], i)
		preImages[crypto.Keccak256Hash(dat[:])] = val
	}

	addInput(0, []byte{0})                                // onchain mode, no pre-image hints fd
	addInput(1, crypto.Keccak256([]byte("hello")))        // pre-state
	addInput(2, crypto.Keccak256([]byte("world")))        // input
	addInput(3, crypto.Keccak256([]byte("hello world!"))) // claim to verify
	addPreimage([]byte("hello"))                          // pre-state pre-image
	addPreimage([]byte("world"))                          // input pre-image

	po := &testOracle{
		hint: func(v []byte) error {
			t.Logf("received hint: %x", v)
			return nil
		},
		getPreimage: func(k [32]byte) ([]byte, error) {
			t.Logf("reading pre-image: %x", k)
			if v, ok := preImages[k]; ok {
				return v, nil
			} else {
				return nil, fmt.Errorf("unknown pre-image %x", k)
			}
		},
	}
	instState := fast.NewInstrumentedState(vmState, po, os.Stdout, os.Stderr)

	var preimageKey [32]byte
	var preimagePartOffset uint64
	slowPreimageOracle := oracle.PreImageReaderFn(func(key [32]byte, offset uint64) (dat [32]byte, datlen uint8, err error) {
		preimageKey = key
		preimagePartOffset = offset
		if v, ok := preImages[key]; ok {
			if offset == uint64(len(v)) {
				return [32]byte{}, 0, nil // datlen==0 signals EOF
			}
			if offset > uint64(len(v)) {
				err = fmt.Errorf("cannot read past pre-image (%x) size: %d >= %d", key, offset, len(v))
				return
			}
			// TODO: length-prefix part
			datlen = uint8(copy(dat[:], v[offset:]))
			return
		} else {
			err = fmt.Errorf("unknown pre-image %x", key)
			return
		}
	})

	sender := common.HexToAddress("0xaaaa")
	stepCode := loadStepContractCode(t)
	oracleCode := loadPreimageOracleContractCode(t)
	vmenv := newEVMEnv(t, stepCode, oracleCode)

	var lastSym elf.Symbol
	for i := 0; i < 2000_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)

		if sym.Name != lastSym.Name {
			instr := vmState.Instr()
			fmt.Printf("i: %4d  pc: 0x%x  instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, instr, sym.Name, sym.Size)
		}
		lastSym = sym

		if sym.Name == "runtime.throw" {
			throwArg := vmState.Registers[10]
			throwArgLen := vmState.Registers[11]
			if throwArgLen > 1000 {
				throwArgLen = 1000
			}
			x := vmState.Memory.ReadMemoryRange(throwArg, throwArgLen)
			dat, _ := io.ReadAll(x)
			if utf8.Valid(dat) {
				fmt.Printf("THROW! %q\n", string(dat))
			} else {
				fmt.Printf("THROW! %016x: %x\n", throwArg, dat)
			}
			break
		}

		pc := vmState.PC

		witness, err := instState.Step(true)
		require.NoError(t, err, "fast VM must run step")

		// TODO: if pre-image access
		// - prepare EVM oracle
		// - prepare Slow oracle

		// TODO: re-run in slow VM, diff output
		// TODO: re-run in EVM, diff output

		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		t.Fatalf("failed with exit code %d", vmState.ExitCode)
	}
}

func TestFastMinimal(t *testing.T) {
	programELF, err := elf.Open("../tests/go-tests/bin/minimal")
	require.NoError(t, err)
	defer programELF.Close()

	vmState, err := fast.LoadELF(programELF)
	require.NoError(t, err, "must load test suite ELF binary")

	err = fast.PatchVM(programELF, vmState)
	require.NoError(t, err, "must patch VM")

	symbols, err := fast.Symbols(programELF)
	require.NoError(t, err)

	vmState.PreimageOracle = func(key [32]byte) ([]byte, error) {
		return nil, fmt.Errorf("unknown key %x", key)
	}

	for i := 0; i < 2000_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)
		instr := vmState.Instr()
		fmt.Printf("i: %4d  pc: 0x%x  offset: %03x instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, vmState.PC-sym.Value, instr, sym.Name, sym.Size)
		if sym.Name == "runtime.throw" {
			throwArg := vmState.Registers[10]
			throwArgLen := vmState.Registers[11]
			if throwArgLen > 1000 {
				throwArgLen = 1000
			}
			x := vmState.Memory.ReadMemoryRange(throwArg, throwArgLen)
			dat, _ := io.ReadAll(x)
			if utf8.Valid(dat) {
				fmt.Printf("THROW! %q\n", string(dat))
			} else {
				fmt.Printf("THROW! %016x: %x\n", throwArg, dat)
			}
			break
		}

		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)

		}
		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		t.Fatalf("failed with exit code %d", vmState.ExitCode)
	}
}

func TestFastMinimalEVM(t *testing.T) {
	programELF, err := elf.Open("../tests/go-tests/bin/minimal")
	require.NoError(t, err)
	defer programELF.Close()

	vmState, err := fast.LoadELF(programELF)
	require.NoError(t, err, "must load test suite ELF binary")

	err = fast.PatchVM(programELF, vmState)
	require.NoError(t, err, "must patch VM")

	symbols, err := fast.Symbols(programELF)
	require.NoError(t, err)

	vmState.PreimageOracle = func(key [32]byte) ([]byte, error) {
		return nil, fmt.Errorf("unknown key %x", key)
	}

	sender := common.HexToAddress("0xaaaa")
	stepCode := loadStepContractCode(t)
	vmenv := newEVMEnv(t, stepCode, []byte{1}) // no pre-image oracle

	so := oracle.NewStateOracle()
	pre := vmState.Merkleize(so)

	for i := 0; i < 2000_000; i++ {
		sym := symbols.FindSymbol(vmState.PC)
		instr := vmState.Instr()
		fmt.Printf("i: %4d  pc: 0x%x  offset: %03x instr: %08x  symbol name: %s size: %d\n", i, vmState.PC, vmState.PC-sym.Value, instr, sym.Name, sym.Size)
		if sym.Name == "runtime.throw" {
			throwArg := vmState.Registers[10]
			throwArgLen := vmState.Registers[11]
			if throwArgLen > 1000 {
				throwArgLen = 1000
			}
			x := vmState.Memory.ReadMemoryRange(throwArg, throwArgLen)
			dat, _ := io.ReadAll(x)
			if utf8.Valid(dat) {
				fmt.Printf("THROW! %q\n", string(dat))
			} else {
				fmt.Printf("THROW! %016x: %x\n", throwArg, dat)
			}
			break
		}

		if err := fast.Step(vmState, os.Stdout, os.Stderr); err != nil {
			t.Fatalf("VM err at step %d, PC %d: %v", i, vmState.PC, err)
		}

		so.BuildAccessList(true)
		slowPost, err := slow.Step(pre, so, nil)
		if err != nil {
			t.Fatalf("slow VM err at step %d: %v", i, err)
		}

		al := so.AccessList()

		if i%10000 == 0 { // every 10k steps, check our work
			so.BuildAccessList(false)
			fastPost := vmState.Merkleize(so)
			if slowPost != fastPost {
				so.Diff(slowPost, fastPost, 1)
				t.Fatalf("slow state %x must match fast state %x", slowPost, fastPost)
			}
		}

		input := oracle.Input(al, pre)
		startingGas := uint64(30_000_000)
		ret, _, err := vmenv.Call(vm.AccountRef(sender), stepAddr, input, startingGas, big.NewInt(0))
		require.NoError(t, err, "evm must not fail (ret: %x)", ret)
		evmPost := common.BytesToHash(ret)
		if slowPost != evmPost {
			so.Diff(slowPost, evmPost, 1)
			t.Fatalf("slow state %x must match EVM state %x", slowPost, evmPost)
		}

		pre = slowPost

		if vmState.Exited {
			break
		}
	}
	require.True(t, vmState.Exited, "ran out of steps")
	if vmState.ExitCode != 0 {
		t.Fatalf("failed with exit code %d", vmState.ExitCode)
	}
}
*/
