package log

import (
	"fmt"
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
	reset  string = "\x1B[0m"
	red    string = "\x1B[91m"
	yellow string = "\x1B[33m"
	dim    string = "\x1B[2m"
)

var (
	mutex sync.Mutex

	// terminal (text color) modes
	colors = map[LogLevel]string{
		EMERGENCY: red,
		CRITICAL:  red,
		ERROR:     red,
		WARNING:   yellow,
		NOTICE:    reset,
		INFO:      reset,
		VERBOSE:   dim,
		DEBUG:     dim,
		TRACE:     dim,
	}
)

func log(out *os.File, colorcode string, level string, v []interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	args := make([]interface{}, len(v)+2, len(v)+3)
	args[0] = colorcode + time.Now().Format(ISO8601Milli)
	args[1] = level
	copy(args[2:], v)
	if colorcode != "" {
		args = append(args, reset)
	}
	fmt.Fprintln(out, args...)
}

func stacktrace(out *os.File, colorcode string, level string, v []interface{}) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	log(out, colorcode, fmt.Sprintf(level+" <%s@L%d>", filepath.Base(file), line), v)
}

type Logger struct {
	//*log.Logger
	Level                LogLevel
	StandardStream       *os.File
	StandardStreamColors map[LogLevel]string
	ErrorStream          *os.File
	ErrorStreamColors    map[LogLevel]string
}

func (instance *Logger) Setup(level LogLevel, stdout *os.File, stderr *os.File) {
	mutex.Lock()
	defer mutex.Unlock()
	instance.Level = level
	instance.StandardStream = stdout
	info, err := stdout.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
		instance.StandardStreamColors = colors
	} else {
		instance.StandardStreamColors = map[LogLevel]string{}
	}
	instance.ErrorStream = stderr
	info, err = stderr.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
		instance.ErrorStreamColors = colors
	} else {
		instance.ErrorStreamColors = map[LogLevel]string{}
	}
}

func (instance *Logger) Emergency(v ...interface{}) {
	if instance.Level >= EMERGENCY {
		log(instance.ErrorStream, instance.ErrorStreamColors[EMERGENCY], "[EMERG]  ", v)
	}
}

func (instance *Logger) Critical(v ...interface{}) {
	if instance.Level >= CRITICAL {
		log(instance.ErrorStream, instance.ErrorStreamColors[CRITICAL], "[CRIT]   ", v)
	}
}

func (instance *Logger) Error(v ...interface{}) {
	if instance.Level >= ERROR {
		log(instance.ErrorStream, instance.ErrorStreamColors[ERROR], "[ERROR]  ", v)
	}
}

func (instance *Logger) Warn(v ...interface{}) {
	if instance.Level >= WARNING {
		log(instance.ErrorStream, instance.ErrorStreamColors[WARNING], "[WARN]   ", v)
	}
}

func (instance *Logger) Notice(v ...interface{}) {
	if instance.Level >= NOTICE {
		log(instance.StandardStream, instance.StandardStreamColors[NOTICE], "[NOTICE] ", v)
	}
}

func (instance *Logger) Info(v ...interface{}) {
	if instance.Level >= INFO {
		log(instance.StandardStream, instance.StandardStreamColors[INFO], "[INFO]   ", v)
	}
}

func (instance *Logger) Verbose(v ...interface{}) {
	if instance.Level >= VERBOSE {
		log(instance.StandardStream, instance.StandardStreamColors[VERBOSE], "[VERBOSE]", v)
	}
}

func (instance *Logger) Debug(v ...interface{}) {
	instance.debug(v)
}

func (instance *Logger) debug(v []interface{}) {
	if instance.Level >= DEBUG {
		stacktrace(instance.StandardStream, instance.StandardStreamColors[DEBUG], "[DEBUG]  ", v)
	}
}

func (instance *Logger) Trace(v ...interface{}) {
	instance.trace(v)
}

func (instance *Logger) trace(v []interface{}) {
	if instance.Level >= TRACE {
		stacktrace(instance.StandardStream, instance.StandardStreamColors[TRACE], "[TRACE]  ", v)
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

func Setup(level LogLevel, stdout *os.File, stderr *os.File) {
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
