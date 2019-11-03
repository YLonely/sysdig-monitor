package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// L is the overall logger
var L = logrus.New()

func init() {
	L.SetOutput(os.Stdout)
	L.SetLevel(logrus.InfoLevel)
}
