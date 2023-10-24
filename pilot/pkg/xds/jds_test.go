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

package xds

import (
	"reflect"
	"testing"

	"istio.io/istio/pkg/util/sets"
)

func Test_buildServiceList(t *testing.T) {
	type args struct {
		resourceNames []string
	}
	tests := []struct {
		name string
		args args
		want map[string]sets.String
	}{
		{
			args: args{
				resourceNames: []string{"s1.ns1", "s2.ns1", "s3.ns2"},
			},
			want: map[string]sets.String{
				"ns1": {
					"s1": struct{}{},
					"s2": struct{}{},
				},
				"ns2": {
					"s3": struct{}{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildServiceList(tt.args.resourceNames); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildServiceList() = %v, want %v", got, tt.want)
			}
		})
	}
}
