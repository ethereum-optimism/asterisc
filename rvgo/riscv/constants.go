package riscv

const (
	SysExitGroup        = 94
	SysMmap             = 222
	SysRead             = 63
	SysWrite            = 64

	FdStdin         = 0
	FdStdout        = 1
	FdStderr        = 2
	FdHintRead      = 3
	FdHintWrite     = 4
	FdPreimageRead  = 5
	FdPreimageWrite = 6

	ErrUnrecognizedResource           = uint64(0xf0012)
	ErrUnknownAtomicOperation         = uint64(0xf001a70)
	ErrUnknownOpCode                  = uint64(0xf001c0de)
	ErrInvalidSyscall                 = uint64(0xf001ca11)
	ErrInvalidRegister                = uint64(0xbad4e9)
	ErrNotAlignedAddr                 = uint64(0xbad10ad0)
	ErrLoadExceeds8Bytes              = uint64(0xbad512e0)
	ErrStoreExceeds8Bytes             = uint64(0xbad512e8)
	ErrStoreExceeds32Bytes            = uint64(0xbad512e1)
	ErrUnexpectedRProofLoad           = uint64(0xbad22220)
	ErrUnexpectedRProofStoreUnaligned = uint64(0xbad22221)
	ErrUnexpectedRProofStore          = uint64(0xbad2222f)
	ErrUnknownCSRMode                 = uint64(0xbadc0de0)
	ErrBadAMOSize                     = uint64(0xbada70)
	ErrFailToReadPreimage             = uint64(0xbadf00d0)
	ErrBadMemoryProof                 = uint64(0xbadf00d1)
)
