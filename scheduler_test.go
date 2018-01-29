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

// bulk_resource_scheduler is a Go library intended to match resources
// to requirements such that each resource fullfills 0 or 1 requirement.
package bulk_resource_scheduler

import (
    "errors"
    "testing"
)

type tprop struct {
    name string
    sense Sense
    val int
}
func (p tprop) Name() string {
    return p.name
}
func (p tprop) Sense() Sense {
    return p.sense
}
func (p tprop) Matches(mprop Property) (bool, error) {
    tp, ok := mprop.(tprop)
    if !ok {
        return false, errors.New("Wrong Type")
    }
    if tp.name != p.name {
        return false, errors.New("Names don't match")
    }
    if tp.val == p.val {
        return true, nil
    }
    return false, nil
}

type tres struct {
    name string
    props map[string]Property
}
func (tr tres) Name() string {
    return tr.name
}
func (tr tres) Properties() map[string]Property {
    return tr.props
}
func (tr tres) AddProp(tp tprop) {
    if tr.props == nil {
        tr.props = make(map[string]Property)
    }
    tr.props[tp.name] = tp
}

type treq struct {
    name string
    props []Property
    min, max int
}
func (tr treq) Name() string {
    return tr.name
}
func (tr treq) Properties() []Property {
    return tr.props
}
func (tr treq) Count() (Min, Max int) {
    return tr.min, tr.max
}
func (tr treq) AddProp(tp tprop) {
    if tr.props == nil {
        tr.props = []Property { tp }
        return
    }
    tr.props = append(tr.props, tp)
}

func TestMatch1prop1res1req(t *testing.T) {
    var resource tres
    resource.name = "testres"
    resprop := tprop{"n1", Require, 1}
    resource.AddProp(resprop)
    var requirement treq
    requirement.name = "need1"
    requirement.min = 1
    requirement.max = 1
    reqprop := tprop{"n1", Require, 1}
    requirement.AddProp(reqprop)
    // verify prop match:
    match, ok := reqprop.Matches(resprop)
    if !(match && ok == nil) {
        t.Errorf("Prop Matches failed - Test Not set up right")
    }
    m, _ := Matches(requirement, resource)
    if !m {
        t.Errorf("Match failed - should be identical")
    }
    reqs := []Requirement { requirement }
    ress := []Resource { resource }
    result, err := Schedule(ress, reqs)
    if (err != nil) {
        t.Errorf("Unexpected error: %T", err)
    }
    if len(result) != 1 {
        t.Errorf("No result, wtf?")
    }
    rs_req := result[resource.Name()]
    if rs_req.Name() != requirement.Name() {
        t.Errorf("rs_req (%T) != requirement (%T)", rs_req, requirement)
    }
}
