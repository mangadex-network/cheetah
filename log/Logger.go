package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type LogLevel int

const (
	EMERGENCY LogLevel = 0
	CRITICAL  LogLevel = 1
	ERROR     LogLevel = 2
	WARNING   LogLevel = 3
	NOTICE    LogLevel = 4
	INFO      LogLevel = 5
	VERBOSE   LogLevel = 6
	DEBUG     LogLevel = 7
	TRACE     LogLevel = 8

	// timestamp format
	ISO8601Milli = "2006-01-02T15:04:05.000"

	// terminal (text color) modes
	red    string = "\x1B[91m"
	yellow string = "\x1B[33m"
	dim    string = "\x1B[2m"
	reset  string = "\x1B[0m"
)

var mutex sync.Mutex

func log(out io.Writer, colorcode string, level string, v []interface{}) {
	args := make([]interface{}, len(v)+3)
	args[0] = colorcode + time.Now().Format(ISO8601Milli)
	args[1] = level
	copy(args[2:], v)
	args[len(args)-1] = reset
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Fprintln(out, args...)
}

func stacktrace(out io.Writer, level string, v []interface{}) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	log(out, dim, fmt.Sprintf(level+" <%s@L%d>", filepath.Base(file), line), v)
}

type Logger struct {
	//*log.Logger
	Level          LogLevel
	StandardStream io.Writer
	ErrorStream    io.Writer
}

func (instance *Logger) Setup(level LogLevel, stdout io.Writer, stderr io.Writer) {
	mutex.Lock()
	defer mutex.Unlock()
	instance.Level = level
	instance.StandardStream = stdout
	instance.ErrorStream = stderr
}

func (instance *Logger) Emergency(v ...interface{}) {
	if instance.Level >= EMERGENCY {
		log(os.Stderr, red, "[EMERG]  ", v)
	}
}

func (instance *Logger) Critical(v ...interface{}) {
	if instance.Level >= CRITICAL {
		log(os.Stderr, red, "[CRIT]   ", v)
	}
}

func (instance *Logger) Error(v ...interface{}) {
	if instance.Level >= ERROR {
		log(os.Stderr, red, "[ERROR]  ", v)
	}
}

func (instance *Logger) Warn(v ...interface{}) {
	if instance.Level >= WARNING {
		log(os.Stderr, yellow, "[WARN]   ", v)
	}
}

func (instance *Logger) Notice(v ...interface{}) {
	if instance.Level >= NOTICE {
		log(os.Stdout, reset, "[NOTICE] ", v)
	}
}

func (instance *Logger) Info(v ...interface{}) {
	if instance.Level >= INFO {
		log(os.Stdout, reset, "[INFO]   ", v)
	}
}

func (instance *Logger) Verbose(v ...interface{}) {
	if instance.Level >= VERBOSE {
		log(os.Stdout, dim, "[VERBOSE]", v)
	}
}

func (instance *Logger) Debug(v ...interface{}) {
	instance.debug(v)
}

func (instance *Logger) debug(v []interface{}) {
	if instance.Level >= DEBUG {
		stacktrace(os.Stdout, "[DEBUG]  ", v)
	}
}

func (instance *Logger) Trace(v ...interface{}) {
	instance.trace(v)
}

func (instance *Logger) trace(v []interface{}) {
	if instance.Level >= TRACE {
		stacktrace(os.Stdout, "[TRACE]  ", v)
	}
}

/*********************
*** DEFAULT LOGGER ***
*********************/

var current *Logger = &Logger{
	Level:          INFO,
	StandardStream: os.Stdout,
	ErrorStream:    os.Stderr,
}

func Setup(level LogLevel, stdout io.Writer, stderr io.Writer) {
	current.Setup(level, stdout, stderr)
}

func Emergency(v ...interface{}) {
	current.Emergency(v...)
}

func Critical(v ...interface{}) {
	current.Critical(v...)
}

func Error(v ...interface{}) {
	current.Error(v...)
}

func Warn(v ...interface{}) {
	current.Warn(v...)
}

func Notice(v ...interface{}) {
	current.Notice(v...)
}

func Info(v ...interface{}) {
	current.Info(v...)
}

func Verbose(v ...interface{}) {
	current.Verbose(v...)
}

func Debug(v ...interface{}) {
	current.debug(v)
}

func Trace(v ...interface{}) {
	current.trace(v)
}
