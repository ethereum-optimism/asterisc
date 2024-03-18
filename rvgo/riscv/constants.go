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
)
