package logger

import "fmt"

const (
	persistenceFile  = "file"
	persistenceMongo = "mongo"
	persistencePG    = "postgres"
)

func getPersistence(persistenceType string, c Config, fields ...string) (Logger, error) {
	switch persistenceType {
	case persistenceFile:
		return NewFileLogger(c, fields...)
	case persistencePG:
		return NewPGLogger(c, fields...)
	default:
		return nil, fmt.Errorf("cannot match log config")
	}
}
