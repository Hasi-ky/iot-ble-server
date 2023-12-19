package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	"path/filepath"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type fileConfig struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
	Level      logrus.Level
	Formatter  logrus.Formatter
}

type fileHook struct {
	config fileConfig
	writer io.Writer
}

// NewFileLogger create a new file logger
func NewFileLogger(c Config, fields ...string) (Logger, error) {
	logLevel, err := logrus.ParseLevel(c.Level)
	if err != nil {
		log.Printf("failed to parse log level (%s), use default level (info)\n", c.Level)
		logLevel = logrus.InfoLevel
	}

	var fileHook logrus.Hook
	if c.Path != "" {
		err = os.MkdirAll(filepath.Dir(c.Path), 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create log directory: %s", err.Error())
		} else {
			fileHook, err = newFileHook(fileConfig{
				Filename:   c.Path,
				Formatter:  newFormatter(c.Format),
				Level:      logLevel,
				MaxAge:     c.Age.Max,  //days
				MaxSize:    c.Size.Max, // megabytes
				MaxBackups: c.Backup.Max,
				Compress:   true,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create log file hook: %s", err.Error())
			}
		}
	}

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logLevel
	entry.Logger.Level = logLevel
	entry.Logger.Formatter = newFormatter(c.Format)
	if fileHook != nil {
		entry.Logger.Hooks.Add(fileHook)
	}
	logrusFields := logrus.Fields{}
	for index := 0; index < len(fields)-1; index = index + 2 {
		logrusFields[fields[index]] = fields[index+1]
	}
	return &logger{entry.WithFields(logrusFields)}, nil
}

func newFileHook(config fileConfig) (logrus.Hook, error) {
	hook := fileHook{
		config: config,
	}

	var zeroLevel logrus.Level
	if hook.config.Level == zeroLevel {
		hook.config.Level = logrus.InfoLevel
	}
	var zeroFormatter logrus.Formatter
	if hook.config.Formatter == zeroFormatter {
		hook.config.Formatter = new(logrus.TextFormatter)
	}

	hook.writer = &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		LocalTime:  config.LocalTime,
		Compress:   config.Compress,
	}

	return &hook, nil
}

// Levels Levels
func (hook *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire Fire
func (hook *fileHook) Fire(entry *logrus.Entry) (err error) {
	if hook.config.Level < entry.Level {
		return nil
	}
	b, err := hook.config.Formatter.Format(entry)
	if err != nil {
		return err
	}
	hook.writer.Write(b)
	return nil
}
