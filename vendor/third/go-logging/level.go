// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"errors"
	"strings"
	"sync"
)

var ErrInvalidLogLevel = errors.New("logger: invalid log level")

// Level defines all available log levels for log messages.
type Level int

const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var levelNames = []string{
	"CRITICAL",
	"ERROR",
	"WARNING",
	"NOTICE",
	"INFO",
	"DEBUG",
}

// String returns the string representation of a logging level.
func (p Level) String() string {
	return levelNames[p]
}

// LogLevel returns the log level from a string representation.
func LogLevel(level string) (Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(name, level) {
			return Level(i), nil
		}
	}
	return ERROR, ErrInvalidLogLevel
}

type Leveled interface {
	GetLevel(string) Level
	SetLevel(Level, string)
	IsEnabledFor(Level, string) bool
	CodoonSetLevel(string) Level // set level info to debug, or vice versa
	GetLevelExt() map[string]int
	SetLevelExt(int, string)
}

// LeveledBackend is a log backend with additional knobs for setting levels on
// individual modules to different levels.
type LeveledBackend interface {
	Backend
	Leveled
}

type moduleLeveled struct {
	levels    map[string]Level
	backend   Backend
	formatter Formatter
	once      sync.Once
	mtx       sync.RWMutex
}

// AddModuleLevel wraps a log backend with knobs to have different log levels
// for different modules.
func AddModuleLevel(backend Backend) LeveledBackend {
	var leveled LeveledBackend
	var ok bool
	if leveled, ok = backend.(LeveledBackend); !ok {
		leveled = &moduleLeveled{
			levels:  make(map[string]Level),
			backend: backend,
		}
	}
	return leveled
}

// GetLevel returns the log level for the given module.
func (l *moduleLeveled) GetLevel(module string) Level {
	l.mtx.RLock()
	level, exists := l.levels[module]
	if exists == false {
		level, exists = l.levels[""]
		// no configuration exists, default to debug
		if exists == false {
			level = DEBUG
		}
	}
	l.mtx.RUnlock()
	return level
}

// SetLevel sets the log level for the given module
func (l *moduleLeveled) SetLevel(level Level, module string) {
	l.mtx.Lock()
	if module == "*" {
		for m, _ := range l.levels {
			l.levels[m] = level
		}
	} else {
		l.levels[module] = level
	}
	l.mtx.Unlock()
}

// CodoonSetLevel set level info to debug, or vice versa
func (l *moduleLeveled) CodoonSetLevel(module string) Level {
	oldLevel := l.GetLevel(module)
	if oldLevel == INFO {
		l.SetLevel(DEBUG, module)
		return DEBUG
	} else if oldLevel == DEBUG {
		l.SetLevel(INFO, module)
		return INFO
	}
	return oldLevel
}

// GetLevelExt get all levels map[module]level
// Use level as type int to avoid importing go-logging in other packages
func (l *moduleLeveled) GetLevelExt() map[string]int {
	ret := map[string]int{}
	for k, v := range l.levels {
		ret[k] = int(v)
	}
	return ret
}

func (l *moduleLeveled) SetLevelExt(levelInt int, module string) {
	level := Level(levelInt)
	l.SetLevel(level, module)
}

// IsEnabledFor will return true if logging is enabled for the given module.
func (l *moduleLeveled) IsEnabledFor(level Level, module string) bool {
	return level <= l.GetLevel(module)
}

func (l *moduleLeveled) Log(level Level, calldepth int, rec *Record) (err error) {
	if l.IsEnabledFor(level, rec.Module) {
		// TODO get rid of traces of formatter here. BackendFormatter should be used.
		rec.formatter = l.getFormatterAndCacheCurrent()
		err = l.backend.Log(level, calldepth+1, rec)
	}
	return
}

func (l *moduleLeveled) getFormatterAndCacheCurrent() Formatter {
	l.once.Do(func() {
		if l.formatter == nil {
			l.formatter = getFormatter()
		}
	})
	return l.formatter
}
