package log

import (
	"os"
	"log"
	"io"
	"path/filepath"
)

func EmpWriter()  *empWriter {
	return &empWriter{}
}

type empWriter struct {}

func (w *empWriter) Write(b []byte) (n int, err error) {
	return
}

var outPut io.Writer

func SetDefaultLoggerOutput(logPath string) (err error) {
	fileDir := filepath.Dir(logPath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fileDir, os.FileMode(0700));err != nil {
			return err
		}
	}
	outPut, err = os.OpenFile(logPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0700)
	if err != nil {
		return err
	}
	l  = &defaultLogger{log.New(outPut, "", log.LstdFlags|log.Lshortfile)}
	return nil
}

func GetDefaultLoggerOutput() io.Writer {
	if outPut != nil {
		return outPut
	}
	return os.Stdout
}