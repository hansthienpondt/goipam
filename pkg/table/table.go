package table

import (
	"net"

	// "github.com/hansthienpondt/goipam/pkg/labels"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/k-sone/critbitgo"
	"inet.af/netaddr"
)

// NewRouteTable function defines a new routing table

type RouteTable struct {
	tree *critbitgo.Net
}

func NewRouteTable() *RouteTable {
	return &RouteTable{tree: critbitgo.NewNet()}
}

func (r *RouteTable) Add(route *Route) (err error) {
	return r.tree.Add(route.IPNet(), route.GetLabels())
}
func (r *RouteTable) AddCIDR(s string, value interface{}) (err error) {
	cidr, err := netaddr.ParseIPPrefix(s)
	if err != nil {
		return err
	}
	return r.tree.Add(cidr.Masked().IPNet(), value)
}
func (r *RouteTable) Clear() {
	r.tree.Clear()
}

func (r *RouteTable) Size() int {
	return r.tree.Size()
}
func (r *RouteTable) Get(n netaddr.IPPrefix) (value interface{}, ok bool, err error) {
	return r.tree.Get(n.Masked().IPNet())
}
func (r *RouteTable) GetCIDR(s string) (value interface{}, ok bool, err error) {
	return r.tree.GetCIDR(s)
}

func (r *RouteTable) Match(n netaddr.IPPrefix) (cidr netaddr.IPPrefix, value interface{}, err error) {
	_, value, err = r.tree.Match(n.Masked().IPNet())
	return n, value, err
}

func (r *RouteTable) MatchCIDR(s string) (cidr netaddr.IPPrefix, value interface{}, err error) {
	net, value, err := r.tree.MatchCIDR(s)
	cidr, _ = netaddr.FromStdIPNet(net)

	return cidr, value, err
}

func (r *RouteTable) GetTable() []Route {
	var result []Route
	f := func(n *net.IPNet, v interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		labels, _ := v.(*labels.Set)
		route := NewRoute(pfx)
		route.UpdateLabel(*labels)
		result = append(result, *route)
		return true
	}
	result = []Route{}
	r.tree.Walk(nil, f)

	return result
}

func (r *RouteTable) GetTableByCIDR() map[netaddr.IPPrefix]interface{} {
	var result map[netaddr.IPPrefix]interface{}
	f := func(n *net.IPNet, v interface{}) bool {
		cidr, _ := netaddr.FromStdIPNet(n)
		result[cidr] = v
		return true
	}
	result = make(map[netaddr.IPPrefix]interface{})
	r.tree.Walk(nil, f)

	return result
}

func (r *RouteTable) ParentsByCIDR(cidr netaddr.IPPrefix) []netaddr.IPPrefix {
	var result []netaddr.IPPrefix
	f := func(n *net.IPNet, _ interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		if pfx.Bits() < cidr.Bits() {
			result = append(result, pfx)
		}
		return true
	}
	result = []netaddr.IPPrefix{}
	r.tree.WalkMatch(cidr.IPNet(), f)
	return result
}

func (r *RouteTable) Parents(cidr netaddr.IPPrefix) []Route {
	var result []Route
	f := func(n *net.IPNet, v interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		labels, _ := v.(*labels.Set)
		if pfx.Bits() < cidr.Bits() {
			route := NewRoute(pfx)
			route.UpdateLabel(*labels)
			result = append(result, *route)
		}
		return true
	}
	result = []Route{}
	r.tree.WalkMatch(cidr.IPNet(), f)
	return result
}

func (r *RouteTable) ChildrenByCIDR(cidr netaddr.IPPrefix) []netaddr.IPPrefix {
	var result []netaddr.IPPrefix
	f := func(n *net.IPNet, _ interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		if pfx.Bits() > cidr.Bits() {
			result = append(result, pfx)
		}
		return true
	}
	result = []netaddr.IPPrefix{}
	r.tree.WalkPrefix(cidr.IPNet(), f)
	return result
}

func (r *RouteTable) Children(cidr netaddr.IPPrefix) []Route {
	var result []Route
	f := func(n *net.IPNet, v interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		labels, _ := v.(labels.Set)
		if pfx.Bits() > cidr.Bits() {
			route := NewRoute(pfx)
			route.UpdateLabel(labels)
			result = append(result, *route)
		}
		return true
	}
	result = []Route{}
	r.tree.WalkPrefix(cidr.IPNet(), f)
	return result
}

func (r *RouteTable) GetByLabel(selector labels.Selector) []Route {
	var result []Route
	f := func(n *net.IPNet, v interface{}) bool {
		pfx, _ := netaddr.FromStdIPNet(n)
		labels, _ := v.(labels.Set)
		route := NewRoute(pfx)
		route.UpdateLabel(labels)
		if selector.Matches(route.GetLabels()) {
			result = append(result, *route)
		}
		return true
	}
	result = []Route{}
	r.tree.Walk(nil, f)
	return result
}

func (r *RouteTable) FindFreePrefix(cidr netaddr.IPPrefix, bitlen uint8) (p netaddr.IPPrefix, ok bool) {
	var bldr netaddr.IPSetBuilder
	bldr.AddPrefix(cidr)
	for _, pfx := range r.ChildrenByCIDR(cidr) {
		bldr.RemovePrefix(pfx)
	}

	s, err := bldr.IPSet()
	if err != nil {
		return netaddr.IPPrefix{}, false
	}
	p, _, ok = s.RemoveFreePrefix(bitlen)
	return p, ok
}
