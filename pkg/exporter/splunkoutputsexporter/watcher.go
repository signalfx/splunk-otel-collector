// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkoutputsexporter

import (
	"github.com/fsnotify/fsnotify"
)

// fileWatcher abstracts the subset of *fsnotify.Watcher used by the exporter.
// It lets tests drive reconcile logic with a fake instead of a real OS watcher,
// which avoids spinning fsnotify's background I/O goroutine during unit tests.
type fileWatcher interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// fsnotifyWatcher adapts *fsnotify.Watcher to fileWatcher. The Events and
// Errors channels are exposed as accessor methods because fsnotify exposes
// them as struct fields.
type fsnotifyWatcher struct {
	w *fsnotify.Watcher
}

func newFSNotifyWatcher() (*fsnotifyWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsnotifyWatcher{w: w}, nil
}

func (f *fsnotifyWatcher) Add(name string) error         { return f.w.Add(name) }
func (f *fsnotifyWatcher) Close() error                  { return f.w.Close() }
func (f *fsnotifyWatcher) Events() <-chan fsnotify.Event { return f.w.Events }
func (f *fsnotifyWatcher) Errors() <-chan error          { return f.w.Errors }
