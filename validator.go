// Copyright 2021 xgfone
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

// Validator is used to validate the data is valid or not.
type Validator interface {
	Validate(data interface{}) error
}

// ValidatorFunc is a function type implementing the interface Validator.
type ValidatorFunc func(data interface{}) error

// Validate implements the interface Validator.
func (v ValidatorFunc) Validate(data interface{}) error { return v(data) }

// NothingValidator returns a Validator that does nothing.
func NothingValidator() Validator { return ValidatorFunc(nothingValidator) }

func nothingValidator(interface{}) error { return nil }
