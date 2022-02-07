package gonebot

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// import "github.com/sirupsen/logrus"

type formatter struct{}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	// red := "\033[31m%s\033[0m"
	// green := "\033[32m%s\033[0m"
	// yellow := "\033[33m%s\033[0m"
	// blue := "\033[34m%s\033[0m"
	// magenta := "\033[35m%s\033[0m"
	// cyan := "\033[36m%s\033[0m"
	// white := "\033[37m%s\033[0m"

	return []byte(fmt.Sprintf(
		"%s [%s]: %s\n",
		entry.Time.Format("2006-01-02 15:04:05"),
		entry.Level.String(),
		entry.Message)), nil
}

func withStyle(s string, color int, background int, style int) string {
	return fmt.Sprintf("\033[%d;%d;%dm%s\033[0m", style, color, background, s)
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// 	FullTimestamp:   true,
	// 	ForceColors:     true,
	// })
	logrus.SetFormatter(&formatter{})
}
