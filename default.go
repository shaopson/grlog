package grlog

import (
	"fmt"
	"io"
	"os"
)

var std = New(os.Stderr, LevelInfo, Fstd)

func Default() *Logger {
	return std
}

func Flags() int {
	return std.Flags()
}

func SetFlags(flags int) {
	std.SetFlags(flags)
}

func Level() int {
	return std.Level()
}

func SetLevel(level int) {
	std.SetLevel(level)
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

func Error(format string, v ...any) {
	std.mu.Lock()
	if std.isDiscard {
		std.mu.Unlock()
		return
	}
	std.mu.Unlock()
	std.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

func Warn(format string, v ...any) {
	std.mu.Lock()
	if std.level < LevelWarn || std.isDiscard {
		std.mu.Unlock()
		return
	}
	std.mu.Unlock()
	std.logger.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func Info(format string, v ...any) {
	std.mu.Lock()
	if std.level < LevelInfo || std.isDiscard {
		std.mu.Unlock()
		return
	}
	std.mu.Unlock()
	std.logger.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func Debug(format string, v ...any) {
	std.mu.Lock()
	if std.level < LevelDebug || std.isDiscard {
		std.mu.Unlock()
		return
	}
	std.mu.Unlock()
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
