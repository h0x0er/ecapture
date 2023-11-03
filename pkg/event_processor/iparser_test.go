// Copyright 2022 CFC4N <cfc4n.cs@gmail.com>. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event_processor

import (
	"testing"
)

func TestNewParser(t *testing.T) {

	payload := []byte{0, 0, 18, 4, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 100, 0, 4, 2, 0, 0, 0, 0, 2, 0, 0, 0, 0}

	tests := []struct {
		name    string
		payload []byte
		want    ParserType
	}{
		{name: "HTTP2Frame Parser", payload: payload, want: ParserTypeHttp2Frame},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewParser(tt.payload); got.ParserType() != tt.want {
				t.Errorf("NewParser() = %v, want %v", got, tt.want)
			}
		})
	}
}
