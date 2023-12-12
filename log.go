package grlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const (
	Fdate      = 1 << iota     // the date in the local time zone: 2009/01/23
	Ftime                      // the time in the local time zone: 01:23:23
	Fmtime                     // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Flongfile                  // full file name and line number: /a/b/c/d.go:23
	Fshortfile                 // final file name element and line number: d.go:23. overrides Llongfile
	FUTC                       // if Ldate or Ltime is set, use UTC rather than the local time zone
	Fmsgprefix                 // move the "prefix" from the beginning of the line to before the message
	Fstd       = Fdate | Ftime // initial values for the standard logger
)

const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type Logger struct {
	level     int
	logger    *log.Logger
	isDiscard bool
	mu        sync.Mutex
}

func New(outer io.Writer, level int, flags int) *Logger {
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

func (self *Logger) Level() int {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.level
}

func (self *Logger) SetLevel(level int) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.level = level
}

func (self *Logger) SetOutput(w io.Writer) {
	if w == io.Discard {
		self.mu.Lock()
		self.isDiscard = true
		self.mu.Unlock()
	}
	self.logger.SetOutput(w)
}

func (self *Logger) Error(format string, v ...any) {
	self.mu.Lock()
	if self.isDiscard {
		self.mu.Unlock()
		return
	}
	self.mu.Unlock()
	self.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

func (self *Logger) Warn(format string, v ...any) {
	self.mu.Lock()
	if self.level < LevelWarn || self.isDiscard {
		self.mu.Unlock()
		return
	}
	self.mu.Unlock()
	self.logger.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func (self *Logger) Info(format string, v ...any) {
	self.mu.Lock()
	if self.level < LevelInfo || self.isDiscard {
		self.mu.Unlock()
		return
	}
	self.mu.Unlock()
	self.logger.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func (self *Logger) Debug(format string, v ...any) {
	self.mu.Lock()
	if self.level < LevelDebug || self.isDiscard {
		self.mu.Unlock()
		return
	}
	self.mu.Unlock()
	self.logger.Output(2, fmt.Sprintf("[DEBUG] "+format, v...))
}

func (self *Logger) Panic(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	self.logger.Output(2, "[ERROR] "+s)
	panic(s)
}

func (self *Logger) Fatal(format string, v ...any) {
	self.logger.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
	os.Exit(1)
}
