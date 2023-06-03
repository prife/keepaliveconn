package linebreak

import (
	"bytes"
)

type LinebufWriter struct {
	buf bytes.Buffer
	Log func([]byte)
}

func (l *LinebufWriter) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	n = len(b)
	l.buf.Write(b)
	if !bytes.Contains(b, []byte{'\n'}) {
		return
	}

	lastChar := b[len(b)-1]
	lines := bytes.Split(b, []byte{'\n'})
	l.buf.Reset()
	if lastChar != '\n' {
		lastLine := lines[len(lines)-1]
		l.buf.Write(lastLine)
	}
	lines = lines[:len(lines)-1]
	for _, line := range lines {
		l.Log(line)
	}
	return
}

func (l *LinebufWriter) LastLine() []byte {
	if l.buf.Len() > 0 {
		return l.buf.Bytes()
	}
	return nil
}

func (l *LinebufWriter) HasLastLine() bool {
	return l.buf.Len() > 0
}
