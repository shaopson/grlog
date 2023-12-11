package grlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
)

type Level uint8

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type Logger struct {
	level     Level
	logger    *log.Logger
	isDiscard int32
}

func New(outer io.Writer, level Level, flags int) *Logger {
	if level > LevelDebug {
		level = LevelDebug
	}
	logger := &Logger{
		level:  level,
		logger: log.New(outer, "", flags),
	}
	return logger
}

func (self *Logger) Flags() int {
	return self.logger.Flags()
}

func (self *Logger) SetFlags(flags int) {
	self.logger.SetFlags(flags)
}

func (self *Logger) SetOutput(w io.Writer) {
	isDiscard := int32(0)
	if w == io.Discard {
		isDiscard = 1
	}
	atomic.StoreInt32(&self.isDiscard, isDiscard)
	self.logger.SetOutput(w)
}

func (self *Logger) Error(format string, v ...any) {
	if atomic.LoadInt32(&self.isDiscard) != 0 {
		return
	}
	self.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

func (self *Logger) Warn(format string, v ...any) {
	if self.level < LevelWarn || atomic.LoadInt32(&self.isDiscard) != 0 {
		return
	}
	self.logger.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func (self *Logger) Info(format string, v ...any) {
	if self.level < LevelInfo || atomic.LoadInt32(&self.isDiscard) != 0 {
		return
	}
	self.logger.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func (self *Logger) Debug(format string, v ...any) {
	if self.level < LevelDebug || atomic.LoadInt32(&self.isDiscard) != 0 {
		return
	}
	self.logger.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
}

func (self *Logger) Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	self.logger.Output(2, "[ERROR] "+s)
	panic(s)
}

func (self *Logger) Fatalf(format string, v ...any) {
	self.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
	os.Exit(1)
}
