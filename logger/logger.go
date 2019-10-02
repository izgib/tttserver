package logger

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func CreateDebugLogger() *zerolog.Logger {
	writer := zerolog.NewConsoleWriter()
	writer.TimeFormat = time.RFC3339
	writer.FormatCaller = func(i interface{}) string {
		if i == nil {
			return ""
		}
		value := fmt.Sprintf("%v", i)
		// далее использутся только репозитории из github и gitlab, отщивыаем всё вплоть до этих слов
		pos := strings.Index(value, "github.com")
		if pos < 0 {
			pos = strings.Index(value, "gitlab.")
			if pos >= 0 {
				value = value[pos:]
			}
		} else {
			value = value[pos:]
		}
		return "(" + value + ")"
	}
	writer.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("\033[1m%v\033[0m", i)
	}
	writer.FormatTimestamp = func(i interface{}) string {
		if i == nil {
			return ""
		}
		return fmt.Sprintf("\033[33;1m%v\033[0m", i)
	}
	writer.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("\033[35m%s\033[0m", i)
	}
	writer.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("[%s]", i)
	}
	writer.FormatErrFieldName = func(i interface{}) string {
		return fmt.Sprintf("\033[31m%s\033[0m", i)
	}
	writer.FormatErrFieldValue = func(i interface{}) string {
		return fmt.Sprintf("\033[31m[%v]\033[0m", i)
	}
	logger := zerolog.New(writer).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	return &logger
}

func CreateProductionLogger(writer io.Writer, level zerolog.Level) zerolog.Logger {
	return zerolog.New(writer).Level(level).With().Caller().Timestamp().Logger()
}
