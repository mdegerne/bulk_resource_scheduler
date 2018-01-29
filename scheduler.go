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
	"sort"
	"strings"
)

// Matches compares the list of Properties in the res to the list of
// Properties and returns true if the Resource is acceptable. Acceptable
// means:
//
// 1. all Properties in req which have the Sense of
// Required have a corresponding (Name() is same) Property in res that
// fullfills res_property.match(req_property).
//
// 2. all Properties in req which have a Sense of Never DO NOT have
// a corresponding property in res which fullfills
// res_property.match(req_property)
func Matches(req Requirement, res Resource) (acceptable bool, preference int) {
	acceptable = true
	preference = 0
	for _, p := range req.Properties() {
		switch p.Sense() {
		case Require:
			prop, found := res.Properties()[p.Name()]
			if !found {
				acceptable = false
			}
			m, ok := p.Matches(prop)
			if ok == nil && (!m) {
				acceptable = false
			}
		case Prefer:
			prop, found := res.Properties()[p.Name()]
			if !found {
				preference += 1
			}
			m, ok := p.Matches(prop)
			if ok == nil && (!m) {
				preference += 1
			}
		case Avoid:
			prop, found := res.Properties()[p.Name()]
			m, _ := p.Matches(prop)
			if found && m {
				preference -= 1
			}
		case Never:
			prop, found := res.Properties()[p.Name()]
			m, _ := p.Matches(prop)
			if found && m {
				acceptable = false
			}
		}
	}
	return
}

type respref struct {
	res  Resource
	pref int
}
type resprefs []respref

var rprefless = func(rps resprefs, i, j int) bool {
	return true
}

func (r resprefs) Len() int {
	return len(r)
}
func (r resprefs) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r resprefs) Less(i, j int) bool {
	return rprefless(r, i, j)
}

type ByNeed []Requirement

var byneedless = func(b ByNeed, i, j int) bool {
	min_i, _ := b[i].Count()
	min_j, _ := b[j].Count()
	return min_i < min_j
}

func (b ByNeed) Len() int {
	return len(b)
}
func (b ByNeed) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b ByNeed) Less(i, j int) bool {
	return byneedless(b, i, j)
}

// Schedule matches resources to requirements such that each resource
// fullfills 0 or 1 requirements with all of thre requirements being met.
// If not all requirements can be met, an error is returned and the map
// will not be complete.
// 
// Note:
// Name() is used as a proxy for both Requests and Resources in maps so must be unique
func Schedule(resources []Resource,
	requirements []Requirement) (map[string]Requirement, error) {
	// TODO: Reduce runtime complexity - currently O(nRes * nReq)
	// TODO: Reduce geographic complexity - Better structure(s)
	acceptable := make(map[string][]respref)
	n_assigned := make(map[string]int)
	acceptable_to := make(map[string][]Requirement)
	assigned := make(map[string]Requirement)
	var errs []string
	for _, req := range requirements {
		for _, res := range resources {
			acc, pref := Matches(req, res)
			if acc {
                if acceptable[req.Name()] == nil {
                    acceptable[req.Name()] = []respref{ respref{res, pref} }
                } else {
                    acceptable[req.Name()] = append(acceptable[req.Name()], respref{res, pref})
                }
                if acceptable_to[res.Name()] == nil {
                    acceptable_to[res.Name()] = []Requirement{ req }
                } else {
                    acceptable_to[res.Name()] = append(acceptable_to[res.Name()], req)
                }
			}
		}
	}
	s_requirements := make([]Requirement, len(requirements))
	copy(s_requirements, requirements)
	byneedless = func(b ByNeed, i, j int) bool {
		min_i, _ := b[i].Count()
		min_j, _ := b[j].Count()
		acc_i := len(acceptable[b[i].Name()])
		acc_j := len(acceptable[b[j].Name()])
		return (acc_i - min_i) < (acc_j - min_j)
	}
	sort.Sort(ByNeed(s_requirements))
	rprefless = func(rps resprefs, i, j int) bool {
		_, i_assigned := assigned[rps[i].res.Name()]
		_, j_assigned := assigned[rps[j].res.Name()]
		if i_assigned || j_assigned {
			return j_assigned
		}
		if rps[i].pref != rps[j].pref {
			return rps[i].pref < rps[j].pref
		}
		i_acc_to := len(acceptable_to[rps[i].res.Name()])
		j_acc_to := len(acceptable_to[rps[j].res.Name()])
		return i_acc_to < j_acc_to
	}
	for _, req := range s_requirements {
		// sort acceptable by prefer/avoid (primary), num_acceptable_to (secondary)
        if len(acceptable[req.Name()]) > 1 {
            sort.Sort(resprefs(acceptable[req.Name()]))
        }
		// fill minimum requirement from acceptable
		min, _ := req.Count()
		for i := 0; i < len(acceptable[req.Name()]) && n_assigned[req.Name()] < min; i = i+1 {
            rp := acceptable[req.Name()][i]
			if _, ok := assigned[rp.res.Name()]; !ok {
				assigned[rp.res.Name()] = req
				n_assigned[req.Name()] += 1
			}
		}
		if min > n_assigned[req.Name()] {
			errs = append(errs, fmt.Sprintf("Unable to find $d resources for %s requirement", min, req.Name()))
		}
	}
	for _, req := range requirements {
		_, max := req.Count()
		for i := 0; i < len(acceptable[req.Name()]) && n_assigned[req.Name()] < max; i = i + 1 {
            rp := acceptable[req.Name()][i]
			if _, ok := assigned[rp.res.Name()]; !ok {
				assigned[rp.res.Name()] = req
				n_assigned[req.Name()] += 1
			}
		}
	}
	if len(errs) > 0 {
		return assigned, errors.New(strings.Join(errs, "\n"))
	}
	return assigned, nil
}
