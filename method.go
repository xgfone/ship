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
	"fmt"
	"reflect"
	"strings"
)

// DefaultMethodMapping is the default mapping to map the method into router.
var DefaultMethodMapping = map[string]string{
	"Create": "POST",
	"Delete": "DELETE",
	"Update": "PUT",
	"Get":    "GET",
}

// MapMethodIntoRouter maps the struct method into the router,
// and returns the mapped paths.
//
// If the return value is nil, it represents that no method is mapped.
//
// By default, mapping is DefaultMethodMapping if not given.
//
// Example
//
//    type TestStruct struct{}
//    func (t TestStruct) Create(ctx ship.Context) error { return nil }
//    func (t TestStruct) Delete(ctx ship.Context) error { return nil }
//    func (t TestStruct) Update(ctx ship.Context) error { return nil }
//    func (t TestStruct) Get(ctx ship.Context) error    { return nil }
//    func (t TestStruct) Has(ctx ship.Context) error    { return nil }
//    func (t TestStruct) NotHandler()                   {}
//
//    ts := TestStruct{}
//    router := NewRouter()
//    paths := MapMethodIntoRouter(router, ts, "/v1")
//
// It's equal to the operation as follow:
//
//    router.Get("/v1/teststruct/get", ts.Get, "teststruct_get")
//    router.Put("/v1/teststruct/update", ts.Update, "teststruct_update")
//    router.Post("/v1/teststruct/create", ts.Create, "teststruct_create")
//    router.Delete("/v1/teststruct/delete", ts.Delete, "teststruct_delete")
//
// If you don't like the default mapping policy, you can give the customized
// mapping by the last argument, the key of which is the name of the method
// of the type, and the value of that is the request method, such as GET, POST,
// etc.
//
// Notice: the name of type and method will be converted to the lower.
func MapMethodIntoRouter(router Router, _struct interface{}, prefix string,
	mapping ...map[string]string) (paths []string) {

	if _struct == nil {
		panic(fmt.Errorf("the struct must no be nil"))
	}

	if prefix == "/" {
		prefix = ""
	}

	value := reflect.ValueOf(_struct)
	methodMaps := DefaultMethodMapping
	if len(mapping) > 0 {
		methodMaps = mapping[0]
	}

	var err error
	var ctx Context
	errType := reflect.TypeOf(&err).Elem()
	ctxType := reflect.TypeOf(&ctx).Elem()

	_type := reflect.TypeOf(_struct)
	typeName := strings.ToLower(_type.Name())
	for i := _type.NumMethod() - 1; i >= 0; i-- {
		method := _type.Method(i)
		mtype := method.Type

		// func (s StructType) Handler(ctx Context) error
		if mtype.NumIn() != 2 || mtype.NumOut() != 1 {
			continue
		}
		if !mtype.In(1).Implements(ctxType) {
			continue
		}
		if !mtype.Out(0).Implements(errType) {
			continue
		}

		if reqMethod := methodMaps[method.Name]; reqMethod != "" {
			methodName := strings.ToLower(method.Name)
			path := fmt.Sprintf("%s/%s/%s", prefix, typeName, methodName)
			router.Methods([]string{reqMethod}, path, HandlerFunc(func(ctx Context) error {
				vs := method.Func.Call([]reflect.Value{value, reflect.ValueOf(ctx)})
				return vs[0].Interface().(error)
			}), fmt.Sprintf("%s_%s", typeName, methodName))

			paths = append(paths, path)
		}
	}

	return
}
