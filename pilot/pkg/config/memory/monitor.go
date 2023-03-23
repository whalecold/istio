// Copyright Istio Authors
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

package memory

import (
	"sync"

	"go.uber.org/atomic"
	"istio.io/istio/pilot/pkg/model"
	config2 "istio.io/istio/pkg/config"
	"istio.io/pkg/log"
)

const (
	// BufferSize specifies the buffer size of event channel
	BufferSize = 100
)

// Handler specifies a function to apply on a Config for a given event type
type Handler func(config2.Config, config2.Config, model.Event)

// Monitor provides methods of manipulating changes in the config store
type Monitor interface {
	Run(<-chan struct{})
	AppendEventHandler(config2.GroupVersionKind, Handler)
	ScheduleProcessEvent(ConfigEvent)
}

// ConfigEvent defines the event to be processed
type ConfigEvent struct {
	config config2.Config
	old    config2.Config
	event  model.Event
}

type configStoreMonitor struct {
	store    model.ConfigStore
	handlers map[config2.GroupVersionKind][]Handler
	eventCh  chan ConfigEvent
	// If enabled, events will be handled synchronously
	sync bool

	closed atomic.Bool
	lock   sync.Locker
}

// NewMonitor returns new Monitor implementation with a default event buffer size.
func NewMonitor(store model.ConfigStore) Monitor {
	return newBufferedMonitor(store, BufferSize, false)
}

// NewSyncMonitor returns new Monitor implementation which will process events synchronously
func NewSyncMonitor(store model.ConfigStore) Monitor {
	return newBufferedMonitor(store, BufferSize, true)
}

// newBufferedMonitor returns new Monitor implementation with the specified event buffer size
func newBufferedMonitor(store model.ConfigStore, bufferSize int, syncMod bool) Monitor {
	handlers := make(map[config2.GroupVersionKind][]Handler)

	for _, s := range store.Schemas().All() {
		handlers[s.Resource().GroupVersionKind()] = make([]Handler, 0)
	}

	return &configStoreMonitor{
		store:    store,
		handlers: handlers,
		eventCh:  make(chan ConfigEvent, bufferSize),
		sync:     syncMod,
		lock:     &sync.RWMutex{},
	}
}

func (m *configStoreMonitor) ScheduleProcessEvent(configEvent ConfigEvent) {
	m.lock.Lock()
	if m.closed.Load() {
		m.lock.Unlock()
		return
	}
	if m.sync {
		m.lock.Unlock()
		m.processConfigEvent(configEvent)
	} else {
		m.eventCh <- configEvent
		m.lock.Unlock()
	}
}

// waitAndCleanup the close action should after the stop closed to make sure all the
// events in the chan have been consumed.
func (m *configStoreMonitor) waitAndCleanup(stop <-chan struct{}) {
	<-stop

	m.lock.Lock()
	defer m.lock.Unlock()

	m.closed.Store(true)
	close(m.eventCh)
}

func (m *configStoreMonitor) Run(stop <-chan struct{}) {
	if m.sync {
		m.waitAndCleanup(stop)
		return
	}
	// Run should
	go m.waitAndCleanup(stop)

	for ce := range m.eventCh {
		m.processConfigEvent(ce)
	}
}

func (m *configStoreMonitor) processConfigEvent(ce ConfigEvent) {
	if _, exists := m.handlers[ce.config.GroupVersionKind]; !exists {
		log.Warnf("Config GroupVersionKind %s does not exist in config store", ce.config.GroupVersionKind)
		return
	}
	m.applyHandlers(ce.old, ce.config, ce.event)
}

func (m *configStoreMonitor) AppendEventHandler(typ config2.GroupVersionKind, h Handler) {
	m.handlers[typ] = append(m.handlers[typ], h)
}

func (m *configStoreMonitor) applyHandlers(old config2.Config, config config2.Config, e model.Event) {
	for _, f := range m.handlers[config.GroupVersionKind] {
		f(old, config, e)
	}
}
