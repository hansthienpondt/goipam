package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
		if err := rtable.Add(route); err != nil {
			fmt.Println(err)
		}
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
	// Printing the size of the Radix/Patricia tree
	fmt.Println("The tree contains", rtable.Size(), "prefixes")
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	// Inserting an additional single route with a label.
	ipam2Route := table.NewRoute(netaddr.MustParseIPPrefix("192.168.0.192/26"))
	//ipam2Route := table.NewRoute(netaddr.MustParseIPPrefix("192.168.0.128/25"))
	ipam2Route.UpdateLabel(map[string]string{"foo": "bar", "boo": "hoo"})
	if err := rtable.Add(ipam2Route); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Adding CIDR %q with labels %q to the IPAM table\n", ipam2Route.String(), ipam2Route.GetLabels())
		fmt.Printf("CIDR %q has label foo: %t\n", ipam2Route, ipam2Route.Has("foo"))
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	// Lookup Methods in the routing table.
	route1, _ := netaddr.ParseIPPrefix("192.168.0.255/32")
	fmt.Println("Finding the parents for route -- " + route1.String())
	// Marshal it as a JSON.
	c, _ := json.MarshalIndent(rtable.Parents(route1), "", "  ")
	fmt.Println(string(c))

	route2, _ := netaddr.ParseIPPrefix("10.0.0.0/16")
	fmt.Println("Finding the children for route -- " + route2.String())
	d := rtable.Children(route2)
	fmt.Println(d)
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	// Find free prefixes within a certain prefix
	findfree := netaddr.MustParseIPPrefix("10.0.0.0/16")
	var bitlen uint8 = 19
	fmt.Println("Finding a free /" + strconv.Itoa(int(bitlen)) + " prefix in: " + findfree.String())
	if pfx, ok := rtable.FreePrefixes(findfree); ok {
		fmt.Printf("All free prefixes are: %s\n", pfx)
	}

	if pfx, ok := rtable.FindFreePrefix(findfree, bitlen); ok {
		fmt.Println("Returned free prefix is: " + pfx.String())
	}

	//Alternate method: get Route object, search for free prefixes
	if freer1, ok, _ := rtable.Get(findfree); ok {
		pfx, _ := freer1.FindFreePrefix(rtable, bitlen)
		fmt.Println("Alt Method, Returned free prefix is: " + pfx.String())
	}

	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
	fmt.Println("Dumping route-table in JSON format:")
	j, _ := json.MarshalIndent(rtable.GetTable(), "", "  ")
	fmt.Println(string(j))
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
	fmt.Println("Dumping route-table in String format:")
	for _, v := range rtable.GetTable() {
		fmt.Println(v.String(), v.GetLabels())
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	iprange := netaddr.MustParseIPRange("193.168.0.0-193.168.0.132")
	if err := rtable.AddRange(iprange); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Iprange " + iprange.String() + " added to the routing table")
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))

	// Create a selector (label filter) to get routes by label
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement("type", selection.NotIn, []string{"aggregate"})
	selector = selector.Add(*req)
	// Alternate definition of a selector, define by string.
	sel, _ := labels.Parse("type notin (aggregate), description!=hans1")

	dumprtable1 := rtable.GetByLabel(selector)
	dumprtable2 := rtable.GetByLabel(sel)

	fmt.Println("Printing GetByLabel1 -- " + selector.String())
	for _, v := range dumprtable1 {
		fmt.Println(v, v.GetLabels())
	}
	fmt.Println("")
	fmt.Println("Printing GetByLabel2 -- " + sel.String())
	for _, v := range dumprtable2 {
		fmt.Println(v, v.GetLabels())
	}
	// Printing the Stdout seperator
	fmt.Println(strings.Repeat("#", 64))
}
