package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

// levels
const (
	debugLevel = 0
	infoLevel  = 1
	warnLevel  = 2
	errorLevel = 3
	fatalLevel = 4
)

// "{\"level\":\"DEBUG\",\"message:\"%s\"}"
const (
	FormatJson          = "json"
	printDebugLevelJSON = "{\"level\":\"DEBUG\",\"message\":\"%s\",\"timestamp\":\"%s\",\"file\":\"%s\",\"line\":\"%d\"}"
	printInfoLevelJSON  = "{\"level\":\"INFO\",\"message\":\"%s\",\"timestamp\":\"%s\",\"file\":\"%s\",\"line\":\"%d\"}"
	printWarnLevelJSON  = "{\"level\":\"WARN\",\"message\":\"%s\",\"timestamp\":\"%s\",\"file\":\"%s\",\"line\":\"%d\"}"
	printErrorLevelJSON = "{\"level\":\"ERROR\",\"message\":\"%s\",\"timestamp\":\"%s\",\"file\":\"%s\",\"line\":\"%d\"}"
	printFatalLevelJSON = "{\"level\":\"FATAL\",\"message\":\"%s\",\"timestamp\":\"%s\",\"file\":\"%s\",\"line\":\"%d\"}"
)

const (
	printDebugLevel = "[DEBUG] "
	printInfoLevel  = "[INFO] "
	printWarnLevel  = "[WARN] "
	printErrorLevel = "[ERROR] "
	printFatalLevel = "[FATAL] "
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[97m" //92亮绿色 32深绿色，97亮白色
	colorYellow  = "\033[33m"
	colorBoldRed = "\033[1;31m" // Bold red for fatal
)

// Logger warp
type Logger struct {
	mu         sync.RWMutex
	level      int
	baseLogger *log.Logger
	BaseFile   *os.File
	format     string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

func init() {
	// new
	defaultLogger = &Logger{
		level:      debugLevel,
		baseLogger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// New a logger
func New(strLevel, format string, pathname string, flag int, outWriter ...io.Writer) error {
	// level
	var level int
	switch strings.ToUpper(strLevel) {
	case "DEBUG":
		level = debugLevel
	case "INFO":
		level = infoLevel
	case "WARN":
		level = warnLevel
	case "ERROR":
		level = errorLevel
	case "FATAL":
		level = fatalLevel
	default:
		errors.New("unknown level: " + strLevel)
	}
	once.Do(func() {
		// logger
		var baseLogger *log.Logger
		var baseFile *os.File
		writes := append([]io.Writer{os.Stdout}, outWriter...)
		// 将stderr重定向到stdout
		// os.Stderr = os.Stdout
		if pathname != "" {
			now := time.Now()
			filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
				now.Year(),
				now.Month(),
				now.Day(),
				now.Hour(),
				now.Minute(),
				now.Second())

			file, err := os.Create(path.Join(pathname, filename))
			if err != nil {
				Fatal(err)
			}
			writes = append(writes, file)
			baseFile = file
		}
		baseLogger = log.New(io.MultiWriter(writes...), "", flag)
		defaultLogger.mu.Lock()
		defer defaultLogger.mu.Unlock()
		defaultLogger.level = level
		defaultLogger.baseLogger = baseLogger
		defaultLogger.BaseFile = baseFile
		defaultLogger.format = format
	})
	return nil
}

// Close It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.BaseFile != nil {
		logger.BaseFile.Close()
	}

	logger.baseLogger = nil
	logger.BaseFile = nil
}

// read log lv
func (logger *Logger) GetOutputLv() int {
	logger.mu.RLock()
	lv := logger.level
	defer logger.mu.RUnlock()
	return lv
}

// update log level
func (logger *Logger) SetLogLevel(strLevel, format string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	// newLogLevel
	var newLogLevel int
	switch strings.ToLower(strLevel) {
	case "debug":
		newLogLevel = debugLevel
	case "info":
		newLogLevel = infoLevel
	case "warn":
		newLogLevel = warnLevel
	case "error":
		newLogLevel = errorLevel
	case "fatal":
		newLogLevel = fatalLevel
	default:
		newLogLevel = logger.level
	}
	logger.level = newLogLevel
	logger.format = format
}

func SetLogLevelDefault(lv, format string) {
	defaultLogger.SetLogLevel(lv, format)
}

func (logger *Logger) doPrintf(level int, a ...interface{}) {
	empty := ""
	if level < logger.GetOutputLv() || len(a) == 0 {
		return
	}
	format := empty
	if len(a) > 1 {
		format, _ = a[0].(string)
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}
	content := empty
	if format != empty {
		content = fmt.Sprintf(format, a[1:]...)
	} else {
		sb := strings.Builder{}
		for _, v := range a {
			sb.WriteString(fmt.Sprintf("%+v", v))
		}
		content = sb.String()
	}
	// Add color based on log level
	if logger.format == FormatJson {
		file, line := getCallerInfo()
		ts := time.Now().Format(time.RFC3339)
		switch level {
		case debugLevel:
			content = fmt.Sprintf(printDebugLevelJSON, content, ts, file, line)
		case infoLevel:
			content = fmt.Sprintf(printInfoLevelJSON, content, ts, file, line)
		case warnLevel:
			content = fmt.Sprintf(printWarnLevelJSON, content, ts, file, line)
		case errorLevel:
			content = fmt.Sprintf(printErrorLevelJSON, content, ts, file, line)
		case fatalLevel:
			content = fmt.Sprintf(printFatalLevelJSON, content, ts, file, line)
		}
	} else {
		switch level {
		case debugLevel:
			content = printDebugLevel + content
		case infoLevel:
			content = colorGreen + printInfoLevel + content + colorReset
		case warnLevel:
			content = colorYellow + printWarnLevel + content + colorReset
		case errorLevel:
			content = colorRed + printErrorLevel + content + colorReset
		case fatalLevel:
			content = colorBoldRed + printFatalLevel + content + colorReset
		}
	}
	logger.baseLogger.Output(3, content)

	if level == fatalLevel {
		os.Exit(1)
	}
}

// Debug log
func (logger *Logger) Debug(a ...interface{}) {
	logger.doPrintf(debugLevel, a...)
}

// Info log
func (logger *Logger) Info(a ...interface{}) {
	logger.doPrintf(infoLevel, a...)
}

// Warn log
func (logger *Logger) Warn(a ...interface{}) {
	logger.doPrintf(warnLevel, a...)
}

// Error log
func (logger *Logger) Error(a ...interface{}) {
	logger.doPrintf(errorLevel, a...)
}

// Fatal panic
func (logger *Logger) Fatal(a ...interface{}) {
	logger.doPrintf(fatalLevel, a...)
}

// Export It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

// Debug print
func Debug(a ...interface{}) {
	defaultLogger.doPrintf(debugLevel, a...)
}

// Info print
func Info(a ...interface{}) {
	defaultLogger.doPrintf(infoLevel, a...)
}

// Warn print
func Warn(a ...interface{}) {
	defaultLogger.doPrintf(warnLevel, a...)
}

// Error print
func Error(a ...interface{}) {
	defaultLogger.doPrintf(errorLevel, a...)
}

// Fatal print
func Fatal(a ...interface{}) {
	defaultLogger.doPrintf(fatalLevel, a...)
}

func Fatalf(a ...interface{}) {
	defaultLogger.doPrintf(fatalLevel, a...)
}

// Close default logger
func Close() {
	defaultLogger.Close()
}

// 获取调用者信息
func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0
	}
	// 缩短文件名，只保留最后两部分
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}
	return file, line
}
