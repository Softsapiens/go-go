// based on miekg/dns
package main

import (
	"log"
	"errors"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DnsResolver represents a dns resolver
type DnsResolver struct {
	Servers    []string
	RetryTimes int
	r          *rand.Rand
}

// New initializes DnsResolver.
func New(servers []string) *DnsResolver {
	for i := range servers {
		servers[i] = net.JoinHostPort(servers[i], "53")
	}

	return &DnsResolver{servers, len(servers) * 2, rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// NewFromResolvConf initializes DnsResolver from resolv.conf like file.
func NewFromResolvConf(path string) (*DnsResolver, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &DnsResolver{}, errors.New("no such file or directory: " + path)
	}
	config, err := dns.ClientConfigFromFile(path)
	servers := []string{}
	for _, ipAddress := range config.Servers {
		servers = append(servers, net.JoinHostPort(ipAddress, "53"))
	}
	return &DnsResolver{servers, len(servers) * 2, rand.New(rand.NewSource(time.Now().UnixNano()))}, err
}

// LookupHost returns IP addresses of provied host.
// In case of timeout retries query RetryTimes times.
func (r *DnsResolver) LookupHost(host string, recType uint16) ([]dns.RR, error) {
	return r.lookupHost(host, r.RetryTimes, recType)
}

func (r *DnsResolver) lookupHost(host string, triesLeft int, recType uint16) ([]dns.RR, error) {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{dns.Fqdn(host), recType, dns.ClassINET}
	in, err := dns.Exchange(m1, r.Servers[r.r.Intn(len(r.Servers))])

	result := []dns.RR{}

	if err != nil {
		if strings.HasSuffix(err.Error(), "i/o timeout") && triesLeft > 0 {
			triesLeft--
			log.Printf("Due to error [%s], retrying (tries left %s)", err.Error(), triesLeft)
			return r.lookupHost(host, triesLeft, recType)
		}
		return result, err
	}

	if in != nil && in.Rcode != dns.RcodeSuccess {
		return result, errors.New(dns.RcodeToString[in.Rcode])
	}

	return in.Answer, err
}


// LookupResult is Alias for channel type
type LookupResult struct {
	rr []dns.RR
	err error
}

func goLookup(ch chan LookupResult, f func(chan LookupResult)) {
	f(ch)
}

func sync(resolver *DnsResolver, domain string) {
	a, err := resolver.LookupHost(domain, dns.TypeA)
	if err != nil {
		log.Printf("Error %s", err)
	}
	aaaa, _ := resolver.LookupHost(domain, dns.TypeAAAA)
	if err != nil {
		log.Printf("Error %s", err)
	}
	mx, _ := resolver.LookupHost(domain, dns.TypeMX)
	if err != nil {
		log.Printf("Error %s", err)
	}
	ns, _ := resolver.LookupHost(domain, dns.TypeNS)
	if err != nil {
		log.Printf("Error %s", err)
	}
	
	log.Printf("domain %s --> \nA: \n%s\nAAAA: \n%s\nMX: \n%s\nNS: \n%s\n", domain, a, aaaa, mx, ns)
}

func async(resolver *DnsResolver, domain string) {
	chA := make(chan LookupResult, 1)
	chAAAA := make(chan LookupResult, 1)
	chMX := make(chan LookupResult, 1)
	chNS := make(chan LookupResult, 1)

	lkup := func(t uint16, d string) func(chan LookupResult) {
		return func(ch chan LookupResult) {
			r, err := resolver.LookupHost(d, t)
			if err != nil {
				log.Printf("Error %s", err)
			}
			ch<-LookupResult{r, err}
		}
	}

	go goLookup(chA, lkup(dns.TypeA, domain))
	go goLookup(chAAAA, lkup(dns.TypeAAAA, domain))
	go goLookup(chMX, lkup(dns.TypeMX, domain))
	go goLookup(chNS, lkup(dns.TypeNS, domain))

	a := <-chA
	aaaa := <-chAAAA
	mx := <-chMX
	ns := <-chNS

	log.Printf("domain %s --> \nA: \n%s\nAAAA: \n%s\nMX: \n%s\nNS: \n%s\n", domain, a, aaaa, mx, ns)
}

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <domain> [s]\n",
			os.Args[0])
		os.Exit(1)
	}

	domain := os.Args[1]
	synky := os.Args[2]

	resolver := New([]string{"8.8.8.8", "8.8.4.4"})
	// OR
	// resolver := dns_resolver.NewFromResolvConf("resolv.conf")

	// In case of i/o timeout
	resolver.RetryTimes = 5

	if synky == "s" {
		sync(resolver, domain)
	} else {
		async(resolver, domain)
	}
}