// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

// TODO remove Level stuff from the multi logger. Do one thing.

// multiLogger is a log multiplexer which can be used to utilize multiple log
// backends at once.
type multiLogger struct {
	backends []LeveledBackend
}

// MultiLogger creates a logger which contain multiple loggers.
func MultiLogger(backends ...Backend) LeveledBackend {
	var leveledBackends []LeveledBackend
	for _, backend := range backends {
		leveledBackends = append(leveledBackends, AddModuleLevel(backend))
	}
	return &multiLogger{leveledBackends}
}

// Log passes the log record to all backends.
func (b *multiLogger) Log(level Level, calldepth int, rec *Record) (err error) {
	for _, backend := range b.backends {
		if backend.IsEnabledFor(level, rec.Module) {
			// Shallow copy of the record for the formatted cache on Record and get the
			// record formatter from the backend.
			r2 := *rec
			if e := backend.Log(level, calldepth+1, &r2); e != nil {
				err = e
			}
		}
	}
	return
}

// GetLevel returns the highest level enabled by all backends.
func (b *multiLogger) GetLevel(module string) Level {
	var level Level
	for _, backend := range b.backends {
		if backendLevel := backend.GetLevel(module); backendLevel > level {
			level = backendLevel
		}
	}
	return level
}

// SetLevel propagates the same level to all backends.
func (b *multiLogger) SetLevel(level Level, module string) {
	for _, backend := range b.backends {
		backend.SetLevel(level, module)
	}
}

// CodoonSetLevel set level info to debug, or vice versa
func (b *multiLogger) CodoonSetLevel(module string) Level {
	var newLevel Level
	for _, backend := range b.backends {
		oldLevel := backend.GetLevel(module)
		if oldLevel == INFO {
			newLevel = INFO
			backend.SetLevel(DEBUG, module)
		} else if oldLevel == DEBUG {
			newLevel = DEBUG
			backend.SetLevel(INFO, module)
		}
	}
	return newLevel
}

// Do NOT call, just make multiLogger as type LeveledBackend interface
func (b *multiLogger) GetLevelExt() map[string]int {
	return map[string]int{}
}

// Do NOT call, just make multiLogger as type LeveledBackend interface
func (b *multiLogger) SetLevelExt(levelInt int, module string) {

}

// IsEnabledFor returns true if any of the backends are enabled for it.
func (b *multiLogger) IsEnabledFor(level Level, module string) bool {
	for _, backend := range b.backends {
		if backend.IsEnabledFor(level, module) {
			return true
		}
	}
	return false
}
