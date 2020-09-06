package main

import (
  "fmt"
  "io"
  "net"
  "log"
)

func handleConnection(downstream net.Conn, upstream net.Conn) {
  fmt.Printf("%+v\n", downstream)
  buf := make([]byte, 0, 512)
  // tmp buffer for in progress reads
  tmp := make([]byte, 256)
  for {
    n, err := downstream.Read(tmp)
    if err != nil {
      if err != io.EOF {
        fmt.Printf("read error: %s\n", err)
      }
      break
    }
    fmt.Printf("read %d bytes.\n", n)
    buf = append(buf, tmp[:n]...)
    n, err = upstream.Write(tmp[:n])
    if err != nil {
      fmt.Printf("write error: %s\n", err)
      return
    }
    fmt.Printf("wrote %d bytes.\n", n)
  }
}

func main() {
  downstream, err := net.Listen("tcp", "localhost:2410")
  if err != nil {
    log.Fatal(err)
  }

  upstream, err := net.Dial("tcp", "localhost:2411")
  if err != nil {
    log.Fatal(err)
  }
  defer upstream.Close()

  fmt.Println("Now listening")
  for {
    conn, err := downstream.Accept()
    if err != nil {
      log.Fatal(err)
    }
    defer conn.Close()
    go handleConnection(conn, upstream)
  }
}
