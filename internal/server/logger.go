package server

import (
	"github.com/sirupsen/logrus"
	"io"
	"log"
)

const (
	logrusType = "logrus"
)

func newLogger(logType string, infoWriter, errWriter io.Writer) SBucketLogger {
	switch logType {
	case logrusType:
		return newLogrusLogger(infoWriter, errWriter)
	default:
		return newDefaultLogger(infoWriter, errWriter)
	}
}

type defaultLogger struct {
	errLogger  *log.Logger
	infoLogger *log.Logger
}

func newDefaultLogger(infoWriter, errWriter io.Writer) SBucketLogger {
	return &defaultLogger{
		infoLogger: log.New(infoWriter, "", log.Ldate|log.Ltime),
		errLogger:  log.New(errWriter, "", log.Ldate|log.Ltime),
	}
}

func (l defaultLogger) Error(v interface{}) {
	l.errLogger.Println(v)
}

func (l defaultLogger) Info(v interface{}) {
	l.infoLogger.Println(v)
}

func (l defaultLogger) Debug(v interface{}) {
	l.infoLogger.Println(v)
}

func (l defaultLogger) Errorf(m string, params ...interface{}) {
	l.errLogger.Printf(m, params...)
}

func (l defaultLogger) Infof(m string, params ...interface{}) {
	l.infoLogger.Printf(m, params...)
}

func (l defaultLogger) Debugf(m string, params ...interface{}) {
	l.infoLogger.Printf(m, params...)
}

type logrusLogger struct {
	errLogger  *logrus.Logger
	infoLogger *logrus.Logger
}

func newLogrusLogger(infoWriter, errWriter io.Writer) SBucketLogger {
	loggerInfo := logrus.New()
	loggerInfo.Level = logrus.InfoLevel
	loggerInfo.SetFormatter(&logrus.JSONFormatter{})
	loggerInfo.SetOutput(infoWriter)

	loggerError := logrus.New()
	loggerError.Level = logrus.ErrorLevel
	loggerError.SetFormatter(&logrus.JSONFormatter{})
	loggerError.SetOutput(errWriter)

	return &logrusLogger{infoLogger: loggerInfo, errLogger: loggerError}
}

func (l logrusLogger) Error(v interface{}) {
	l.errLogger.Error(v)
}

func (l logrusLogger) Info(v interface{}) {
	l.infoLogger.Info(v)
}

func (l logrusLogger) Debug(v interface{}) {
	l.infoLogger.Debug(v)
}

func (l logrusLogger) Errorf(m string, params ...interface{}) {
	l.errLogger.Errorf(m, params...)
}

func (l logrusLogger) Infof(m string, params ...interface{}) {
	l.infoLogger.Infof(m, params...)
}

func (l logrusLogger) Debugf(m string, params ...interface{}) {
	l.infoLogger.Debugf(m, params...)
}
