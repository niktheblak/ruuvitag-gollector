package multilogger

import (
	"log"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}

type MultiLogger struct {
	Loggers []*log.Logger
}

func New(loggers ...*log.Logger) *MultiLogger {
	return &MultiLogger{
		Loggers: loggers,
	}
}

func (l *MultiLogger) Printf(format string, v ...interface{}) {
	for _, logger := range l.Loggers {
		logger.Printf(format, v...)
	}
}

func (l *MultiLogger) Print(v ...interface{}) {
	for _, logger := range l.Loggers {
		logger.Print(v...)
	}
}

func (l *MultiLogger) Println(v ...interface{}) {
	for _, logger := range l.Loggers {
		logger.Println(v...)
	}
}

func (l *MultiLogger) Fatal(v ...interface{}) {
	l.Loggers[0].Fatal(v...)
}

func (l *MultiLogger) Fatalf(format string, v ...interface{}) {
	l.Loggers[0].Fatalf(format, v...)
}

func (l *MultiLogger) Fatalln(v ...interface{}) {
	l.Loggers[0].Fatalln(v...)
}

func (l *MultiLogger) Panic(v ...interface{}) {
	l.Loggers[0].Panic(v...)
}

func (l *MultiLogger) Panicf(format string, v ...interface{}) {
	l.Loggers[0].Panicf(format, v...)
}

func (l *MultiLogger) Panicln(v ...interface{}) {
	l.Loggers[0].Panicln(v...)
}
