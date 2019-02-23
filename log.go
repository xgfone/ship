// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ship

import (
	"fmt"
	"io"
	"log"
)

// Logger stands for a logger.
type Logger interface {
	Writer() io.Writer
	Trace(format string, args ...interface{}) error
	Debug(foramt string, args ...interface{}) error
	Info(foramt string, args ...interface{}) error
	Warn(foramt string, args ...interface{}) error
	Error(foramt string, args ...interface{}) error
}

// NewNoLevelLogger returns a new Logger based on the std library log.Logger,
// which has no level, that's, its level is always DEBUG.
func NewNoLevelLogger(w io.Writer, flag ...int) Logger {
	_flag := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	if len(flag) > 0 {
		_flag = flag[0]
	}
	return &loggerT{logger: log.New(w, "", _flag), writer: w}
}

type loggerT struct {
	logger *log.Logger
	writer io.Writer
}

func (l *loggerT) Writer() io.Writer {
	return l.writer
}

func (l *loggerT) output(level, format string, args ...interface{}) error {
	return l.logger.Output(4, fmt.Sprintf(level+format, args...))
}

func (l *loggerT) Trace(format string, args ...interface{}) error {
	return l.output("[T] ", format, args...)
}

func (l *loggerT) Debug(format string, args ...interface{}) error {
	return l.output("[D] ", format, args...)
}

func (l *loggerT) Info(format string, args ...interface{}) error {
	return l.output("[I] ", format, args...)
}

func (l *loggerT) Warn(format string, args ...interface{}) error {
	return l.output("[W] ", format, args...)
}

func (l *loggerT) Error(format string, args ...interface{}) error {
	return l.output("[E] ", format, args...)
}
