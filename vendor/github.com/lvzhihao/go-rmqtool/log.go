package rmqtool

import "log"

var (
	Log *Logger
)

type Logger struct {
	instance interface{}
}

func (c *Logger) Set(instance interface{}) {
	c.instance = instance
}

type LoggerInterface interface {
	Error(string, ...interface{})
	Debug(string, ...interface{})
	Fatal(string, ...interface{})
	Panic(string, ...interface{})
	Warn(string, ...interface{})
	Info(string, ...interface{})
}

func init() {
	Log = &Logger{
		instance: &DefaultLog{},
	}
}

func (c *Logger) Error(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Error(format, v...)
}

func (c *Logger) Debug(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Debug(format, v...)
}

func (c *Logger) Info(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Info(format, v...)
}

func (c *Logger) Warn(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Warn(format, v...)
}

func (c *Logger) Panic(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Panic(format, v...)
}

func (c *Logger) Fatal(format string, v ...interface{}) {
	c.instance.(LoggerInterface).Fatal(format, v...)
}

type DefaultLog struct {
	LoggerInterface
}

func (c *DefaultLog) println(t, format string, v ...interface{}) {
	list := append([]interface{}{t}, format)
	log.Println(append(list, v...)...)
}

func (c *DefaultLog) Error(format string, v ...interface{}) {
	c.println("ERROR", format, v...)
}

func (c *DefaultLog) Debug(format string, v ...interface{}) {
	c.println("DEBUG", format, v...)
}

func (c *DefaultLog) Warn(format string, v ...interface{}) {
	c.println("WARN", format, v...)
}

func (c *DefaultLog) Info(format string, v ...interface{}) {
	c.println("INFO", format, v...)
}

func (c *DefaultLog) Fatal(format string, v ...interface{}) {
	log.Fatal(append([]interface{}{format}, v...)...)
}

func (c *DefaultLog) Panic(format string, v ...interface{}) {
	log.Panic(append([]interface{}{format}, v...)...)
}
