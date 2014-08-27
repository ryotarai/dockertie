package main

type Containerizer interface {
	GetContainers()
}

func NewContainerizer(name string) Containerizer {
	switch name {
	case "docker":
		containerizer := DockerContainerizer{}
		return containerizer
	}

	return nil
}

type DockerContainerizer struct {
}

func (c DockerContainerizer) GetContainers() {
}


