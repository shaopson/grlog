package grlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
)

var DefaultFlags = log.Ldate | log.Lshortfile | log.Ltime | log.Lmsgprefix

var std = New(os.Stderr, LevelInfo, DefaultFlags)

func Default() *Logger {
	return std
}

func Flags() int {
	return std.Flags()
}

func SetFlags(flags int) {
	std.SetFlags(flags)
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

func Error(format string, v ...any) {
	if atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

func Warn(format string, v ...any) {
	if std.level < LevelWarn || atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.logger.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func Info(format string, v ...any) {
	if std.level < LevelInfo || atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.logger.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func Debug(format string, v ...any) {
	if std.level < LevelDebug || atomic.LoadInt32(&std.isDiscard) != 0 {
		return
	}
	std.logger.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
}

func Panic(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	std.logger.Output(2, "[ERROR] "+s)
	panic(s)
}

func Fatal(format string, v ...any) {
	std.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
	os.Exit(1)
}
