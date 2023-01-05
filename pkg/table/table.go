package table

import (
	"encoding/binary"
	"net/netip"
	"sync"

	"github.com/kentik/patricia"
	"github.com/kentik/patricia/generics_tree"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

type RIB struct {
	mu   sync.RWMutex
	tree *generics_tree.TreeV6[labels.Set]
}

type RIBIterator struct {
	iter *generics_tree.TreeIteratorV6[labels.Set]
}

func NewRIB() *RIB {
	return &RIB{tree: generics_tree.NewTreeV6[labels.Set]()}
}

func (rib RIB) Add(r route) error {
	var p patricia.IPv6Address

	rib.mu.Lock()
	defer rib.mu.Unlock()
	if r.cidr.Addr().Is4() {
		p = patricia.NewIPv6Address(netip.AddrFrom16(r.cidr.Addr().As16()).AsSlice(), uint(96+r.cidr.Bits()))
	}
	if r.cidr.Addr().Is6() {
		p = patricia.NewIPv6Address(r.cidr.Addr().AsSlice(), uint(r.cidr.Bits()))
	}
	rib.tree.Add(p, r.labels, nil)
	return nil
}

func (rib RIB) Iterate() *RIBIterator {
	return &RIBIterator{
		iter: rib.tree.Iterate()}
}

func (rib RIB) GetTable() (r routes) {
	iter := rib.Iterate()

	for iter.Next() {
		r = append(r, iter.Route())
	}
	return r
}

func (rib RIB) Children(cidr netip.Prefix) (r routes) {
	var rte route

	iter := rib.Iterate()

	for iter.Next() {
		rte = iter.Route()
		if rte.cidr.Overlaps(cidr) && rte.cidr.Bits() > cidr.Bits() {
			r = append(r, iter.Route())
		}
	}
	return r
}

func (rib RIB) Parents(cidr netip.Prefix) (r routes) {
	var rte route

	iter := rib.Iterate()

	for iter.Next() {
		rte = iter.Route()
		if rte.cidr.Overlaps(cidr) && rte.cidr.Bits() < cidr.Bits() {
			r = append(r, iter.Route())
		}
	}
	return r
}
func (i *RIBIterator) Next() bool {
	return i.iter.Next()
}

func (i *RIBIterator) Route() route {
	var addr netip.Addr
	var bits int
	var addrSlice []byte
	var ok bool
	var p netip.Prefix
	var label labels.Set

	addrSlice = make([]byte, 16)
	binary.BigEndian.PutUint64(addrSlice, i.iter.Address().Left)
	binary.BigEndian.PutUint64(addrSlice[8:], i.iter.Address().Right)

	addr, ok = netip.AddrFromSlice(addrSlice)
	if !ok {
		log.Errorf("Unable to set address from Slice")
	}
	bits = int(i.iter.Address().Length)

	if addr.Is4In6() {
		p = netip.PrefixFrom(netip.AddrFrom4(addr.As4()), bits-96)
	} else {
		p = netip.PrefixFrom(addr, bits)
	}

	for _, l := range i.iter.Tags() {
		label = labels.Merge(label, l)
	}
	return route{
		cidr:   p,
		labels: label,
	}
}
