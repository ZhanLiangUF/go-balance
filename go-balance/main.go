package main

import (
  "bytes"
  "fmt"
  "sync"
  "net"
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
  // When getting from a Pool, you need to cast
   s := pool.Get().(*bytes.Buffer)
   // We write to the object
   s.Write([]byte("dirty"))
   // Then put it back
   pool.Put(s)

   // Pools can return dirty results

   // Get 'another' buffer
   s = pool.Get().(*bytes.Buffer)
   // Write to it
   s.Write([]byte("append"))
   // At this point, if GC ran, this buffer *might* exist already, in
   // which case it will contain the bytes of the string "dirtyappend"
   fmt.Println(s)

   b := bytes.NewBuffer([]byte("dirtyasdfasdfasdf boost yeah"))
   // b.ReadBytes(' ')
   // b.ReadBytes(' ')
   url, _ := b.ReadBytes('\n')
   l1 := append(url, byte('\r'), byte('\n'))
   // fmt.Println(url)
   // fmt.Println(string(url))
   fmt.Println(string(l1))
}
