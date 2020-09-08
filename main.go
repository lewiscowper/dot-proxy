package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/miekg/dns"
)

type config struct {
	UpstreamHost string        `split_words:"true" default:"1.1.1.1"`
	UpstreamPort string        `split_words:"true" default:"853"`
	ListenPort   string        `split_words:"true" default:"53"`
	Timeout      time.Duration `default:"500ms"`
}

func main() {
	var c config
	err := envconfig.Process("dotproxy", &c)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGINT)

	client := new(dns.Client)
	client.Net = "tcp-tls"
	client.Dialer = &net.Dialer{
		Timeout: c.Timeout,
	}

	conn, err := client.Dial(fmt.Sprintf("%s:%s", c.UpstreamHost, c.UpstreamPort))
	if err != nil {
		log.Fatal(err)
	}

	listenAddr := fmt.Sprintf(":%s", c.ListenPort)
	tcpServer := &dns.Server{Addr: listenAddr, Net: "tcp"}
	udpServer := &dns.Server{Addr: listenAddr, Net: "udp"}

	dnsHandler := createHandler(client, conn)

	go tcpServer.ListenAndServe()
	go udpServer.ListenAndServe()
	dns.Handle(".", dns.HandlerFunc(dnsHandler))
	log.Println("Now listening")

	sig := <-signalChan
	log.Printf("Received signal: %q, shutting down..", sig.String())
	shutdownServer(tcpServer)
}

func createHandler(client *dns.Client, conn *dns.Conn) func(dns.ResponseWriter, *dns.Msg) {
	return func(w dns.ResponseWriter, m *dns.Msg) {
		msgString := ""
		// m.Question holds the actual queries in the dns message datagram
		for _, q := range m.Question {
			msgString += q.String()
		}
		log.Printf("Received query: '%s'", msgString)
		// By reusing the connection we sacrifice some reliability (if the connection dies we die),
		// for the sake of speed of already having done the TLS negotiation.
		// If more reliability was sought, swapping for client.Exchange would be a good way to not
		// cause a pod restart (in the kubernetes case), but the restart would likely be quick, and
		// would potentially make up in the time saved from doing TLS negotiation.
		response, rtt, err := client.ExchangeWithConn(m, conn)
		if err != nil {
			log.Fatalf("Can't reach upstream\n%s", err)
			return
		}
		log.Printf("Response Time: '%s'", rtt.String())
		w.WriteMsg(response)
	}
}

func shutdownServer(s *dns.Server) {
	err := s.Shutdown()

	if err != nil {
		// log with fatal level here to really kill everything in case we have an error
		log.Fatal("Failed to shutdown server %s", s.Net)
	}
}
