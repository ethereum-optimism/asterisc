package riscv

const (
	SysExit             = 93
	SysExitGroup        = 94
	SysBrk              = 214
	SysMmap             = 222
	SysRead             = 63
	SysWrite            = 64
	SysFcntl            = 25
	SysOpenat           = 56
	SysSchedGetaffinity = 123
	SysSchedYield       = 124
	SysClockGettime     = 113
	SysRtSigprocmask    = 135
	SysSigaltstack      = 132
	SysGettid           = 178
	SysRtSigaction      = 134
	SysClone            = 220
	SysGetrlimit        = 163
	SysMadvise          = 233
	SysEpollCreate1     = 20
	SysEpollCtl         = 21
	SysPipe2            = 59
	SysReadlinnkat      = 78
	SysNewfstatat       = 79
	SysNewuname         = 160
	SysMunmap           = 215
	SysGetRandom        = 278
	SysPrlimit64        = 261
	SysFutex            = 422
	SysNanosleep        = 101

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
	ErrBadAMOSize                     = uint64(0xbada70)
	ErrFailToReadPreimage             = uint64(0xbadf00d0)
	ErrBadMemoryProof                 = uint64(0xbadf00d1)
)
