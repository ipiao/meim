package log

import (
	"fmt"

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

func init() { logger = new(glogger) }

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
