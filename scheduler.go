package gobalance

import (
  "sync"
  "net"
)
// SRV represents a single DNS SRV record
//type SRV struct {
      //  Target   string
      //  Port     uint16
      //  Priority uint16
      //  Weight   uint16
//}
type Lookup func(service string) []net.SRV

type Scheduler struct {
  sync.Mutex
  name string
  backends map[string]*queue
  services map[string][]net.SRV
  Relookup bool
  RelookupInterval time.Duration
  CustomLookup Lookup
}

type queue []net.SRV

func NewScheduler(relookup bool, interval time.Duration, custom Lookup) *Scheduler {
  s := Scheduler{
    backends : make(map[string]*queue),
    services: make(map[string][]net.SRV),
    Relookup: relookup,
    RelookupInterval: interval,
    CustomLookup: custom,
  }

  if relookup {
    // go s.relookupEvery(interval)
  }

  return &s
}

func (s *Scheduler) NextBackend(service string) net.SRV {
  q, ok := s.getQueue(service)
  if !ok {
    s.lookup(service)
  }
}

func (s *Scheduler) getQueue(service string) (q *queue, ok bool) {
  s.Lock()
  // ensures mutual exclusion and only one goroutine can access this map
  q, ok = s.backends[service]
  s.Unlock()
  return
}

func (s *Scheduler) lookup(service string) error {
  var records []net.SRV
  if s.CustomLookup != nil {
    records = s.CustomLookup(service)
  } else {
    _, addrs, err := net.LookupSRV(service, "tcp", s.name)
    if err != nil {
      return err
    }

    records = make([]net.SRV, len(addrs))
    for i := 0; i < len(addrs); i++ {
      records[i] = *addrs[i]
    }
  }
  s.Lock()
  s.services[service] = records
  s.Unlock()
  return nil
}
