package main

import (
  "net"
  balance "github.com/zhanlianguf/go-balance"
  "sync"
  "bytes"
)
type queue []net.SRV

var pool = sync.Pool{
    // New creates an object when the pool has nothing available to return.
    // New must return an interface{} to make it flexible. You have to cast
    // your type after getting it.
    New: func() interface{} {
        // Pools often contain things like *bytes.Buffer, which are
        // temporary and re-usable.
        return &bytes.Buffer{}
    },
}
func main() {

	customLookup := func(service string) []net.SRV {
		return []net.SRV{
			{Target: "whoami1.local", Port: 32768, Weight: 50},
			{Target: "whoami2.local", Port: 32769, Weight: 50},
		}
	}

	matcher := func(uri, host []byte) string {
		return "test.traefik"
	}

	sc := balance.NewScheduler(false, 0, customLookup)
	p := balance.NewProxy(matcher)
	p.Sch = sc

	p.Listen(8090)
}
