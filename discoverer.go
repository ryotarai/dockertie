package main

import (
	"encoding/json"
	"errors"
	"log"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"strings"
)

type Host struct {
	Id string
	Name string
	Tags map[string]string
	Addr string
}

type Discoverer interface {
	GetHosts([]string) ([]Host, error)
}

func NewDiscoverer(name string, c *cli.Context) Discoverer {
	switch name {
	case "ec2":
		return NewEc2Discoverer(c)
	case "json":
		return NewJsonDiscoverer(c)
	}

	return nil
}

type JsonDiscoverer struct {
	Path string
}

func NewJsonDiscoverer(c *cli.Context) JsonDiscoverer {
	path := c.String("json-discoverer-path")
	if (path == "") {
		log.Fatal(errors.New("You must specify --json-discoverer-path option"))
	}

	return JsonDiscoverer{Path: path}
}

func (d JsonDiscoverer) GetHosts(ids []string) ([]Host, error) {
	bytes, err := ioutil.ReadFile(d.Path)
	if (err != nil) {
		log.Fatal(err)
	}

	hosts := []Host{}
	err = json.Unmarshal(bytes, &hosts)
	if (err != nil) {
		log.Fatal(err)
	}

	var filteredHosts []Host
	if (ids == nil) {
		filteredHosts = hosts
	} else {
		filteredHosts = []Host{}
		for _, host := range hosts {
			for _, id := range ids {
				if host.Id == id {
					filteredHosts = append(filteredHosts, host)
				}
				break
			}
		}
	}

	return filteredHosts, nil
}

type Ec2Discoverer struct {
	Client *ec2.EC2
	TagKey string
	TagValue string
}

func NewEc2Discoverer(c *cli.Context) Ec2Discoverer {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}

	tag := c.String("ec2-tag")
	if (tag == "") {
		log.Fatal(errors.New("You must specify --ec2-tag option"))
	}

	t := strings.Split(tag, ":")
	tagKey, tagValue := t[0], t[1]

	var region aws.Region

	switch c.String("ec2-region") {
	case "us-east-1":
		region = aws.USEast
	case "us-west-1":
		region = aws.USWest
	case "us-west-2":
		region = aws.USWest2
	case "eu-west-1":
		region = aws.EUWest
	case "ap-southeast-1":
		region = aws.APSoutheast
	case "ap-southeast-2":
		region = aws.APSoutheast2
	case "ap-northeast-1":
		region = aws.APNortheast
	case "sa-east-1":
		region = aws.SAEast
	}

	client := ec2.New(auth, region)
	return Ec2Discoverer{
		Client: client,
		TagKey: tagKey,
		TagValue: tagValue,
	}
}

func (d Ec2Discoverer) GetHosts(ids []string) ([]Host, error) {
	resp, err := d.Client.Instances(ids, nil)
	if (err != nil) {
		return nil, err
	}

	var hosts []Host
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			tags := map[string]string{}
			for _, tag := range instance.Tags {
				tags[tag.Key] = tag.Value
			}

			if (tags[d.TagKey] != d.TagValue) {
				continue
			}

			host := Host{
				Id: instance.InstanceId,
				Name: tags["Name"],
				Tags: tags,
				Addr: instance.PrivateIpAddress,
			}
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

