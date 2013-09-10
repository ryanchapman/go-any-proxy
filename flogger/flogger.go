package flogger

///////////////////////////////////////////////////////////////////////////////
// Originally Authored by https://github/zdannar. All rights given to 
// https://github/rchapman for use in all future and current projects.
///////////////////////////////////////////////////////////////////////////////

import ( 
    "log"
    "os"
    "fmt"
)

func New(logLevel, logFormat int, levelMap map[int]string) *Flogger {
    return &Flogger{ fd: os.Stdout, 
                     level: logLevel, 
                     logForm: logFormat, 
                     log : log.New(os.Stderr, "", FLOG_FORMAT),
                     lvlMap: levelMap }
}

type Flogger struct {
    lpath string
    fd *os.File
    log *log.Logger
    level int
    logForm int
    lvlMap map[int]string
}

func (f *Flogger) OpenFile(logPath string, openMode int, perms os.FileMode) (error) {

    f.lpath = logPath
    fd, err := os.OpenFile(f.lpath, openMode, perms)
    if err != nil {
        return err
    }
    f.log, f.fd = log.New(fd, "", f.logForm), fd
    return nil
}

func (f *Flogger) Close() error {
    if err := f.fd.Sync(); err != nil {
        return err
    }
    return f.fd.Close()
}

func (f *Flogger) SetLevel(level int) {
    f.level = level
}

func (f *Flogger) SetLevelMap(lmap map[int]string) {
    //TODO: Add level checking
    f.lvlMap = lmap
}

// Provides the mapping for text of the log levels
func (f *Flogger) appLevel(level int, msg string) string {
    return fmt.Sprintf(": %s : ", f.lvlMap[level]) + msg
}

// Base method that actually applies the levels and wraps log.Print, 
// log.Fatal, log.Panic
func (f *Flogger) flog(level int, args ...interface{}) {
    if level < f.level { return }
    ml := []interface{}{f.appLevel(level, "")}
    v := append(ml,args...)
    s := fmt.Sprint(v...)
    f.log.Output(4, s)
    switch {
        case level == FATAL:
            os.Exit(1)
        case level == PANIC:
            panic(s)
    }
}

// Base method that actually applies the levels and wraps log.Printf, 
// log.Fatalf, log.Panicf
func (f *Flogger) flogf(level int, msg string, args ...interface{}) {
    if level < f.level { return }

    ml := f.appLevel(level,msg)
    f.log.Output(4, fmt.Sprintf(ml, args...))

    switch {
        case level == FATAL:
            os.Exit(1)
        case level == PANIC:
            panic(msg)
    }
}

func (f *Flogger) Debugf(msg string, args ...interface{}) {
   f.flogf(DEBUG, msg, args...)
}

func (f *Flogger) Infof(msg string, args ...interface{}) {
   f.flogf(INFO, msg, args...)
}

func (f *Flogger) Warningf(msg string, args ...interface{}) {
   f.flogf(WARNING, msg, args...)
}

func (f *Flogger) Errorf(msg string, args ...interface{}) {
   f.flogf(ERROR, msg, args...)
}

func (f *Flogger) Fatalf(msg string, args ...interface{}) {
   f.flogf(FATAL, msg, args...)
}

func (f *Flogger) Panicf(msg string, args ...interface{}) {
   f.flogf(PANIC, msg, args...)
}

func (f *Flogger) Debug(args ...interface{}) {
   f.flog(DEBUG, args...)
}

func (f *Flogger) Info(args ...interface{}) {
   f.flog(INFO, args...)
}

func (f *Flogger) Warning(args ...interface{}) {
   f.flog(WARNING, args...)
}

func (f *Flogger) Error(args ...interface{}) {
   f.flog(ERROR, args...)
}

func (f *Flogger) Fatal(args ...interface{}) {
   f.flog(FATAL, args...)
}

func (f *Flogger) Panic(args ...interface{}) {
   f.flog(PANIC, args...)
}

