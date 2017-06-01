package log


import (
	"fmt"
	"log"
	"os"
)


const (
	black colorAttribute = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

type colorAttribute int

func color(s string, c colorAttribute) string{
	return fmt.Sprintf("\u001b[%vm%s \u001b[0m",c,s)
}

type defaultLogger struct {
	*log.Logger
}


func (l *defaultLogger) Log(keyvals ...interface{}) {
	var dist string
	var s bool
	for _, v := range keyvals {
		if s {
			dist += fmt.Sprint("=",v, " ")
		} else {
			dist += fmt.Sprint(v)
		}
		s = !s
	}
	if s {
		dist += "=null"
	}
	l.Output(calldepth, header(color("LOG",blue), dist))
}

func (l *defaultLogger) Debug(v ...interface{}) {
	l.Output(calldepth, header("DEBUG", fmt.Sprint(v...)))
}

func (l *defaultLogger) Debugf(format string, v ...interface{}) {
	l.Output(calldepth, header("DEBUG", fmt.Sprintf(format, v...)))
}

func (l *defaultLogger) Info(v ...interface{}) {
	l.Output(calldepth, header(color("INFO",green), fmt.Sprint(v...)))
}

func (l *defaultLogger) Infof(format string, v ...interface{}) {
	l.Output(calldepth, header(color("INFO",green), fmt.Sprintf(format, v...)))
}

func (l *defaultLogger) Warn(v ...interface{}) {
	l.Output(calldepth, header(color("WARN",yellow), fmt.Sprint(v...)))
}

func (l *defaultLogger) Warnf(format string, v ...interface{}) {
	l.Output(calldepth, header(color("WARN",yellow), fmt.Sprintf(format, v...)))
}

func (l *defaultLogger) Error(v ...interface{}) {
	l.Output(calldepth, header(color("ERROR",red), fmt.Sprint(v...)))
}

func (l *defaultLogger) Errorf(format string, v ...interface{}) {
	l.Output(calldepth, header(color("ERROR",red), fmt.Sprintf(format, v...)))
}

func (l *defaultLogger) Fatal(v ...interface{}) {
	l.Output(calldepth, header(color("FATAL",magenta), fmt.Sprint(v...)))
	os.Exit(1)
}

func (l *defaultLogger) Fatalf(format string, v ...interface{}) {
	l.Output(calldepth, header(color("FATAL",magenta), fmt.Sprintf(format, v...)))
	os.Exit(1)
}

func (l *defaultLogger) Panic(v ...interface{}) {
	l.Logger.Panic(v)
}

func (l *defaultLogger) Panicf(format string, v ...interface{}) {
	l.Logger.Panicf(format, v...)
}

func header(lvl, msg string) string {
	return fmt.Sprintf("%s: %s", lvl, msg)
}
