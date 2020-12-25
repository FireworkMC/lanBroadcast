package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	lanbroadcast "github.com/FireworkMC/lanbroadcast"
	"github.com/spf13/cobra"
)

func main() {
	var port int
	var motd string
	var interval time.Duration
	var addr string
	cmd := cobra.Command{
		Run: func(cmd *cobra.Command, args []string) { run(motd, port, interval, addr) },
		Use: "lanBroadcast [--port] [--motd] [--interval] [--address]",
	}
	flags := cmd.Flags()
	flags.IntVar(&port, "port", 25565, "The port of the server")
	flags.StringVar(&motd, "motd", "", "The MOTD to display")
	flags.DurationVar(&interval, "interval", time.Second*5, "The interval between broadcasts")
	flags.StringVar(&addr, "address", "", "The address of the host network. This is automatically detected")
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
func run(motd string, port int, interval time.Duration, addr string) {
	var ip net.IP
	if addr != "" {
		ip = net.ParseIP(addr)
		if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
			log.Fatal("The provided IP is not valid.")
		}
	}
	l, err := lanbroadcast.NewLANBroadcast(context.TODO(), ip, port, motd)
	if err != nil {
		log.Fatal(err)
	}
	l.SetInterval(uint(interval / time.Second))
	done := make(chan bool)

	go func() {
		fmt.Println("Started broadcasting to LAN")
		l.Broadcast()
		close(done)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Println("Received Ctrl+C exiting")
	l.Close()
	<-done
}
