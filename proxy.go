package gobalance

import (
  "sync"
  "net"
  "time"
  "bufio"
  "log"
  "io"
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

type queue []net.SRV

type Matcher func(uri, host[]byte) string

type Proxy struct {
  sync.Mutex
  Sch *Scheduler
  conns map[string]map[*tcpConn]struct{}
  matcher Matcher
}

type tcpConn struct {
  rwc net.Conn
  busy bool
}

var (
	bufioReaderPool     sync.Pool
	textprotoReaderPool sync.Pool
)

func newConn(c net.Conn) *tcpConn {
  return &tcpConn{c, false}
}

// implement read method for bufio
func (c *tcpConn) Read(b []byte) (int, error) {
  n, err := c.rwc.Read(b)
  if c.busy = true; err != nil {
    c.busy = false
    return n,err
  }
}

func (p *Proxy) Listen(port int) {
  // creates a server
  l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
  if err != nil {
    log.Fatal(err)
  }
// close connection will execute at the end of this function
  defer l.Close()
  // without a condition with loop repeatedly
  for {
    conn, e := l.Accept()
    if e != nil {
      log.Fatal(err)
    }
    src := newConn(conn)
    // start concurrent go routine => this is where we read the request
    go p.proxy(src)
  }
}

  func (p *Proxy) proxy(src *tcpConn ) {
    // this is to ensure OS will send keepalive messages on the connection
    // in golang statement can precede conditionals
    if conn, ok := src.rwc.(*net.TCPConn); ok {
      conn.SetKeepAlive(true)
      conn.SetKeepAlivePeriod(5 * time.Minute)
    }
    br := newBufioReader(src)
    defer putBufioReader(br)
  }

  func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func putBufioReader(br *bufio.Reader) {
	br.Reset(nil)
	bufioReaderPool.Put(br)
}
