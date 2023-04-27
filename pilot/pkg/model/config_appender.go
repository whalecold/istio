//  Copyright Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package model

import (
	"istio.io/istio/pkg/config"
)

const ListCacheSize = 5000

type ConfigFilter interface {
	//  Filter returns true if the input object resides in a namespace selected for discovery
	Filter(*config.Config) bool
}

type ConfigFilterFunc func(*config.Config) bool

func (f ConfigFilterFunc) Filter(c *config.Config) bool {
	return f(c)
}

// Filter returns true if the input object resides in a namespace selected for discovery
type Filter func(*config.Config) bool

type ConfigGetter interface {
	Configs() []*config.Config
}

type ConfigAppender interface {
	Append(config *config.Config)
}

type ConfigGetAppender interface {
	ConfigGetter
	ConfigAppender
}

func NewFilterConfigGetAppender(filters ...ConfigFilter) ConfigGetAppender {
	return &filterConfigGetAppender{
		filters: filters,
		configs: []*config.Config{},
	}
}

func NewFilterConfigGetAppenderSize(size int, filters ...ConfigFilter) ConfigGetAppender {
	return &filterConfigGetAppender{
		filters: filters,
		configs: make([]*config.Config, 0, size),
	}
}

type filterConfigGetAppender struct {
	filters []ConfigFilter
	configs []*config.Config
}

func (appender *filterConfigGetAppender) Configs() []*config.Config {
	return appender.configs
}

func (appender *filterConfigGetAppender) Append(c *config.Config) {
	for _, filter := range appender.filters {
		if !filter.Filter(c) {
			return
		}
	}
	appender.configs = append(appender.configs, c)
}

type filteredAppender struct {
	filters  []ConfigFilter
	appender ConfigAppender
}

func NewFilteredConfigAppender(appender ConfigAppender, filters ...ConfigFilter) ConfigAppender {
	return &filteredAppender{
		filters:  filters,
		appender: appender,
	}
}

func (fa *filteredAppender) Append(c *config.Config) {
	for _, filter := range fa.filters {
		if !filter.Filter(c) {
			return
		}
	}
	fa.appender.Append(c)
}
