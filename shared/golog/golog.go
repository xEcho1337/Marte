package golog

import (
	"fmt"
	"os"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Gray   = "\033[90m"
	White  = "\033[97m"
	Yellow = "\033[33m"
)

type Logger struct {
	worker string
}

func New(worker string) *Logger {
	return &Logger{worker: worker}
}

func (l *Logger) format(color string, msg string) string {
	now := time.Now().Format("15:04:05")
	return fmt.Sprintf("%s[Marte] %s[%s] %s(%s): %s%s%s",
		Red, Gray, now, White, l.worker, color, msg, Reset)
}

func (l *Logger) Info(format string, args ...any) {
	fmt.Println(l.format(Yellow, fmt.Sprintf(format, args...)))
}

func (l *Logger) Error(msg string) {
	fmt.Println(l.format(Red, msg))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	fmt.Println(l.format(Yellow, msg))
}

func (l *Logger) Warnf(format string, args ...any) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(format string, args ...any) {
	fmt.Println(l.format(Gray, fmt.Sprintf(format, args...)))
}

func (l *Logger) Fatal(msg string) {
	fmt.Println(l.format(Red, msg))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.Fatal(fmt.Sprintf(format, args...))
}
