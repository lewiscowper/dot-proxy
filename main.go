package main

import (
  "log"
  "net"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/miekg/dns"
)

func shutdownServer(s *dns.Server) {
  err := s.Shutdown()

  if err != nil {
    // log with fatal level here to really kill everything in case we have an error
    log.Fatal("Failed to shutdown server %s", s.Net)
  }
}

func createHandler(client *dns.Client, conn *dns.Conn) func (dns.ResponseWriter, *dns.Msg) {
  return func (w dns.ResponseWriter, m *dns.Msg) {
    msgString := ""
    for _, q := range m.Question {
      msgString += q.String()
    }
    log.Printf("Received query: '%s'", msgString)
    response, rtt, err := client.ExchangeWithConn(m, conn)
    if err != nil {
      log.Fatalf("Can't reach upstream\n%s", err)
      return
    }
    log.Printf("Response Time: '%s'", rtt.String())
    w.WriteMsg(response)
  }
}

func main() {
  signalChan := make(chan os.Signal, 1)
  signal.Notify(signalChan, syscall.SIGTERM)
  signal.Notify(signalChan, syscall.SIGINT)

  c := new(dns.Client)
  c.Net = "tcp-tls"
  c.Dialer = &net.Dialer{
    Timeout: 200 * time.Millisecond,
  }

  conn, err := c.Dial("1.1.1.1:853")
  if err != nil {
    log.Fatal(err)
  }

  tcpServer := &dns.Server{Addr: ":2410", Net: "tcp"}

  dnsHandler := createHandler(c, conn)

  go tcpServer.ListenAndServe()
  dns.Handle(".", dns.HandlerFunc(dnsHandler))
  log.Println("Now listening")

  sig := <-signalChan
  log.Printf("Received signal: %q, shutting down..", sig.String())
  shutdownServer(tcpServer)
}
