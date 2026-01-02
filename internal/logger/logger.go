package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	AppLog      *logrus.Logger
	InitLog     *logrus.Logger
	SbiLog      *logrus.Logger
	AnalyticsLog *logrus.Logger
	ContextLog  *logrus.Logger
)

func init() {
	AppLog = logrus.New()
	InitLog = logrus.New()
	SbiLog = logrus.New()
	AnalyticsLog = logrus.New()
	ContextLog = logrus.New()

	// Set default log level and format
	AppLog.SetLevel(logrus.InfoLevel)
	InitLog.SetLevel(logrus.InfoLevel)
	SbiLog.SetLevel(logrus.InfoLevel)
	AnalyticsLog.SetLevel(logrus.InfoLevel)
	ContextLog.SetLevel(logrus.InfoLevel)

	// Set formatter
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	AppLog.SetFormatter(formatter)
	InitLog.SetFormatter(formatter)
	SbiLog.SetFormatter(formatter)
	AnalyticsLog.SetFormatter(formatter)
	ContextLog.SetFormatter(formatter)

	// Set output
	AppLog.SetOutput(os.Stdout)
	InitLog.SetOutput(os.Stdout)
	SbiLog.SetOutput(os.Stdout)
	AnalyticsLog.SetOutput(os.Stdout)
	ContextLog.SetOutput(os.Stdout)
}

func SetLogLevel(level logrus.Level) {
	AppLog.SetLevel(level)
	InitLog.SetLevel(level)
	SbiLog.SetLevel(level)
	AnalyticsLog.SetLevel(level)
	ContextLog.SetLevel(level)
}

func SetLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	AppLog.SetOutput(file)
	InitLog.SetOutput(file)
	SbiLog.SetOutput(file)
	AnalyticsLog.SetOutput(file)
	ContextLog.SetOutput(file)

	return nil
}
