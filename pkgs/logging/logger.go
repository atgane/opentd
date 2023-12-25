package logging

import (
	"strings"

	"github.com/rs/zerolog"
)

func SetLevel(level string) {
	var gl zerolog.Level = zerolog.InfoLevel
	switch strings.ToLower(level) {
	case "trace":
		gl = zerolog.TraceLevel
	case "debug":
		gl = zerolog.DebugLevel
	case "warn":
		gl = zerolog.WarnLevel
	case "error":
		gl = zerolog.ErrorLevel
	case "fatal":
		gl = zerolog.FatalLevel
	case "panic":
		gl = zerolog.PanicLevel
	}
	zerolog.SetGlobalLevel(gl)
}
