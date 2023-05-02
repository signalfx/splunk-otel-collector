package service

import (
	"io"
	"regexp"

	"github.com/sirupsen/logrus"
)

type LogWriterCloser struct {
	log logrus.FieldLogger
}

func NewLogWriterCloser(log logrus.FieldLogger) *LogWriterCloser {
	return &LogWriterCloser{log: log}
}

func (lwc *LogWriterCloser) Write(p []byte) (n int, err error) {
	lwc.log.Info(string(Scrub(p)))
	return len(p), nil
}

func (lwc *LogWriterCloser) Close() error {
	return nil
}

type LogProvider struct {
	log logrus.FieldLogger
}

func (s *LogProvider) NewFile(p string) io.WriteCloser {
	return NewLogWriterCloser(s.log)
}

func (s *LogProvider) Flush() {
}

var scrubPassword = regexp.MustCompile(`<password>(.*)</password>`)

func Scrub(in []byte) []byte {
	return scrubPassword.ReplaceAll(in, []byte(`<password>********</password>`))
}
