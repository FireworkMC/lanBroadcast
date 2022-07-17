package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	lan "github.com/FireworkMC/lanbroadcast"
)

func main() {
	var port int
	var motd string
	var interval time.Duration
	var addr string

	flag.IntVar(&port, "port", 25565, "The port of the server")
	flag.StringVar(&motd, "motd", "", "The MOTD to display")
	flag.DurationVar(&interval, "interval", time.Second*5, "The interval between broadcasts")
	flag.StringVar(&addr, "address", "", "The address of the host network. This is automatically detected")
	flag.Parse()

	var ip net.IP
	if addr != "" {
		ip = net.ParseIP(addr)
		if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
			log.Fatal("The provided IP is not valid.")
		}
	}
	l, err := lan.New(ip, port, motd)
	if err != nil {
		log.Fatal(err)
	}
	l.SetInterval(uint(interval / time.Second))
	done := make(chan bool)

	go func() {
		fmt.Println("Started broadcasting to LAN")
		l.Broadcast(context.Background())
		close(done)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("\rReceived Ctrl+C exiting")
	l.Close()
	<-done
}
