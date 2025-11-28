package logger

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type StdLogger struct {
	l     *log.Logger
	level string
}

func NewStdLogger(out io.Writer, level string) *StdLogger {
	return &StdLogger{
		l:     log.New(out, "", 0),
		level: strings.ToLower(level),
	}
}

func (s *StdLogger) format(prefix string, msg string, kv ...any) string {
	t := time.Now().Format(time.RFC3339)
	pairs := ""
	if len(kv) > 0 {
		pairs = fmt.Sprint(kv...)
	}
	return fmt.Sprintf("%s [%s] %s %s", t, prefix, msg, pairs)
}

func (s *StdLogger) Info(msg string, kv ...any) {
	s.l.Println(s.format("INFO", msg, kv...))
}

func (s *StdLogger) Error(msg string, kv ...any) {
	s.l.Println(s.format("ERROR", msg, kv...))
}
