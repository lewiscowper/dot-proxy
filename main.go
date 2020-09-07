package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

func shutdownServer(s *dns.Server) {
	err := s.Shutdown()

	if err != nil {
		// log with fatal level here to really kill everything in case we have an error
		log.Fatal("Failed to shutdown server %s", s.Net)
	}
}

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, syscall.SIGINT)

	tcpServer := &dns.Server{Addr: ":2410", Net: "tcp"}

	go tcpServer.ListenAndServe()
	log.Println("Now listening")

	sig := <-signalChan
	log.Printf("Received signal: %q, shutting down..", sig.String())
	shutdownServer(tcpServer)
}
