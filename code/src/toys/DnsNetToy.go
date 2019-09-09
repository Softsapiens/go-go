package main

import (
	"net"
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <domain>\n",
			os.Args[0])
		os.Exit(1)
	}

	domain := os.Args[1]

	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		fmt.Printf("%s IN A %s\n", domain, ip.String())
	}

	// Exists more net.Lookup***
}