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
	filter.Add("health", "healthy")
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
		exec := container.ExecOptions{
			Cmd: []string{"kill", "1"},
			AttachStdout: false,
			AttachStderr: false,
		}
		resp, err := apiClient.ContainerExecCreate(context.Background(), ctr.ID, exec)
		if err != nil {
			log.Printf("[Error] %v", err)
			continue
		}
		if err = apiClient.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{}); err != nil {
			log.Printf("[Error] %v", err)
			continue
		}
		log.Printf("[Info] Killed the container %s", ctr.Names[0])
	}
}
