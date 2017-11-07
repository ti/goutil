package log

import (
	"os"
	"io"
	"path/filepath"
	"log"
	"sync"
)

func EmpWriter()  *empWriter {
	return &empWriter{}
}

type empWriter struct {}

func (w *empWriter) Write(b []byte) (n int, err error) {
	return
}

var out = &outPut{mu: sync.Mutex{},realOut:os.Stdout}

type outPut struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	realOut io.Writer
}

func (o *outPut) Write(p []byte) (n int, err error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.realOut.Write(p)
}

func (o *outPut) SetWrite(w io.Writer) {
	o.mu.Lock()
	defer o.mu.Unlock()
	 o.realOut = w
}


type Writer interface {
	Write(p []byte) (n int, err error)
}

func SetDefaultFileOutPut(filePath string) (err error) {
	outPutFile, err := NewFileLogOutput(filePath)
	if err != nil {
		return err
	}
	out.SetWrite(outPutFile)
	return nil
}

func SetDefaultOutput(o io.Writer) (err error) {
	out.SetWrite(o)
	return nil
}


func NewDefaultLogger(out io.Writer) *defaultLogger {
	return &defaultLogger{
		log.New(out, "", log.LstdFlags|log.Lshortfile),
	}
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
	return out
}