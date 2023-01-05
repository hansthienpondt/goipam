package table

import (
	"encoding/json"
	"fmt"
	"net/netip"

	"k8s.io/apimachinery/pkg/labels"
)

type route struct {
	cidr   netip.Prefix
	labels labels.Set
}

type routes []route

func NewRoute(cidr netip.Prefix, l map[string]string) route {
	var label labels.Set
	if l == nil {
		label = labels.Set{}
	} else {
		label = labels.Set(l)
	}

	return route{
		cidr:   cidr.Masked(),
		labels: label,
	}
}

func (r route) Equal(r2 route) bool {
	if r.cidr == r2.cidr && labels.Equals(r.labels, r2.labels) {
		return true
	}
	return false
}

func (r route) String() string {
	var s string
	s = fmt.Sprintf("%s %s", r.cidr.String(), r.labels.String())
	return s
}

func (r route) Prefix() netip.Prefix {
	return r.cidr
}

func (r route) Labels() labels.Set {
	return r.labels
}

// satisfy the k8s labels.Label interface
func (r route) Get(label string) string {
	return r.labels.Get(label)
}

// satisfy the k8s labels.Label interface
func (r route) Has(label string) bool {
	return r.labels.Has(label)
}

func (r route) Children(rib *RIB) routes {
	return rib.Children(r.cidr)
}
func (r route) Parents(rib *RIB) routes {
	return rib.Parents(r.cidr)
}

func (r route) UpdateLabel(label map[string]string) route {
	r.labels = labels.Merge(labels.Set(label), r.labels)

	return r
}

// Satisfy the json Interface.
func (r route) MarshalJSON() ([]byte, error) {
	var result map[string]string
	result = make(map[string]string)
	result[r.cidr.String()] = r.labels.String()
	return json.Marshal(result)
}
