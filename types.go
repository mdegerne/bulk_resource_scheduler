/*
   Copyright 2018 Mandell Degerness

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package bulk_resource_scheduler

// Modifier determines how to match Properties.
type Modifier int

const (
	Equal = iota
	Contains
	GreaterThanEqual
	LessThanEqual
)

// Sense defines how to interpret a matching characteristic.
type Sense int

const (
	Require = iota
	Prefer
	Avoid
	Never
)

// Property is a single characteristic used to identify a matching resource.
type Property interface {
	Name() string
	Matches(Property) (bool, error)
	Modifier() Modifier
	Sense() Sense
}

// Resource is a resource available for matching.
type Resource interface {
	Name() string
	Properties() []Property
}

// Requirement is a request in search of a resource, all Properties
// must match corresponding Resource Properties for the Resource to match.
type Requirement interface {
	Name() string
	Properties() []Property
	Count() (Min int, Max int)
}
