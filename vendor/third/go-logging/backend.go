// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// defaultBackend is the backend used for all logging calls.
var defaultBackend LeveledBackend

func init() {
	go handleSignals()
}

func handleSignals() {
	for {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGALRM) // kill -10
		s := <-c
		newLevel := defaultBackend.CodoonSetLevel("")
		fmt.Printf("[go-logging] signal received:%s, log level set to [%s]\n", s.String(), levelNames[newLevel])
	}
}

// Backend is the interface which a log backend need to implement to be able to
// be used as a logging backend.
type Backend interface {
	Log(Level, int, *Record) error
}

// Set backend replaces the backend currently set with the given new logging
// backend.
func SetBackend(backends ...Backend) LeveledBackend {
	var backend Backend
	if len(backends) == 1 {
		backend = backends[0]
	} else {
		backend = MultiLogger(backends...)
	}

	defaultBackend = AddModuleLevel(backend)
	return defaultBackend
}

// SetLevel sets the logging level for the specified module. The module
// corresponds to the string specified in GetLogger.
func SetLevel(level Level, module string) {
	defaultBackend.SetLevel(level, module)
}

// GetLevel returns the logging level for the specified module.
func GetLevel(module string) Level {
	return defaultBackend.GetLevel(module)
}
