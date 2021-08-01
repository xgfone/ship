// Copyright 2020 xgfone
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

import "fmt"

func ExampleSplitHostPort() {
	var host, port string

	host, port = SplitHostPort("www.example.com")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("www.example.com:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort(":80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("1.2.3.4:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("[fe80::1122:3344:5566:7788]")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("[fe80::1122:3344:5566:7788]:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	// Output:
	// Host: www.example.com, Port: #
	// Host: www.example.com, Port: 80#
	// Host: , Port: 80#
	// Host: 1.2.3.4, Port: 80#
	// Host: fe80::1122:3344:5566:7788, Port: #
	// Host: fe80::1122:3344:5566:7788, Port: 80#
}
