// This package is modified from the Go log standard package

package grlog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	FlagDate   = 1 << iota // the date in the local time zone: 2009/01/23
	FlagTime               // the time in the local time zone: 01:23:23
	FlagMtime              // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	FlagLFile              // full file name and line number: /a/b/c/d.go:23
	FlagSFile              // final file name element and line number: d.go:23. overrides Llongfile
	FlagUTC                // if Ldate or Ltime is set, use UTC rather than the local time zone
	FlagPrefix             // move the "prefix" from the beginning of the line to before the message
	FlagLevel
	FlagStd = FlagDate | FlagTime | FlagLevel // initial values for the standard logger

	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

// A Logger represents an active logging object that generates lines of
// output to an io.Writer. Each logging operation makes a single call to
// the Writer's Write method. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	prefix    string     // prefix on each line to identify the logger (but see Lmsgprefix)
	flag      int        // properties
	out       io.Writer  // destination for output
	buf       []byte     // for accumulating text to write
	isDiscard int32      // atomic boolean: whether out == io.Discard
	level     int        // log level: info, warn, error, debug
}

// New creates a new Logger
func New(out io.Writer, prefix string, flag int, level int) *Logger {
	l := &Logger{out: out, prefix: prefix, flag: flag, level: level}
	if out == io.Discard {
		l.isDiscard = 1
	}
	return l
}

// SetOutput sets the output destination for the logger.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	isDiscard := int32(0)
	if w == io.Discard {
		isDiscard = 1
	}
	atomic.StoreInt32(&l.isDiscard, isDiscard)

}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
//   - l.prefix (if it's not blank and Lmsgprefix is unset),
//   - date and/or time (if corresponding flags are provided),
//   - file and line number (if corresponding flags are provided),
//   - l.prefix (if it's not blank and Lmsgprefix is set).
func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int, level ...int) {
	if l.flag&FlagPrefix == 0 {
		*buf = append(*buf, l.prefix...)
	}
	if l.flag&(FlagDate|FlagTime|FlagMtime) != 0 {
		if l.flag&FlagUTC != 0 {
			t = t.UTC()
		}
		if l.flag&FlagDate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(FlagTime|FlagMtime) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&FlagMtime != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(FlagSFile|FlagLFile) != 0 {
		if l.flag&FlagSFile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ' ')
		//*buf = append(*buf, ": "...)
	}
	if l.flag&FlagLevel != 0 && len(level) > 0 {
		switch {
		case level[0] <= LevelError:
			*buf = append(*buf, "ERROR "...)
		case level[0] <= LevelWarn:
			*buf = append(*buf, "WARN "...)
		case level[0] <= LevelInfo:
			*buf = append(*buf, "INFO "...)
		default:
			*buf = append(*buf, "DEBUG "...)
		}
	}
	if l.flag&FlagPrefix != 0 {
		*buf = append(*buf, l.prefix...)
	}
}

func (l *Logger) Level() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *Logger) LevelString() string {
	switch {
	case l.level <= LevelError:
		return "ERROR"
	case l.level <= LevelWarn:
		return "ERROR"
	case l.level <= LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func (l *Logger) SetLevel(level int) {
	if level > LevelDebug {
		level = LevelDebug
	} else if level < LevelError {
		level = LevelError
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) Output(calldepth int, s string, level ...int) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(FlagSFile|FlagLFile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line, level...)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...))
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.Output(2, fmt.Sprint(v...))
}

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.Output(2, fmt.Sprintln(v...))
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *Logger) Fatal(v ...any) {
	l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, v ...any) {
	l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *Logger) Fatalln(v ...any) {
	l.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *Logger) Panic(v ...any) {
	s := fmt.Sprint(v...)
	l.Output(2, s)
	panic(s)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *Logger) Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	l.Output(2, s)
	panic(s)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *Logger) Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	l.Output(2, s)
	panic(s)
}

// Flags returns the output flags for the logger.
// The flag bits are Ldate, Ltime, and so on.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// SetFlags sets the output flags for the logger.
// The flag bits are Ldate, Ltime, and so on.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// Writer returns the output destination for the logger.
func (l *Logger) Writer() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}

func (l *Logger) Error(format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.mu.Lock()
	if l.level < LevelError {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	l.Output(2, fmt.Sprintf(format, v...), LevelError)
}

func (l *Logger) Warn(format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.mu.Lock()
	if l.level < LevelWarn {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	l.Output(2, fmt.Sprintf(format, v...), LevelWarn)
}

func (l *Logger) Info(format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.mu.Lock()
	if l.level < LevelInfo {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	l.Output(2, fmt.Sprintf(format, v...), LevelInfo)
}

func (l *Logger) Debug(format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.mu.Lock()
	if l.level < LevelDebug {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	l.Output(2, fmt.Sprintf(format, v...), LevelDebug)
}

func (l *Logger) Log(level int, format string, v ...any) {
	if atomic.LoadInt32(&l.isDiscard) != 0 {
		return
	}
	l.Output(2, fmt.Sprintf(format, v...), level)
}
