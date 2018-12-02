// Copyright 2018 xgfone <xgfone@126.com>
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
	"testing"
)

func TestURLParams(t *testing.T) {
	ups := NewURLParam(3)
	ups.Set("n1", "v1")
	ups.Set("n2", "v2")

	if ups.Get("n1") != "v1" || ups.Get("n2") != "v2" {
		t.Fail()
	}

	ups.Set("v1", "v3")

	if ups.Get("v1") != "v3" {
		t.Fail()
	}

	if _ups := ups.(*urlParams); len(_ups.kvs) != 3 {
		t.Fail()
	}
}
