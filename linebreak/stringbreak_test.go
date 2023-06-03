package linebreak

import (
	"fmt"
	"testing"
)

func TestStringBreak_Write(t *testing.T) {
	log := func(line string) {
		fmt.Println(line)
	}

	tests := []struct {
		name  string
		log   func(line string)
		args  []byte
		wantN int
	}{
		{name: "t1", log: log, args: []byte("hello, world"), wantN: len("hello, world")},
		{name: "t2", log: log, args: []byte("hello, world\n"), wantN: len("hello, world\n")},
		{name: "t3", log: log, args: []byte("hello, world\nhello"), wantN: len("hello, world\nhello")},
		{name: "t4", log: log, args: []byte("hello, world\nhello\n"), wantN: len("hello, world\nhello\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &StringBreak{
				Log: tt.log,
			}
			gotN, _ := l.Write(tt.args)
			if gotN != tt.wantN {
				t.Errorf("Write() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func Test_checkLineBreak(t *testing.T) {
	tests := []struct {
		name string
		args []byte
		want bool
	}{
		{name: "t1", args: []byte("hello, world"), want: false},
		{name: "t2", args: []byte("hello, world\nline2"), want: true},
		{name: "t3", args: []byte("hello, world\rline2\r"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkLineBreak(tt.args); got != tt.want {
				t.Errorf("checkLineBreak() = %v, want %v", got, tt.want)
			}
		})
	}
}
