package table

import (
	"encoding/json"
	"net"

	// "github.com/hansthienpondt/goipam/pkg/labels"
	"k8s.io/apimachinery/pkg/labels"

	"inet.af/netaddr"
)

type Route struct {
	cidr   *netaddr.IPPrefix
	labels *labels.Set
}

func (r *Route) String() string {
	return r.cidr.String()
}
func (r *Route) MarshalJSON() ([]byte, error) {
	result := make(map[string]string)
	l := *r.GetLabels()
	result[r.String()] = l.String()
	return json.Marshal(result)
}
func (r *Route) UpdateLabel(label map[string]string) {
	var newlabel labels.Set
	newlabel = labels.Merge(labels.Set(label), *r.labels)
	r.labels = &newlabel
}
func (r *Route) GetParents(t *RouteTable) []Route {
	return t.Parents(*r.cidr)
}

func (r *Route) GetChildren(t *RouteTable) []Route {
	return t.Children(*r.cidr)
}

// satisfy k8s labels Interface
func (r *Route) Has(label string) bool {
	return r.labels.Has(label)
}

func (r *Route) Get(label string) string {
	return r.labels.Get(label)
}

func (r *Route) GetLabels() *labels.Set {
	return r.labels
}
func (r *Route) IPPrefix() netaddr.IPPrefix {
	return r.cidr.Masked()
}
func (r *Route) IPNet() *net.IPNet {
	return r.cidr.Masked().IPNet()
}

func (r *Route) FindFreePrefix(t *RouteTable, bitlen uint8) (p netaddr.IPPrefix, ok bool) {
	var bldr netaddr.IPSetBuilder
	bldr.AddPrefix(*r.cidr)
	for _, route := range r.GetChildren(t) {
		bldr.RemovePrefix(*route.cidr)
	}

	s, err := bldr.IPSet()
	if err != nil {
		return netaddr.IPPrefix{}, false
	}
	p, _, ok = s.RemoveFreePrefix(bitlen)
	return p, ok
}
func NewRoute(cidr netaddr.IPPrefix) *Route {
	return &Route{
		cidr:   &cidr,
		labels: &labels.Set{},
	}
}
