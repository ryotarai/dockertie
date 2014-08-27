package main

import (
	"log"
	"strconv"
	"github.com/fsouza/go-dockerclient"
	"github.com/codegangsta/cli"
	"strings"
	"sync"
)

type Container struct {
	Id string
	Name string
	Path string
	Args []string
	Env map[string]string
	Host Host
}

type Containerizer interface {
	GetContainersOnHost(host Host) ([]Container, error)
	GetContainersOnHosts(hosts []Host) ([]Container, error)
}

func NewContainerizer(name string, c *cli.Context) Containerizer {
	switch name {
	case "docker":
		containerizer := NewDockerContainerizer(c)
		return containerizer
	}

	return nil
}

type DockerContainerizer struct {
	DefaultPort int
}

func NewDockerContainerizer(c *cli.Context) DockerContainerizer {
	return DockerContainerizer{
		DefaultPort: c.Int("docker-http-port"),
	}
}

func (c DockerContainerizer) getClient(host Host) (*docker.Client, error) {
	port := c.DefaultPort
	endpoint := "tcp://" + host.Addr + ":" + strconv.Itoa(port)
	log.Printf("Docker endpoint: %s", endpoint)
	return docker.NewClient(endpoint)
}

func (c DockerContainerizer) GetContainersOnHost(host Host) ([]Container, error) {
	client, err := c.getClient(host)
	if (err != nil) {
		return nil, err
	}

	log.Println("Listing containers...")
	dockerContainers, err := client.ListContainers(
		docker.ListContainersOptions{},
	)
	if (err != nil) {
		return nil, err
	}

	containers := []Container{}
	for _, dockerContainer := range dockerContainers {
		inspection, err := client.InspectContainer(dockerContainer.ID)
		if (err != nil) {
			return nil, err
		}

		env := map[string]string{}
		for _, envStr := range inspection.Config.Env {
			keyValue := strings.SplitN(envStr, "=", 2)
			env[keyValue[0]] = keyValue[1]
		}

		container := Container{
			Id: inspection.ID,
			Name: inspection.Name,
			Path: inspection.Path,
			Args: inspection.Args,
			Env: env,
			Host: host,
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func (c DockerContainerizer) GetContainersOnHosts(hosts []Host) ([]Container, error) {
	containers := []Container{}

	receiver := make(chan []Container)
	fin := make(chan bool)

	go func() {
		var wg sync.WaitGroup
		for _, host := range hosts {
			wg.Add(1)
			go func() {
				c, err := c.GetContainersOnHost(host)
				if (err != nil) {
					log.Println(err)
					return
				}
				receiver <- c
				wg.Done()
			}()
		}
		wg.Wait()
		fin <- true
	}()

	for {
		select {
		case receive := <-receiver:
			containers = append(containers, receive...)
		case <-fin:
			return containers, nil
		}
	}
}

