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

// Logger is logger interface.
//
// Notice: The implementation maybe also has the method { Writer() io.Writer }
// to get the underlynig writer.
type Logger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewLoggerFromStdlog converts stdlib log to Logger.
//
// Notice: the returned logger has also implemented the interface
// { Writer() io.Writer }.
func NewLoggerFromStdlog(logger *log.Logger) Logger {
	return stdlog{logger}
}

// NewLoggerFromWriter returns a new logger by creating a new stdlib log.
//
// Notice: the returned logger has also implemented the interface
// { Writer() io.Writer }.
func NewLoggerFromWriter(w io.Writer, prefix string, flags ...int) Logger {
	flag := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	if len(flags) > 0 {
		flag = flags[0]
	}
	return stdlog{log.New(w, prefix, flag)}
}

type stdlog struct {
	*log.Logger
}

func (l stdlog) output(level, format string, args ...interface{}) {
	if l.Logger == nil {
		return
	} else if len(args) == 0 {
		l.Output(3, level+format)
	} else {
		l.Output(3, fmt.Sprintf(level+format, args...))
	}
}

func (l stdlog) Tracef(format string, args ...interface{}) {
	l.output("[T] ", format, args...)
}

func (l stdlog) Debugf(format string, args ...interface{}) {
	l.output("[D] ", format, args...)
}

func (l stdlog) Infof(format string, args ...interface{}) {
	l.output("[I] ", format, args...)
}

func (l stdlog) Warnf(format string, args ...interface{}) {
	l.output("[W] ", format, args...)
}

func (l stdlog) Errorf(format string, args ...interface{}) {
	l.output("[E] ", format, args...)
}
