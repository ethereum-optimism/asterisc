package cmd

import (
	"fmt"
	"io"

	"log/slog"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

func Logger(w io.Writer, lvl slog.Level) log.Logger {
	return log.NewLogger(log.LogfmtHandlerWithLevel(w, lvl))
}

// LoggingWriter is a simple util to wrap a logger,
// and expose an io Writer interface,
// for the program running within the VM to write to.
type LoggingWriter struct {
	Name string
	Log  log.Logger
}

func logAsText(b string) bool {
	for _, c := range b {
		if (c < 0x20 || c >= 0x7F) && (c != '\n' && c != '\t') {
			return false
		}
	}
	return true
}

func (lw *LoggingWriter) Write(b []byte) (int, error) {
	t := string(b)
	if logAsText(t) {
		lw.Log.Info("", "text", t)
	} else {
		lw.Log.Info("", "data", hexutil.Bytes(b))
	}
	return len(b), nil
}

// HexU32 to lazy-format integer attributes for logging
type HexU32 uint32

func (v HexU32) String() string {
	return fmt.Sprintf("%08x", uint32(v))
}

func (v HexU32) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
