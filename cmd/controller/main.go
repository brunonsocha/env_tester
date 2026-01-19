package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/goccy/go-yaml"
)

var data = [][]string{{"Iteration", "Timestamp", "Container ID", "Container Name", "Recovery Time", "State"}}

type Config struct {
	Kills struct {
		Target_label string `yaml:"target_label"`
		Kill_interval int `yaml:"kill_interval"`
		Max_iterations int `yaml:"max_iterations"`
		Death_timeout int `yaml:"death_timeout"`
		Recovery_timeout int `yaml:"recovery_timeout"`
		Signal string `yaml:"signal"`
	} `yaml:"kills"`
}

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		log.Fatal("[Error] No config found.")
	}
	f, err := os.Open(cfgPath)
	if err != nil {
		log.Fatalf("[Error] %v", err)
	}
	defer f.Close()
	cfg, err := loadCfg(f)
	if err != nil {
		log.Fatalf("[Error] %v", err)
	}
	log.Printf("[Info] Config loaded.")
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()
	filter := filters.NewArgs()
	filter.Add("label", cfg.Kills.Target_label)
	filter.Add("health", "healthy")
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	success := 0
	totalTime := time.Duration(0)
	maxTime := time.Duration(0)
	Outer:
	for i := 0; i < cfg.Kills.Max_iterations; i++ {
		time.Sleep(time.Second * time.Duration(cfg.Kills.Kill_interval))
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
			Cmd: []string{"kill", "-s", cfg.Kills.Signal, "-1"},
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
				row := []string{
					fmt.Sprintf("%d", i+1),
					time.Now().Format(time.RFC3339),
					ctr.ID,
					ctr.Names[0],
					"",
					"FAILED",
				}
				data = append(data, row)
				continue Outer
			}
			if stateResp.State.Health.Status != "healthy" {
				startTime = time.Now()
				break
			}
			if time.Since(killTime) >= (time.Second * time.Duration(cfg.Kills.Death_timeout)) {
				log.Printf("[Error] Container %s didn't die.", ctr.Names[0])
				row := []string{
					fmt.Sprintf("%d", i+1),
					time.Now().Format(time.RFC3339),
					ctr.ID,
					ctr.Names[0],
					"",
					"FAILED",
				}
				data = append(data, row)
				continue Outer
			}
		}
		for {
			time.Sleep(500 * time.Millisecond)
			stateResp, err := apiClient.ContainerInspect(context.Background(), ctr.ID)
			if err != nil {
				log.Printf("[Error] %v", err)
				row := []string{
					fmt.Sprintf("%d", i+1),
					time.Now().Format(time.RFC3339),
					ctr.ID,
					ctr.Names[0],
					"",
					"FAILED",
				}
				data = append(data, row)
				continue Outer
			}
			if (stateResp.State.Health != nil && stateResp.State.Health.Status == "healthy" && stateResp.State.Running) || (stateResp.State.Health == nil && stateResp.State.Running == true) {
				endTime = time.Now()
				break
			}
			if time.Since(startTime) >= (time.Second * time.Duration(cfg.Kills.Recovery_timeout)) {
				log.Printf("[Info] Container %s didn't restart.", ctr.Names[0])
				row := []string{
					fmt.Sprintf("%d", i+1),
					time.Now().Format(time.RFC3339),
					ctr.ID,
					ctr.Names[0],
					"",
					"FAILED",
				}
				data = append(data, row)
				continue Outer
			}
		}
		timeSpent := endTime.Sub(startTime)
		log.Printf("[Info] Container %s recovered in %v", ctr.Names[0], timeSpent)
		row := []string{
			fmt.Sprintf("%d", i+1),
			time.Now().Format(time.RFC3339),
			ctr.ID,
			ctr.Names[0],
			fmt.Sprintf("%.4f", timeSpent.Seconds()),
			"RECOVERED",
		}
		data = append(data, row)
		if timeSpent > maxTime {
			maxTime = timeSpent
		}
		success++
		totalTime += timeSpent
	}
	avgTime := time.Duration(0)
	if success > 0 {
		avgTime = totalTime / time.Duration(success)
	}
	log.Printf("[Info] Finished testing after %d kills.\n[STATS]\nSuccess: %d/%d\nAverage time: %v\nMax time: %v\nKill method: %s", cfg.Kills.Max_iterations, success, cfg.Kills.Max_iterations, avgTime, maxTime, cfg.Kills.Signal)
}

func loadCfg(f io.Reader) (*Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("Incorrect config structure.")
	}
	return &cfg, nil
}
