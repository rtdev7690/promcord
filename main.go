// Application which greets you.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rich7690/promcord/internal/discord"
	"github.com/rich7690/promcord/internal/server"
)

var (
	commit  string
	version string
)

var (
	DiscordToken   string = os.Getenv("DISCORD_TOKEN")
	PerspectiveKey string = os.Getenv("PERSPECTIVE_KEY")
	StopDelay      *int = flag.Int("stop-delay", 5, "amount of time to wait before exiting. Giving time for connections to drain")
)

func main() {
	flag.Parse()

	log.Println("Starting Server. Version: ", version, " Commit: ", commit)
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)

	go func() {
		err := server.StartServer(ctx)
		if err != nil && err == http.ErrServerClosed {
			log.Printf("Error: %v\n", err)
		}
	}()

	d, err := discord.StartBot(ctx, DiscordToken, PerspectiveKey)
	if err != nil {
		log.Fatal(err)
	}
	err = d.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()

	<-sigs
	cancel()
	log.Println("Exiting")
}
