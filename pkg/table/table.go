package table

import (
	"fmt"
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
	if _, ok, _ := r.Get(route.IPPrefix()); ok {
		return fmt.Errorf("%q already exists in the routing table.", route.String())
	}
	// Custom Logic goes here.
	return r.tree.Add(route.IPNet(), route)
}

func (r *RouteTable) AddRange(iprange netaddr.IPRange) error {
	var bldr netaddr.IPSetBuilder
	bldr.AddRange(iprange)
	ipset, err := bldr.IPSet()
	if err != nil {
		return err
	}
	// Check all prefixes before adding the range.
	for _, ipprefix := range ipset.Prefixes() {
		_, ok, _ := r.Get(ipprefix)
		if ok {
			return fmt.Errorf("%q already exists in the routing table, not adding %q", ipprefix, iprange)
		}
	}
	// Add the different CIDRs to the routing table
	for _, ipprefix := range ipset.Prefixes() {
		route := NewRoute(ipprefix)
		route.UpdateLabel(map[string]string{"iprange": iprange.String()})
		if err := r.Add(route); err != nil {
			return err
		}
	}
	return nil
}

func (r *RouteTable) Clear() {
	r.tree.Clear()
}

func (r *RouteTable) Children(cidr netaddr.IPPrefix) (routes Routes) {
	f := func(n *net.IPNet, v interface{}) bool {
		if route, ok := v.(*Route); ok {
			if route.IPPrefix().Bits() > cidr.Bits() {
				routes = append(routes, route)
			}
		}
		return true
	}
	r.tree.WalkPrefix(cidr.IPNet(), f)
	return
}

func (r *RouteTable) ContainedIP(ip netaddr.IP) (contained bool, err error) {
	return r.tree.ContainedIP(ip.IPAddr().IP)
}

func (r *RouteTable) Delete(route *Route) (value interface{}, ok bool, err error) {
	// Custom Logic goes here.

	return r.tree.Delete(route.IPNet())
}

func (r *RouteTable) freePrefixes(cidr netaddr.IPPrefix) (s *netaddr.IPSet, ok bool) {
	var bldr netaddr.IPSetBuilder
	bldr.AddPrefix(cidr)
	for _, pfx := range r.Children(cidr) {
		bldr.RemovePrefix(pfx.IPPrefix())
	}
	s, err := bldr.IPSet()
	if err != nil {
		return &netaddr.IPSet{}, false
	}
	return s, true
}
func (r *RouteTable) FreePrefixes(cidr netaddr.IPPrefix) (l []netaddr.IPPrefix, ok bool) {
	if set, ok := r.freePrefixes(cidr); ok {
		l := set.Prefixes()
		return l, ok
	}
	return []netaddr.IPPrefix{}, false
}
func (r *RouteTable) FindFreePrefix(cidr netaddr.IPPrefix, bitlen uint8) (p netaddr.IPPrefix, ok bool) {
	if set, ok := r.freePrefixes(cidr); ok {
		p, _, ok = set.RemoveFreePrefix(bitlen)
		return p, ok
	}
	return netaddr.IPPrefix{}, false
}

func (r *RouteTable) Get(n netaddr.IPPrefix) (route *Route, ok bool, err error) {
	value, ok, err := r.tree.Get(n.Masked().IPNet())
	if route, ok := value.(*Route); ok {
		return route, ok, err
	}
	return nil, ok, err
}

func (r *RouteTable) GetByLabel(selector labels.Selector) (routes Routes) {
	f := func(n *net.IPNet, v interface{}) bool {
		if route, ok := v.(*Route); ok {
			if selector.Matches(route.GetLabels()) {
				routes = append(routes, route)
			}
		}
		return true
	}
	r.tree.Walk(nil, f)
	return
}

func (r *RouteTable) GetTable() (routes Routes) {
	f := func(n *net.IPNet, v interface{}) bool {
		if route, ok := v.(*Route); ok {
			routes = append(routes, route)
		}
		return true
	}
	r.tree.Walk(nil, f)
	return
}

func (r *RouteTable) Match(n netaddr.IPPrefix) (route *Route, err error) {
	_, value, err := r.tree.Match(n.Masked().IPNet())
	if route, ok := value.(*Route); ok {
		return route, err
	}
	return nil, err
}

func (r *RouteTable) MatchIP(ip netaddr.IP) (route *Route, err error) {
	_, value, err := r.tree.MatchIP(ip.IPAddr().IP)
	if route, ok := value.(*Route); ok {
		return route, err
	}
	return nil, err
}

func (r *RouteTable) Parents(cidr netaddr.IPPrefix) (routes Routes) {
	f := func(n *net.IPNet, v interface{}) bool {
		if route, ok := v.(*Route); ok {
			if route.IPPrefix().Bits() < cidr.Bits() {
				routes = append(routes, route)
			}
		}
		return true
	}
	r.tree.WalkMatch(cidr.IPNet(), f)
	return
}

func (r *RouteTable) Size() int {
	return r.tree.Size()
}

func (r *RouteTable) Update(route *Route) (err error) {
	// Custom Logic goes here.

	return r.tree.Add(route.IPNet(), route)
}
