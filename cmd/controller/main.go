package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func main() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	filter := filters.NewArgs()
	filter.Add("label", "tested=true")
	for {
		time.Sleep(time.Second * 5)
		containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: false, Filters: filter})
		if err != nil {
			log.Printf("[Error] %v", err)
			continue
		}
		if len(containers) == 0 {
			log.Printf("[Info] No containers up.")
			continue
		}
		ctr := containers[rand.Intn(len(containers))]
		if err = apiClient.ContainerKill(context.Background(), ctr.ID, "SIGKILL"); err != nil {
			log.Printf("Couldn't kill the container %v", ctr)
		}
	}
}
