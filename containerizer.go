package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

type Container struct {
	Id   string
	Name string
	Path string
	Args []string
	Env  map[string]string
	Host Host
}

type ContainerConfig struct {
	Image string
	Cmd   []string
	Env   []string
}

type Containerizer interface {
	GetContainersOnHost(Host) ([]Container, error)
	GetContainersOnHosts([]Host) ([]Container, error)
	CreateContainer(Host, ContainerConfig) (*docker.Container, error)
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
	var (
		port int
		err error
	)
	if s, ok := host.ContainerizerInfo["DockerPort"]; ok {
		port, err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	} else {
		port = c.DefaultPort
	}
	endpoint := "tcp://" + host.Addr + ":" + strconv.Itoa(port)
	log.Printf("Docker endpoint: %s", endpoint)
	return docker.NewClient(endpoint)
}

func (c DockerContainerizer) GetContainersOnHost(host Host) ([]Container, error) {
	client, err := c.getClient(host)
	if err != nil {
		return nil, err
	}

	log.Println("Listing containers...")
	dockerContainers, err := client.ListContainers(
		docker.ListContainersOptions{},
	)
	if err != nil {
		return nil, err
	}

	containers := []Container{}
	for _, dockerContainer := range dockerContainers {
		inspection, err := client.InspectContainer(dockerContainer.ID)
		if err != nil {
			return nil, err
		}

		env := map[string]string{}
		for _, envStr := range inspection.Config.Env {
			keyValue := strings.SplitN(envStr, "=", 2)
			env[keyValue[0]] = keyValue[1]
		}

		container := Container{
			Id:   inspection.ID,
			Name: inspection.Name,
			Path: inspection.Path,
			Args: inspection.Args,
			Env:  env,
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
			go func(host Host) {
				defer func() {
					wg.Done()
				}()
				c, err := c.GetContainersOnHost(host)
				if err != nil {
					log.Println(err)
					return
				}
				receiver <- c
			}(host)
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

func (c DockerContainerizer) FindAvailableHost(hosts []Host) (*Host, error) {
	for _, host := range hosts {
		containers, err := c.GetContainersOnHost(host)
		if err != nil {
			continue
		}

		for _, _ = range containers {
		}

		return &host, nil
	}

	return nil, errors.New("Cannot find available host")
}

func (c DockerContainerizer) CreateContainer(host Host, config ContainerConfig) (*docker.Container, error) {
	log.Println(config)
	dockerConfig := docker.Config{
		Image: config.Image,
		Cmd: config.Cmd,
	}
	log.Println(dockerConfig)

	options := docker.CreateContainerOptions{
		Config: &dockerConfig,
	}

	client, err := c.getClient(host)
	if err != nil {
		return nil, err
	}

	log.Println(options)

	return client.CreateContainer(options)
}
