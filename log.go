package grlog

import (
	"io"
	"log"
)

type Level uint8

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type Logger struct {
	level  Level
	logger *log.Logger
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

func (self *Logger) SetFlags(flags int) {
	self.logger.SetFlags(flags)
}

func (self *Logger) SetOutput(w io.Writer) {
	self.logger.SetOutput(w)
}

func (self *Logger) Error(format string, v ...any) {
	self.logger.Printf("[ERROR] "+format, v...)
}

func (self *Logger) Warn(format string, v ...any) {
	if self.level < LevelWarn {
		return
	}
	self.logger.Printf("[WARN] "+format, v...)
}

func (self *Logger) Info(format string, v ...any) {
	if self.level < LevelInfo {
		return
	}
	self.logger.Printf("[INFO] "+format, v...)
}

func (self *Logger) Debug(format string, v ...any) {
	if self.level < LevelDebug {
		return
	}
	self.logger.Printf("[DEBUG] "+format, v...)
}

func (self *Logger) Panicf(format string, v ...any) {
	self.logger.Panicf(format, v...)
}

func (self *Logger) Fatalf(format string, v ...any) {
	self.logger.Fatalf(format, v...)
}
