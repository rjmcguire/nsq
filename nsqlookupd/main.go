package main

import (
	"../nsq"
	"../util"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var (
	showVersion             = flag.Bool("version", false, "print version string")
	tcpAddress              = flag.String("tcp-address", "0.0.0.0:4160", "<addr>:<port> to listen on for TCP clients")
	httpAddress             = flag.String("http-address", "0.0.0.0:4161", "<addr>:<port> to listen on for HTTP clients")
	verbose                 = flag.Bool("verbose", false, "enable verbose logging")
	inactiveProducerTimeout = flag.Duration("inactive-producer-timeout", 300*time.Second, "duration of time a producer will remain in the active list since its last ping")
	tombstoneLifetime       = flag.Duration("tombstone-lifetime", 45*time.Second, "duration of time a producer will remain tombstoned if registration remains")
	broadcastAddress        = flag.String("broadcast-address", "", "address of this lookupd node, (default to the OS hostname)")
)

var protocols = map[string]nsq.Protocol{}
var lookupd *NSQLookupd

func main() {
	flag.Parse()

	if *broadcastAddress == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal(err)
		}
		*broadcastAddress = hostname
	}

	if *showVersion {
		fmt.Printf("nsqlookupd v%s\n", util.BINARY_VERSION)
		return
	}

	signalChan := make(chan os.Signal, 1)
	exitChan := make(chan int)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, os.Interrupt)

	tcpAddr, err := net.ResolveTCPAddr("tcp", *tcpAddress)
	if err != nil {
		log.Fatal(err)
	}

	httpAddr, err := net.ResolveTCPAddr("tcp", *httpAddress)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("nsqlookupd v%s", util.BINARY_VERSION)

	lookupd = NewNSQLookupd()
	lookupd.tcpAddr = tcpAddr
	lookupd.httpAddr = httpAddr
	lookupd.broadcastAddress = *broadcastAddress
	lookupd.inactiveProducerTimeout = *inactiveProducerTimeout
	lookupd.tombstoneLifetime = *tombstoneLifetime
	lookupd.Main()
	<-exitChan
	lookupd.Exit()
}
