package core

// Logger stands for a logger.
type Logger interface {
	Debug(foramt string, args ...interface{})
	Info(foramt string, args ...interface{})
	Warn(foramt string, args ...interface{})
	Error(foramt string, args ...interface{})
}
