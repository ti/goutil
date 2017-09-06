package log

import (
	"os"
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

var outPut *os.File

func SetDefaultLoggerOutput(logPath string) (err error) {
	outPut, err = NewFileLogOutput(logPath)
	if err != nil {
		return err
	}
	stdLog.SetOutput(outPut)
	return nil
}

func NewFileLogOutput(logPath string) (file *os.File, err error) {
	fileDir := filepath.Dir(logPath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fileDir, os.FileMode(0700));err != nil {
			return nil, err
		}
	}
	return os.OpenFile(logPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0700)
}

func GetDefaultLoggerOutput() io.Writer {
	if outPut != nil {
		return outPut
	}
	return os.Stdout
}