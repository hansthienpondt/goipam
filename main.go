package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"net"

	//"github.com/hansthienpondt/goipam/pkg/labels"

	"github.com/hansthienpondt/goipam/pkg/table"
	"inet.af/netaddr"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func main() {
	rtable := table.NewRouteTable()

	// Printing the size of the Radix/Patricia tree
	fmt.Println("The tree contains", rtable.Size(), "prefixes")
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	cidrs := map[string]map[string]string{
		"10.0.0.0/8": {
			"description": "rfc1918",
		},
		"10.0.0.0/16": {
			"description": "10.0/16-subnet",
		},
		"10.1.0.0/16": {
			"description": "10.1/16-subnet",
		},
		"10.0.0.0/24": {
			"description": "10.0.0/24-subnet",
		},
		"10.0.1.0/24": {
			"description": "10.0.1/24-subnet",
		},
		"192.0.0.0/12": {
			"rir": "RIPE",
		},
		"192.168.0.0/16": {
			"type": "aggregate",
		},
		"192.169.0.0/16": {
			"type": "aggregate",
		},
		"192.168.0.0/24": {
			"type": "prefix",
		},
		"192.168.0.0/25": {
			"type":        "prefix",
			"description": "hans1",
		},
		"192.168.0.128/25": {
			"type":        "prefix",
			"description": "hans2",
		},
		"85.255.192.0/12": {
			"type": "prefix",
			"rir":  "RIPE",
		},
	}
	ProdLabel := map[string]string{"env": "prod"}

	for k, v := range cidrs {
		route := table.NewRoute(netaddr.MustParseIPPrefix(k).Masked())
		route.UpdateLabel(v)
		route.UpdateLabel(ProdLabel)
		fmt.Printf("Adding CIDR %s with labels '%s' to the IPAM table\n", route.String(), route.GetLabels())
		rtable.Add(route)
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
	// Printing the size of the Radix/Patricia tree
	fmt.Println("The tree contains", rtable.Size(), "prefixes")
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
	//ipamRoute := table.NewRoute(netaddr.MustParseIPPrefix("20.0.0.0/16"))
	//ipamRoute.Attr.SetAttribute(encoding.Attribute{Key: "insertedby", Value: "hans"})
	//ipamRoute.Attr.SetAttribute(encoding.Attribute{Key: "2ndlabel", Value: "hans2"})
	//rtable.Add(ipamRoute.Cidr.IPNet(), ipamRoute.Attr)

	ipam2Route := table.NewRoute(netaddr.MustParseIPPrefix("192.168.0.192/26"))
	ipam2Route.UpdateLabel(map[string]string{"foo": "bar", "boo": "hoo"})
	rtable.Add(ipam2Route)
	fmt.Println(ipam2Route)
	fmt.Println(ipam2Route.Has("foo"))

	fmt.Println("WalkMatch -- find the parent for a given route")
	route1, _ := netaddr.ParseIPPrefix("192.168.0.255/32")
	//c := rtable.Parents2(route1)
	//fmt.Println(c)
	c, _ := json.MarshalIndent(rtable.Parents(route1), "", "  ")
	fmt.Println(string(c))

	fmt.Println("Walk -- find the children for a given route")
	route2, _ := netaddr.ParseIPPrefix("10.0.0.0/16")
	d := rtable.Children(route2)
	fmt.Println(d)

	fmt.Println("----")
	// Find free prefixes within a certain prefix
	var t netaddr.IPSetBuilder

	t.AddPrefix(netaddr.MustParseIPPrefix("10.0.0.0/16"))
	e := rtable.ChildrenByCIDR(netaddr.MustParseIPPrefix("10.0.0.0/16"))
	for _, pfx := range e {
		t.RemovePrefix(pfx)
	}
	set, _ := t.IPSet()
	fmt.Println(set.Prefixes())

	fmt.Println(rtable.FindFreePrefix(netaddr.MustParseIPPrefix("10.0.0.0/16"), 19))

	// Allocate a CIDR with prefixlength X , return the prefix and new list of free prefixes.
	p, newSet, _ := set.RemoveFreePrefix(24)
	fmt.Println(p, newSet.Prefixes())

	fmt.Println(set.ContainsPrefix(netaddr.MustParseIPPrefix("11.255.0.0/16")))

	//pre := table.NewRoute(netaddr.MustParseIPPrefix("10.0.0.0/16"))
	//pre.Attr.SetAttribute(encoding.Attribute{Key: "foo", Value: "bar"})
	//pre.Attr.SetAttribute(encoding.Attribute{Key: "boo", Value: "hoo"})
	//pre.Attr.SetAttribute(encoding.Attribute{Key: "foo", Value: "bar"})

	//fmt.Println(*pre)
	//fmt.Printf("%T %T", pre.Cidr, pre.Label)

	j, _ := json.MarshalIndent(rtable.GetTable(), "", "  ")
	fmt.Println(string(j))
	fmt.Println(strings.Repeat("#", 64))

	for _, v := range rtable.GetTable() {
		fmt.Println(v.String(), v.GetLabels())
	}
	fmt.Println(strings.Repeat("#", 64))

	iprange := netaddr.MustParseIPRange("1.2.3.4-5.6.7.8")
	fmt.Println(iprange.String())

	fmt.Println(rtable.GetCIDR("85.240.192.0/12"))
	_, ci, _ := net.ParseCIDR("85.255.192.0/12")
	ci2, _ := netaddr.ParseIPPrefix("85.255.192.0/12")
	fmt.Println(ci, ci2.Masked())
	fmt.Println(rtable.Get(ci2))
	fmt.Println(rtable.Match(ci2))

	selector := labels.NewSelector()
	req, _ := labels.NewRequirement("type", selection.NotIn, []string{"aggregate"})
	selector = selector.Add(*req)
	fmt.Printf("final selector: %v\n", selector.String())
	fmt.Printf("%T %v\n", selector, selector.String())

	sel, _ := labels.Parse("type notin (aggregate), description!=hans1")
	for _, v := range rtable.GetByLabel(sel) {
		fmt.Println(v.String(), v.GetLabels())

		//fmt.Println(v.GetChildren(rtable))
		free, ok := v.FindFreePrefix(rtable, 26)
		if ok {
			fmt.Println(free)
		}
	}

	fmt.Println(strings.Repeat("#", 64))
}
