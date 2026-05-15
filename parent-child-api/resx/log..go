package resx

import (
	"io"
	"os"

	"log/slog"

	"github.com/IAmMrChen/slogplus"
)

var Log *log

type log struct {
	Root *slog.Logger
	Job  *slog.Logger
}

func InitLog() {
	if Log != nil {
		return
	}

	setupLog(Conf.Env, newLogWriter())
}

func setupLog(env string, writer io.Writer) {
	switch env {
	case "local", "dev":
		slogplus.SetupDevelopmentTo(writer)
	default:
		slogplus.SetupProductionTo(writer)
	}

	defaultLog := slog.Default()
	Log = &log{
		Root: defaultLog,
		Job:  defaultLog.With("logger", "job"),
	}
}

func newLogWriter() io.Writer {
	return os.Stdout
}
