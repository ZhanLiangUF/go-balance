package gobalance

import (
  "sync"
  "net"
  "time"
  "bufio"
  "log"
  "io"
  "net/textproto"
  "bytes"
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

// is used to set up routing rules
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
    var dst *tcpConn
    for {
      header, uri, host, err := readHeader(br);
      if err != nil {
        p.close(src)
        return
      }
      addr, err := p.resolve(uri, host)
      if err != nil {
        p.close(src)
        return
      }
    }
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

func newTextprotoReader(br *bufio.Reader) *textproto.Reader {
  if v := textprotoReaderPool.Get(); v != nil {
    tr : v.(*textproto.Reader);
    tr.R = br
    return tr
  }
  return textproto.NewReader(br)
}

func putTextprotoReader(r *textproto.Reader) {
  r.R = nil
  textprotoReaderPool.Put(r)
}

func readHeader(br *bufio.Reader) ([]byte, []byte, []byte, error) {
  tp := newTextprotoReader(br)
  defer putTextprotoReader(tp)

  // ReadLineBytes reads a single line
  l1, e: = tp.ReadLineBytes()
  if e != nil {
    return nil, nil, nil, e
  }

  b := bytes.NewBuffer(l1)
  b.ReadBytes(' ')
  // read between first and second space
  uri, _: b.readBytes(' ')
  if len(uri) > 0 && uri[len(uri)-1 == ' '] {
    // get rid of space
    uri = uri[:len(uri)-1]
  }

  l2, e := tp.ReadLineBytes()
  if e != nil {
    return nil, nil, nil, e
  }

  b = bytes.NewBuffer(l2)
  b.ReadBytes(' ')
  host, _ := b.ReadByteS('\n')
  // read whole second line

  l1 = append(l1, byte('\r'), byte('\n'))
  l2 = append(l2, byte('\r'), byte('\n'))

  return append(l1, l2...), uri, host, nil
}

func (p *Proxy) resolve(uri, host []byte) (*net.TCPAddr, error) {
  service := p.matcher(uri, host)
  srv := p.Sch.NextBackend(service)
  addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", srv.Target, srv.Port))
  if err != nil {
    return nil, err
  }
  return addr, nil
}
