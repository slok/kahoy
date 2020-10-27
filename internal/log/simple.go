package log

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

// SimpleFormatter is a logrus compatible formatter that only prints the
// the message of the logger.
const SimpleFormatter = simpleFormatter(0)

type simpleFormatter int

var _ logrus.Formatter = SimpleFormatter

func (s simpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	// Set prefix in case of error or warning.
	switch entry.Level {
	case logrus.ErrorLevel, logrus.FatalLevel:
		fmt.Fprintf(b, "ERROR: ")
	case logrus.WarnLevel:
		fmt.Fprintf(b, "WARN: ")
	}

	// Set message.
	fmt.Fprintf(b, entry.Message)

	b.WriteByte('\n')
	return b.Bytes(), nil
}
