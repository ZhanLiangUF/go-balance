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
    // make is use to allocate and initialize
    backends : make(map[string]*queue),
    services: make(map[string][]net.SRV),
    Relookup: relookup,
    RelookupInterval: interval,
    CustomLookup: custom,
  }

  if relookup {
    // update SRV periodically
    go s.relookupEvery(interval)
  }

  return &s
}

func (s *Scheduler) getSRVs(service string) (backends []net.SRV, ok bool) {
  s.Lock()
  backends, ok = s.services[service]
  s.Unlock()
  return
}

func (s *Scheduler) relookupEvery(d time.Duration) {
  ticker := time.NewTicker(d)
  defer ticker.stop()
  for {
    select{
    case <-ticker.C:
      s.Lock()
      services := make([]string, len(s.services))
      i := 0
      for k := range s.services {
        services[i] = k
        i++
      }
      s.Unlock()
      for _, service := range services {
        go s.lookup(service)
      }
    }
  }
}

func (s *Scheduler) NextBackend(service string) net.SRV {
  q, ok := s.getQueue(service)
  if !ok {
    s.lookup(service)
  }

  if q == nil || q.Len() == 0 {
    s.requeue(service)
  }

  return s.pop(service)
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

func (s *Scheduler) requeue(service string) {
  records, _ := s.getSRVs(service)
  nRecords := len(records)

  if nRecords == 0 {
    return
  }

  total := uint(0)
  for _, val := range records {
    total += uint(val.Weight)
  }

  unordered := make([]int, nREcords)
  for i, val := range records{
    pct := 1.0
    if total != 0 {
      pct = float64(val.Weight) / float64(total) * 10
    }
    unordered[i] = int(pct)
  }

  ordered := append(unordered[:0:0], unordered...)
  sort.Ints(ordered)

  q := queue{}
  max := ordered[nRecords-1]
  for rep := 1; rep <= max; rep++ {
    for index := 0; index < nRecords; index++ {
      if unordered[index]-rep >= 0 {
        q = append(q, records[index])
      }
    }
  }

  ptr := &q
  heap.Init(ptr)
  s.Lock()
  s.backends[service] = ptr
  s.Unlock()
}
