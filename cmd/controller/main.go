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

// gotta implement a config

func main() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	filter := filters.NewArgs()
	filter.Add("label", "tested=true")
	filter.Add("health", "healthy")
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	Outer:
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
		ctr := containers[seed.Intn(len(containers))]
		if ctr.Labels["controller"] == "true" {
			continue
		}
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
		log.Printf("[Info] Killing the container %s", ctr.Names[0])
		if err = apiClient.ContainerExecStart(context.Background(), resp.ID, container.ExecStartOptions{}); err != nil {
			log.Printf("[Error] %v", err)
			continue
		}
		var startTime time.Time
		var endTime time.Time
		killTime := time.Now()
		for {
			time.Sleep(500 * time.Millisecond)
			stateResp, err := apiClient.ContainerInspect(context.Background(), ctr.ID)
			if err != nil {
				log.Printf("[Error] %v", err)
				continue Outer
			}
			if stateResp.State.Health.Status != "healthy" {
				startTime = time.Now()
				break
			}
			if time.Since(killTime) >= (time.Second * 10) {
				log.Printf("[Error] Container %s didn't die.", ctr.Names[0])
				continue Outer
			}
		}
		for {
			time.Sleep(500 * time.Millisecond)
			stateResp, err := apiClient.ContainerInspect(context.Background(), ctr.ID)
			if err != nil {
				log.Printf("[Error] %v", err)
				continue Outer
			}
			if (stateResp.State.Health != nil && stateResp.State.Health.Status == "healthy" && stateResp.State.Running) || (stateResp.State.Health == nil && stateResp.State.Running == true) {
				endTime = time.Now()
				break
			}
			if time.Since(startTime) >= (time.Second * 30) {
				log.Printf("[Info] Container %s didn't restart.", ctr.Names[0])
				continue Outer
			}
		}
		log.Printf("[Info] Killed the container %s - it recovered in %v", ctr.Names[0], endTime.Sub(startTime))
	}
}
