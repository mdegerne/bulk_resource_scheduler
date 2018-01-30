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
    "fmt"
    "reflect"
    "strings"
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
func (tr *tres) AddProp(tp tprop) {
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
func (tr *treq) AddProp(tp tprop) {
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
    match, err := reqprop.Matches(resprop)
    if !(match && err == nil) {
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
    if rs_req == nil || rs_req.Name() != requirement.Name() {
        t.Errorf("rs_req (%T) != requirement (%T)", rs_req, requirement)
    }
}

// Self test - test the test functions
func TestMatchProp(t *testing.T) {
    resprop := tprop{"n1", Require, 1}
    reqprop1 := tprop{"n2", Require, 1}
    _, err := reqprop1.Matches(resprop)
    if (err == nil) {
        t.Error("Expected failure to match - wrong name")
    }
    reqprop2 := tprop{"n1", Require, 2}
    match, err := reqprop2.Matches(resprop)
    if (err != nil) {
        t.Error("Unexpected failure of Matches")
    }
    if match {
        t.Error("Should not have matched reqprop2")
    }
    reqprop3 := tprop{"n1", Require, 1}
    match, err = reqprop3.Matches(resprop)
    if (err != nil) {
        t.Error("Unexpected failure of Matches")
    }
    if !match {
        t.Error("Should have matched reqprop3")
    }
}


type varTestsStruct struct {
    resprops []tprop
    reqprops []tprop
    match bool
    pref int
}
var varTests = []varTestsStruct {
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Require, 1}, tprop{"n2", Require, 1} }, true, 0 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Require, 1}, tprop{"n2", Require, 2} }, false, 0 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n4", Require, 1}, tprop{"n2", Require, 2} }, false, 0 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Prefer, 1}, tprop{"n2", Require, 1} }, true, 1 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n4", Prefer, 1}, tprop{"n2", Require, 1} }, true, 0 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Avoid, 1}, tprop{"n2", Require, 1} }, true, -1 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Never, 1}, tprop{"n2", Require, 1} }, false, 0 },
    { []tprop{ tprop{ "n1", Require, 1}, tprop{"n2", Require, 1}, tprop{"n3", Require, 1} }, []tprop{ tprop{ "n3", Never, 1}, tprop{"n2", Prefer, 1} }, false, 1 },
}

func TestVariants(t *testing.T) {
    var resource tres
    var requirement treq
    for i, tstruct := range varTests {
        resource.name = fmt.Sprintf("testres%d", i)
        resource.props = map[string]Property{}
        for _, resprop := range tstruct.resprops {
            resource.AddProp(resprop)
        }
        requirement.name = fmt.Sprintf("testreq%d", i)
        requirement.props = nil
        for _, reqprop := range tstruct.reqprops {
            requirement.AddProp(reqprop)
        }
        requirement.min = 1
        requirement.max = 1
        m, pref := Matches(requirement, resource)
        if m != tstruct.match {
            t.Errorf("Match failed (i=%d) %v != %v", i, m, tstruct.match)
        }
        if pref != tstruct.pref {
            t.Errorf("Match pref failed (i=%d) %v != %v", i, pref, tstruct.pref)
        }

    }
}

type schedulerTest struct {
    resources []tres
    requirements []treq
    expectedErrors []string
    expectedMap map[string]string
}

var stests = []schedulerTest {
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 2}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 2}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1}},
                1,
                2,
            },
        },
        []string{},
        map[string]string{"R1": "Req2", "R2": "Req1"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 2}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 2}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Prefer, 1}},
                1,
                2,
            },
        },
        []string{},
        map[string]string{"R1": "Req2", "R2": "Req1", "R3": "Req2"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 1}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Never, 1},tprop{"P2", Prefer, 1}},
                1,
                2,
            },
        },
        []string{"Req2"},
        map[string]string{"R1": "Req1", "R2": "Req1"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 1}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Prefer, 1}},
                1,
                2,
            },
        },
        []string{},
        map[string]string{"R1": "Req1", "R2": "Req1", "R3": "Req2"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
            tres{
                "R4",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R5",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 1}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Prefer, 1}},
                1,
                2,
            },
            treq{
                "Req3",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Require, 1}},
                1,
                1,
            },
        },
        []string{},
        map[string]string{"R1": "Req1", "R2": "Req1", "R3": "Req3", "R4": "Req2", "R5": "Req2"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
            tres{
                "R4",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R5",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 1}},
                1,
                3,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Prefer, 1}},
                1,
                2,
            },
            treq{
                "Req3",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Require, 1}},
                1,
                1,
            },
        },
        []string{},
        map[string]string{"R1": "Req1", "R2": "Req1", "R3": "Req3", "R4": "Req1", "R5": "Req2"},
    },
    schedulerTest{
        []tres{
            tres{
                "R1",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R2",
                map[string]Property{"P1": tprop{ "P1", Require, 1}},
            },
            tres{
                "R3",
                map[string]Property{"P1": tprop{ "P1", Require, 1}, "P2": tprop{ "P2", Require, 1}},
            },
        },
        []treq{
            treq{
                "Req1",
                []Property{tprop{"P1", Require, 1}},
                1,
                2,
            },
            treq{
                "Req2",
                []Property{tprop{"P1", Require, 1},tprop{"P2", Avoid, 1}},
                1,
                2,
            },
        },
        []string{},
        map[string]string{"R1": "Req1", "R2": "Req2", "R3": "Req1"},
    },
}

func TestSchedulerVariants(t *testing.T) {
    for i, tstruct := range stests {
        reslice := make([]Resource, len(tstruct.resources))
        for n, r := range tstruct.resources {
            reslice[n] = r
        }
        var reqslice []Requirement = make([]Requirement, len(tstruct.requirements))
        for n, r2 := range tstruct.requirements {
            reqslice[n] = r2
        }
        resmap, err := Schedule(reslice, reqslice)
        for _, serr := range tstruct.expectedErrors {
            if ! strings.Contains(err.Error(), serr) {
                t.Errorf("Case #%d:Expected error: \"%s\", not found in returned errors %+v\n", i, serr, err)
            }
        }
        fixedmap := map[string]string{}
        for k,v := range resmap {
            fixedmap[k] = v.Name()
        }
        if ! reflect.DeepEqual(fixedmap, tstruct.expectedMap) {
            t.Errorf("Case #%d:Expected result map: %+v, Actual result map: %+v", i, tstruct.expectedMap, fixedmap)
        }
    }
}
