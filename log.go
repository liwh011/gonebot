package gonebot

import (
	"fmt"
	"path"
	"regexp"

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

	result := ""
	base := fmt.Sprintf("%s [%s]", entry.Time.Format("2006-01-02 15:04:05"), entry.Level.String())

	if entry.Caller != nil {
		fullName := entry.Caller.Function
		pattern := regexp.MustCompile(`^(?:(.+)\/)?([^\/]+?)\.([^\/]+)$`)
		matches := pattern.FindStringSubmatch(fullName)
		// pkgPath := matches[1] // 上级路径
		pkgName := matches[2]
		funcName := matches[3]
		_, filename := path.Split(entry.Caller.File)

		if entry.Level == logrus.ErrorLevel {
			result = fmt.Sprintf(
				"%s %s | %s | %s - line %d - %s",
				base,
				pkgName,
				entry.Message,
				filename,
				entry.Caller.Line,
				funcName,
			)
		} else {
			result = fmt.Sprintf("%s %s | %s", base, pkgName, entry.Message)
		}
	} else {
		result = fmt.Sprintf("%s %s", base, entry.Message)
	}

	fields := entry.Data
	if len(fields) > 0 {
		for k, v := range fields {
			result += fmt.Sprintf(" | %s: %v", k, v)
		}
	}

	return []byte(result + "\n"), nil
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
	logrus.SetReportCaller(true)
}
