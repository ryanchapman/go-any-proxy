package flogger


import ( 
    "log"
    "os"
)

const (
    DEBUG int = iota
    INFO
    WARNING
    ERROR
    FATAL
    PANIC
)

var (
    FLOG_APPEND int = os.O_WRONLY|os.O_APPEND|os.O_CREATE
    FLOG_FORMAT int = log.LstdFlags|log.Lshortfile

    FLOG_LEVELS = map[int] string {
        DEBUG : "DEBUG",
        INFO : "INFO",
        WARNING : "WARNING",
        ERROR : "ERROR",
        FATAL : "FATAL",
        PANIC : "PANIC",
    }
    defLogr *Flogger = New(DEBUG, FLOG_FORMAT, FLOG_LEVELS)
)

func Close() error {
    return defLogr.Close() 
}

func OpenFile(logPath string, openMode int, perms os.FileMode) (error) {
    return defLogr.OpenFile(logPath, openMode, perms)
}

func SetLevel(level int) {
    defLogr.SetLevel(level)
}

func SetLevelMap(lmap map[int]string) {
    defLogr.SetLevelMap(lmap)
}

func Debugf(msg string, args ...interface{}) {
   defLogr.Debugf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
   defLogr.Infof(msg, args...)
}

func Warningf(msg string, args ...interface{}) {
   defLogr.Warningf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
   defLogr.Errorf(msg, args...)
}

func Fatalf(msg string, args ...interface{}) {
   defLogr.Fatalf(msg, args...)
}

func Panicf(msg string, args ...interface{}) {
   defLogr.Panicf(msg, args...)
}

func Debug(args ...interface{}) {
   defLogr.Debug(args...)
}

func Info(args ...interface{}) {
   defLogr.Info(args...)
}

func Warning(args ...interface{}) {
   defLogr.Warning(args...)
}

func Error(args ...interface{}) {
   defLogr.Error(args...)
}

func Fatal(args ...interface{}) {
   defLogr.Fatal(args...)
}

func Panic(args ...interface{}) {
   defLogr.Panic(args...)
}

func RedirectStreams() {
   defLogr.RedirectStreams()
}
