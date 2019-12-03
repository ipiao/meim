package log

import glog "github.com/golang/glog"

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

var logger Logger

// 支持外部替换
func Export(l Logger)                           { logger = l }
func Info(args ...interface{})                  { logger.Info(args...) }
func Infof(format string, args ...interface{})  { logger.Infof(format, args...) }
func Warn(args ...interface{})                  { logger.Warn(args...) }
func Warnf(format string, args ...interface{})  { logger.Warnf(format, args...) }
func Error(args ...interface{})                 { logger.Error(args...) }
func Errorf(format string, args ...interface{}) { logger.Errorf(format, args...) }
func Fatal(args ...interface{})                 { logger.Fatal(args...) }
func Fatalf(format string, args ...interface{}) { logger.Fatalf(format, args...) }

func init() { logger = new(glogger) }

type glogger struct{}

func (l *glogger) Info(args ...interface{})                  { glog.Info(args...) }
func (l *glogger) Infof(format string, args ...interface{})  { glog.Infof(format, args...) }
func (l *glogger) Warn(args ...interface{})                  { glog.Warning(args...) }
func (l *glogger) Warnf(format string, args ...interface{})  { glog.Warningf(format, args...) }
func (l *glogger) Error(args ...interface{})                 { glog.Error(args...) }
func (l *glogger) Errorf(format string, args ...interface{}) { glog.Errorf(format, args...) }
func (l *glogger) Fatal(args ...interface{})                 { glog.Fatal(args...) }
func (l *glogger) Fatalf(format string, args ...interface{}) { glog.Fatalf(format, args...) }
