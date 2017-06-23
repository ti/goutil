package log

func EmpWriter()  *empWriter {
	return &empWriter{}
}

type empWriter struct {}

func (w *empWriter) Write(b []byte) (n int, err error) {
	return
}
