package goutils

import (
	"log"
	"os"

	"go.uber.org/zap"
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
	Sync()
}

func DefaultLogger() *Logger {
	var logger *zap.Logger
	if os.Getenv("DEBUG") == "true" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	return ZapLogger(logger)
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

func (c *Logger) Sync() {
	c.instance.(LoggerInterface).Sync()
}

type DefaultLogInstance struct {
	LoggerInterface
}

func (c *DefaultLogInstance) println(t, format string, v ...interface{}) {
	list := append([]interface{}{t}, format)
	log.Println(append(list, v...)...)
}

func (c *DefaultLogInstance) Error(format string, v ...interface{}) {
	c.println("ERROR", format, v...)
}

func (c *DefaultLogInstance) Debug(format string, v ...interface{}) {
	c.println("DEBUG", format, v...)
}

func (c *DefaultLogInstance) Warn(format string, v ...interface{}) {
	c.println("WARN", format, v...)
}

func (c *DefaultLogInstance) Info(format string, v ...interface{}) {
	c.println("INFO", format, v...)
}

func (c *DefaultLogInstance) Fatal(format string, v ...interface{}) {
	log.Fatal(append([]interface{}{format}, v...)...)
}

func (c *DefaultLogInstance) Panic(format string, v ...interface{}) {
	log.Panic(append([]interface{}{format}, v...)...)
}

func (c *DefaultLogInstance) Sync() {
	//none
}

func ZapLogger(logger *zap.Logger) *Logger {
	return &Logger{
		instance: &ZapLoggerInstance{
			logger: logger.Sugar(),
		},
	}
}

type ZapLoggerInstance struct {
	LoggerInterface
	logger *zap.SugaredLogger
}

func (c *ZapLoggerInstance) Error(format string, v ...interface{}) {
	c.logger.Errorw(format, v...)
}

func (c *ZapLoggerInstance) Debug(format string, v ...interface{}) {
	c.logger.Debugw(format, v...)
}

func (c *ZapLoggerInstance) Warn(format string, v ...interface{}) {
	c.logger.Warnw(format, v...)
}

func (c *ZapLoggerInstance) Info(format string, v ...interface{}) {
	c.logger.Infow(format, v...)
}

func (c *ZapLoggerInstance) Fatal(format string, v ...interface{}) {
	c.logger.Fatalw(format, v...)
}

func (c *ZapLoggerInstance) Panic(format string, v ...interface{}) {
	c.logger.Panicw(format, v...)
}

func (c *ZapLoggerInstance) Sync() {
	c.logger.Sync()
}
