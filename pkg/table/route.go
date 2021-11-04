package table

import (
	"encoding/json"
	"net"

	"k8s.io/apimachinery/pkg/labels"

	"inet.af/netaddr"
)

type Route struct {
	cidr   *netaddr.IPPrefix
	labels *labels.Set
}

func NewRoute(cidr netaddr.IPPrefix) *Route {
	return &Route{
		cidr:   &cidr,
		labels: &labels.Set{},
	}
}

type Routes []*Route

// satisfy the fmt.Stringer Interface.
func (r *Route) String() string {
	return r.cidr.String()
}

// Helper function to find a free prefix with provided bitlen under the route/cidr.
func (r *Route) freePrefixes(t *RouteTable) (s *netaddr.IPSet, ok bool) {
	return t.freePrefixes(r.IPPrefix())
}

func (r *Route) FindFreePrefix(t *RouteTable, bitlen uint8) (p netaddr.IPPrefix, ok bool) {
	return t.FindFreePrefix(r.IPPrefix(), bitlen)
}

// satisfy the k8s labels.Label interface
func (r *Route) Get(label string) string {
	return r.labels.Get(label)
}

// Retrieve all children of a given route/cidr, main method implemented on struct RouteTable.
func (r *Route) GetChildren(t *RouteTable) Routes {
	return t.Children(*r.cidr)
}

// Retrieve all the labels assigned to a given route/cidr.
func (r *Route) GetLabels() *labels.Set {
	return r.labels
}

// Retrieve all parents of a given route/cidr, main method implemented on struct RouteTable.
func (r *Route) GetParents(t *RouteTable) Routes {
	return t.Parents(*r.cidr)
}

// satisfy the k8s labels.Label interface
func (r *Route) Has(label string) bool {
	return r.labels.Has(label)
}

// Helper function to provide std net.IPNet object.
func (r *Route) IPNet() *net.IPNet {
	return r.cidr.Masked().IPNet()
}

// Helper function to return netaddr.IPPrefix (Masked) object.
func (r *Route) IPPrefix() netaddr.IPPrefix {
	return r.cidr.Masked()
}

// Satisfy the json Interface.
func (r *Route) MarshalJSON() ([]byte, error) {
	result := make(map[string]string)
	l := *r.GetLabels()
	result[r.String()] = l.String()
	return json.Marshal(result)
}

// Helper function to merge the currently assigned labels with a new labels.Set
func (r *Route) UpdateLabel(label map[string]string) {
	var mergedlabel labels.Set
	mergedlabel = labels.Merge(labels.Set(label), *r.labels)
	r.labels = &mergedlabel
}

/*
func (r Routes) String() string {
	routes := make([]string, 0, len(r))
	for _, value := range r {
		routes = append(routes, )
	}
	return "yes"
}
*/
// Satisfy the json Interface.
func (rts Routes) MarshalJSON() ([]byte, error) {
	result := make(map[string]labels.Labels)
	for _, route := range rts {
		label := *route.GetLabels()
		result[route.String()] = label
	}
	return json.Marshal(result)
}
