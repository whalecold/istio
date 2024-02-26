package envoy

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

type Watcher interface {
	Run() error
	Notify() <-chan struct{}
	Close()
}

type AggregateWatcher struct {
	watchers []Watcher
	notify   chan struct{}
}

func NewAggregateWatcher(cfg ProxyConfig) (Watcher, error) {
	w := &AggregateWatcher{
		watchers: make([]Watcher, 0, 2),
		notify:   make(chan struct{}),
	}
	fw, err := NewFileWatcher(cfg.ConfigPath, w.notify)
	if err != nil {
		return nil, err
	}
	w.watchers = append(w.watchers, fw)
	w.watchers = append(w.watchers, NewSignalWatcher(w.notify))
	return w, nil
}

func (aw *AggregateWatcher) Run() error {
	for _, w := range aw.watchers {
		if err := w.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (aw *AggregateWatcher) Notify() <-chan struct{} {
	return aw.notify
}

func (aw *AggregateWatcher) Close() {
	for _, w := range aw.watchers {
		w.Close()
	}
	close(aw.notify)
}

type FileWatcher struct {
	filePath string
	w        *fsnotify.Watcher
	event    chan struct{}
}

func NewFileWatcher(filePath string, event chan struct{}) (Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("init file watch failed: %v", err)
	}
	return &FileWatcher{
		filePath: filePath,
		w:        watcher,
		event:    event,
	}, nil
}

func (fw *FileWatcher) Run() error {
	go func() {
		for {
			select {
			case event, ok := <-fw.w.Events:
				if !ok {
					return
				}
				if strings.Contains(event.String(), "WRITE") {
					fw.event <- struct{}{}
				}
				// TODO deal with error
			case _, ok := <-fw.w.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return fw.w.Add(fw.filePath)
}

func (fw *FileWatcher) Notify() <-chan struct{} {
	return fw.event
}

func (fw *FileWatcher) Close() {
	fw.w.Close()
}

type SignalWatcher struct {
	event chan struct{}
	stop  chan struct{}
}

func NewSignalWatcher(event chan struct{}) Watcher {
	return &SignalWatcher{
		event: event,
		stop:  make(chan struct{}),
	}
}

func (sw *SignalWatcher) Run() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-c:
				sw.event <- struct{}{}
			case <-sw.stop:
				return
			}
		}
	}()
	return nil
}

func (sw *SignalWatcher) Notify() <-chan struct{} {
	return sw.event
}

func (sw *SignalWatcher) Close() {
	close(sw.stop)
}
