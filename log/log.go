package log

import (
	"fmt"
	"log"

	glog "github.com/golang/glog"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Flush()
}

var logger Logger

// 支持外部替换
func Export(l Logger)                           { logger = l }
func Debug(args ...interface{})                 { logger.Debug(args...) }
func Debugf(format string, args ...interface{}) { logger.Debugf(format, args...) }
func Info(args ...interface{})                  { logger.Info(args...) }
func Infof(format string, args ...interface{})  { logger.Infof(format, args...) }
func Warn(args ...interface{})                  { logger.Warn(args...) }
func Warnf(format string, args ...interface{})  { logger.Warnf(format, args...) }
func Error(args ...interface{})                 { logger.Error(args...) }
func Errorf(format string, args ...interface{}) { logger.Errorf(format, args...) }
func Fatal(args ...interface{})                 { logger.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { logger.Fatalf(format, args...) }
func Flush()                                    { logger.Flush() }

func init() {
	logger = new(stdLogger)
}

type glogger struct{}

func (l *glogger) Debug(args ...interface{})                 { fmt.Println(args...) }
func (l *glogger) Debugf(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) }
func (l *glogger) Info(args ...interface{})                  { glog.Info(args...) }
func (l *glogger) Infof(format string, args ...interface{})  { glog.Infof(format, args...) }
func (l *glogger) Warn(args ...interface{})                  { glog.Warning(args...) }
func (l *glogger) Warnf(format string, args ...interface{})  { glog.Warningf(format, args...) }
func (l *glogger) Error(args ...interface{})                 { glog.Error(args...) }
func (l *glogger) Errorf(format string, args ...interface{}) { glog.Errorf(format, args...) }
func (l *glogger) Fatal(args ...interface{})                 { glog.Fatal(args...) }
func (l *glogger) Fatalf(format string, args ...interface{}) { glog.Fatalf(format, args...) }
func (l *glogger) Flush()                                    { glog.Flush() }

type stdLogger struct {
}

func (l *stdLogger) Debug(args ...interface{}) { log.Println("[DEBUG] ", args) }
func (l *stdLogger) Debugf(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}
func (l *stdLogger) Info(args ...interface{})                 { log.Println("[INFO] ", args) }
func (l *stdLogger) Infof(format string, args ...interface{}) { log.Printf("[INFO] "+format, args...) }
func (l *stdLogger) Warn(args ...interface{})                 { log.Println("[WARN] ", args) }
func (l *stdLogger) Warnf(format string, args ...interface{}) { log.Printf("[WARN] "+format, args...) }
func (l *stdLogger) Error(args ...interface{})                { log.Println("[ERROR] ", args) }
func (l *stdLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}
func (l *stdLogger) Fatal(args ...interface{}) { log.Fatal(args...) }
func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
func (l *stdLogger) Flush() {}
