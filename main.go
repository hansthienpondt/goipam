package main

import (
	"flag"
	"fmt"
	"net/netip"
	"os"

	"github.com/hansthienpondt/goipam/pkg/table"
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
	//log.SetLevel(log.WarnLevel)
	// Set Calling method to true
	log.SetReportCaller(true)
}
func main() {
	flag.Parse()
	log.Debugf("Program Initialized")

	rtable := table.NewRIB()
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
			"description": "test1",
			"rir":         "RIPE",
		},
		"192.168.0.0/16": {
			"description": "test2",
			"type":        "aggregate",
		},
		"192.169.0.0/16": {
			"description": "test3",
			"type":        "aggregate",
		},
		"192.168.0.0/24": {
			"description": "test4",
			"type":        "prefix",
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
			"type":        "prefix",
			"rir":         "RIPE",
			"description": "test5",
		},
		"100.255.254.1/31": {
			"type":        "prefix",
			"rir":         "RIPE",
			"description": "test6",
		},
		"100.255.254.1/35": {
			"type":        "prefix",
			"rir":         "RIPE",
			"description": "test6",
		},
		"2a02:1800::/24": {
			"type":        "prefix",
			"rir":         "RIPE",
			"family":      "ipv6",
			"description": "test7",
		},
	}

	//fmt.Println(rtable)

	for k, v := range cidrs {
		p, err := netip.ParsePrefix(k)
		if err != nil {
			log.Errorf("Error parsing, skipping %s with error %v", k, err)
			continue
		}
		r := table.NewRoute(p, v)
		//r = r.UpdateLabel(v)
		rtable.Add(r)
	}

	iterator := rtable.Iterate()

	for iterator.Next() {
		fmt.Printf("%v\n", iterator.Route())
		j, _ := iterator.Route().MarshalJSON()
		fmt.Printf("%v\n", string(j))
	}

	for _, h := range rtable.GetTable() {
		fmt.Println(h)
	}
	tst := netip.MustParsePrefix("10.0.1.0/23")
	fmt.Println(rtable.Children(tst), rtable.Parents(tst))

	rte := table.NewRoute(tst, nil)
	fmt.Println(rte.Children(rtable), rte.Parents(rtable))
}
