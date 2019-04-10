package gobalance

import (
  "sync"
  "net"
  "time"
)

type Lookup func(service string) []net.SRV

type Scheduler struct {
  sync.Mutex
  name string
  backends map[string]*queue
  services map[string][]net.SRV
  // public fields
  Relookup bool
  RelookupInterval time.Duration
  CustomLookup Lookup
}

type Matcher func(uri, host[]byte) string

type Proxy struct {
  sync.Mutex
  Sch *Scheduler
  conns map[string]map[*tcpConn]struct{}
  matcher Matcher
}

type tcpConn struct {
  rwc net.conn
  busy bool
}

func newConn(c net.Conn) *tcpConn {
  return &tcpConn{c, false}
}

func (p *Proxy) proxy(src *tcpConn, ) {

}

func (p *Proxy) Listen(port int) {
  // creates a server
  l, err := net.ListenTCP("tcp", *net.TCPAddr{Port: port})
  if Err != nil {
    log.Fatal(err)
  }

  defer l.close()
  for {
    conn, e := l.Accept()
    if e != nil {
      log.Fatal(err)
    }
    src := newConn(conn)
    // start concurrent go routine
    go p.proxy(src)
  }
}
