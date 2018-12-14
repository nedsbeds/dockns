package main

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/lextoumbourou/goodhosts"
)

func main() {

	cli, err := client.NewClientWithOpts(client.WithVersion("1.21"))
	if err != nil {
		panic(err)
	} else {
		fmt.Print("Watching for events \n")
	}

	events, errs, cancel := systemEventsSince(cli)

	defer cancel()

	for {
		select {
		case event := <-events:
			if event.Type == eventtypes.ContainerEventType {

				if event.Action == "stop" || event.Action == "start" {
					alterHosts(event.Action, cli, event.ID)
				}
			}
		case err := <-errs:
			if err == io.EOF {

			}

		}
		time.Sleep(100 * time.Millisecond)
	}
}

func alterHosts(action string, client client.APIClient, containerID string) {
	containerJSON, err := client.ContainerInspect(context.Background(), containerID)

	if err != nil {
		panic(err)
	}

	for label := range containerJSON.Config.Labels {
		if label == "traefik.frontend.rule" {

			hostnames := extractHostsFromLabel(containerJSON.Config.Labels[label])
			hostsFile, hostErr := goodhosts.NewHosts()

			for hostname := range hostnames {

				if hostErr != nil {
					fmt.Println("Could not parse hosts file")
				}

				if action == "start" {
					if !hostsFile.Has("127.0.0.1", hostnames[hostname]) {
						fmt.Printf("Adding %q \n", hostnames[hostname])
						hostsFile.Add("127.0.0.1", hostnames[hostname])
					}
				} else if action == "stop" {
					if hostsFile.Has("127.0.0.1", hostnames[hostname]) {
						fmt.Printf("Removing %q \n", hostnames[hostname])
						hostsFile.Remove("127.0.0.1", hostnames[hostname])
					}
				}
			}

			if flushErr := hostsFile.Flush(); err != nil {
				panic(flushErr)
			}
		}
	}

	if err != nil {
		panic(err)
	}

}

//grab a slice containing hostnames from the docker label
func extractHostsFromLabel(label string) []string {
	label = strings.Replace(label, "Host:", "", -1)
	label = strings.Replace(label, ",", " ", -1)

	return strings.Split(label, " ")
}

//starts listening for docker system events from now
func systemEventsSince(client client.APIClient) (<-chan eventtypes.Message, <-chan error, func()) {
	eventOptions := types.EventsOptions{
		Since: strconv.FormatInt(time.Now().Unix(), 10),
	}
	ctx, cancel := context.WithCancel(context.Background())
	events, errs := client.Events(ctx, eventOptions)

	return events, errs, cancel
}
