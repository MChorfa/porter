package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/namsral/flag"
)

const defaultTick = 60 * time.Second

type config struct {
	contentType string
	server      string
	statusCode  int
	tick        time.Duration
	url         string
	userAgent   string
}

func (c *config) init(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.String(flag.DefaultConfigFlagname, "", "Path to config file")

	var (
		statusCode = flags.Int("status", 200, "Response HTTP status code")
		tick       = flags.Duration("tick", defaultTick, "Ticking interval")
		server     = flags.String("server", "", "Server HTTP header value")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.statusCode = *statusCode
	c.tick = *tick
	c.server = *server

	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	c := &config{}

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					c.init(os.Args)
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()

	if err := run(ctx, c, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, c *config, stdout io.Writer) error {
	c.init(os.Args)
	log.SetOutput(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(c.tick):
			fmt.Printf("tick %d", c.tick)
			// _, err := porterdService.startServer(c.url)
			// if err != nil {
			// 	return err
			// }
		}
	}
}
