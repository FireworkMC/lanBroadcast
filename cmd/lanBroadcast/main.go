package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	lanbroadcast "github.com/FireworkMC/lanBroadcast"
	"github.com/spf13/cobra"
)

func main() {
	var port int
	var motd string
	cmd := cobra.Command{Run: func(cmd *cobra.Command, args []string) {
		run(motd, port)
	}}
	flags := cmd.Flags()
	flags.IntVar(&port, "port", 25565, "The port of the server")
	flags.StringVar(&motd, "motd", "", "The MOTD to display")
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}

}
func run(motd string, port int) {
	l, err := lanbroadcast.NewLANBroadcast(context.TODO(), nil, port, motd)
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)

	go func() {
		l.Broadcast()
		close(done)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	l.Close()
	<-done
}
