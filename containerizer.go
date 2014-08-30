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
	Id             string
	Name           string
	Path           string
	Args           []string
	Env            map[string]string
	Host           Host
	CpuCapacity    int32
	MemoryCapacity int32
}

type ContainerConfig struct {
	Image          string
	Cmd            []string
	Env            []string
	Tags           map[string]string
	CpuCapacity    int32
	MemoryCapacity int32
}

type Containerizer interface {
	GetContainersOnHost(Host) ([]Container, error)
	GetContainersOnHosts([]Host) ([]Container, error)
	FindAvailableHost([]Host, ContainerConfig) (*Host, error)
	RunContainer(Host, ContainerConfig) (*docker.Container, error)
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
		portStr string
	)
	if s, ok := host.ContainerizerInfo["DockerPort"]; ok {
		portStr = s
	} else {
		portStr = strconv.Itoa(c.DefaultPort)
	}
	endpoint := "tcp://" + host.Addr + ":" + portStr
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

		var (
			cpuCapacity    int32
			memoryCapacity int32
		)

		if s, ok := env["DOCKERTIE_CPU_CAPACITY"]; ok {
			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			cpuCapacity = int32(i)
		}

		if s, ok := env["DOCKERTIE_MEMORY_CAPACITY"]; ok {
			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			memoryCapacity = int32(i)
		}

		container := Container{
			Id:             inspection.ID,
			Name:           inspection.Name,
			Path:           inspection.Path,
			Args:           inspection.Args,
			Env:            env,
			Host:           host,
			CpuCapacity:    cpuCapacity,
			MemoryCapacity: memoryCapacity,
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

func (c DockerContainerizer) FindAvailableHost(hosts []Host, config ContainerConfig) (*Host, error) {
	for _, host := range hosts {
		containers, err := c.GetContainersOnHost(host)
		if err != nil {
			continue
		}

		var (
			sumOfCpuCapacity    int32
			sumOfMemoryCapacity int32
		)

		for _, container := range containers {
			sumOfCpuCapacity += container.CpuCapacity
			sumOfMemoryCapacity += container.MemoryCapacity
		}

		sumOfCpuCapacity += config.CpuCapacity
		sumOfMemoryCapacity += config.MemoryCapacity

		if sumOfCpuCapacity <= host.CpuCapacity && sumOfMemoryCapacity <= host.MemoryCapacity {
			return &host, nil
		}
	}

	return nil, errors.New("Cannot find available host")
}

func (c DockerContainerizer) RunContainer(host Host, config ContainerConfig) (*docker.Container, error) {
	tags := config.Tags
	if tags == nil {
		tags = map[string]string{}
	}

	tags["CPU_CAPACITY"] = strconv.Itoa(int(config.CpuCapacity))
	tags["MEMORY_CAPACITY"] = strconv.Itoa(int(config.MemoryCapacity))

	env := config.Env
	for key, value := range tags {
		key = strings.ToUpper(key)
		key = strings.Replace(key, " ", "_", -1)
		env = append(env, "DOCKERTIE_" + key + "=" + value)
	}

	dockerConfig := docker.Config{
		Image: config.Image,
		Cmd:   config.Cmd,
		Env:   env,
	}

	options := docker.CreateContainerOptions{
		Config: &dockerConfig,
	}

	client, err := c.getClient(host)
	if err != nil {
		return nil, err
	}

	container, err := client.CreateContainer(options)
	if err != nil {
		return nil, err
	}

	hostConfig := docker.HostConfig{}
	err = client.StartContainer(container.ID, &hostConfig)
	if err != nil {
		return nil, err
	}

	return container, nil
}
