package main

type Discoverer interface {
}

func NewDiscoverer(name string) Discoverer {
	switch name {
	case "ec2":
		discoverer := Ec2Discoverer{}
		return discoverer
	}

	return nil
}

type Ec2Discoverer struct {
}


