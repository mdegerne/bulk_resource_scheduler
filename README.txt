
This is a package created largely as an exercise in learning the Go language. There is, I'm sure,
much to be desired in terms of stylistic issues as well as issues unique to a Go newbie. Use at
your own risk.

PACKAGE DOCUMENTATION

package bulk_resource_scheduler
    import "github.com/mdegerne/bulk_resource_scheduler"

    bulk_resource_scheduler is a Go library intended to match resources to
    requirements such that each resource fullfills 0 or 1 requirement.

CONSTANTS

const (
    Require = iota
    Prefer
    Avoid
    Never
)

FUNCTIONS

func Matches(req Requirement, res Resource) (acceptable bool, preference int)
    Matches compares the list of Properties in the res to the list of
    Properties and returns true if the Resource is acceptable. Acceptable
    means:

    1. all Properties in req which have the Sense of Required have a
    corresponding (Name() is same) Property in res that fullfills
    res_property.match(req_property).

    2. all Properties in req which have a Sense of Never DO NOT have a
    corresponding property in res which fullfills
    res_property.match(req_property)

func Schedule(resources []Resource,
    requirements []Requirement) (map[string]Requirement, error)
    Schedule matches resources to requirements such that each resource
    fullfills 0 or 1 requirements with all of thre requirements being met.
    If not all requirements can be met, an error is returned and the map
    will not be complete.

    Note: Name() is used as a proxy for both Requests and Resources in maps
    so must be unique

TYPES

type ByNeed []Requirement

func (b ByNeed) Len() int

func (b ByNeed) Less(i, j int) bool

func (b ByNeed) Swap(i, j int)

type Property interface {
    Name() string
    Matches(Property) (bool, error)
    Sense() Sense
}
    Property is a single characteristic used to identify a matching
    resource.

type Requirement interface {
    Name() string
    Properties() []Property
    Count() (Min int, Max int)
}
    Requirement is a request in search of a resource, all Properties must
    match corresponding Resource Properties for the Resource to match.

type Resource interface {
    Name() string
    Properties() map[string]Property // string must be property.Name()
}
    Resource is a resource available for matching.

type Sense int
    Sense defines how to interpret a matching characteristic.


