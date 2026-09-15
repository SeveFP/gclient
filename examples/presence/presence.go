package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/jech/gclient"
)

func main() {
	var username, password string
	var insecure bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: %s group\n", os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.StringVar(&username, "username", "presence-example",
		"`username` to use for login")
	flag.StringVar(&password, "password", "",
		"`password` to use for login")
	flag.BoolVar(&insecure, "insecure", false,
		"don't check server certificates")
	flag.BoolVar(&gclient.Debug, "debug", false,
		"enable protocol logging")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	client := gclient.NewClient()

	if insecure {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.SetHTTPClient(&http.Client{
			Transport: t,
		})

		d := *websocket.DefaultDialer
		d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.SetDialer(&d)
	}

	ctx := context.Background()
	if err := client.Connect(ctx, flag.Arg(0)); err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	if err := client.Join(ctx, flag.Arg(0), username, password); err != nil {
		log.Fatalf("join: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	streams := make(map[string]string)

	for {
		select {
		case <-stop:
			fmt.Println("stopping")
			return
		case event, ok := <-client.EventCh:
			if !ok {
				fmt.Println("event channel closed")
				return
			}
			if event == nil {
				return
			}

			switch event := event.(type) {
			case gclient.JoinedEvent:
				if event.Kind == "fail" {
					log.Printf("join failed: %s", event.Value)
					return
				}
				if event.Kind == "join" {
					if err := client.Request(map[string][]string{
						"": {"audio", "video"},
					}); err != nil {
						log.Printf("request streams: %v", err)
					}
				}
			case gclient.DownConnEvent:
				var stopMessage string
				switch event.Label {
				case "camera":
					fmt.Printf("%s started sharing: Camera\n", event.Username)
					stopMessage = fmt.Sprintf("%s stopped sharing: Camera", event.Username)
				case "screenshare":
					fmt.Printf("%s started sharing: Screen\n", event.Username)
					stopMessage = fmt.Sprintf("%s stopped sharing: Screen", event.Username)
				default:
					fmt.Printf("%s started streaming\n", event.Username)
					stopMessage = fmt.Sprintf("%s stopped streaming", event.Username)
				}
				streams[event.Id] = stopMessage
			case gclient.CloseEvent:
				if stopMessage, ok := streams[event.Id]; ok {
					fmt.Println(stopMessage)
					delete(streams, event.Id)
				}
			case error:
				log.Printf("client error: %v", event)
				return
			default:
				fmt.Printf("event: %T\n", event)
			}
		}
	}
}
