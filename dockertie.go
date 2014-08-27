package main

import (
	"os"
	"net/http"
	"log"
	"strconv"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

func main() {
	app := cli.NewApp()
	app.Name = "dockertie"
	app.Usage = "Port for Docker"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "containerizer",
			Value: "docker",
			Usage: "Containerizer (valid options: docker)",
		},
		cli.StringFlag{
			Name: "discoverer",
			Value: "ec2",
			Usage: "Discoverer (valid options: ec2, json)",
		},
		cli.StringFlag{
			Name: "bind",
			Value: "",
			Usage: "IP Address to bind",
		},
		cli.IntFlag{
			Name: "port",
			Value: 8080,
			Usage: "HTTP Port",
		},
		cli.IntFlag{
			Name: "docker-http-port",
			Value: 4243,
			Usage: "HTTP Port for Docker API",
		},
		cli.StringFlag{
			Name: "ec2-region",
			Value: "us-east-1",
			Usage: "EC2 Region (us-east-1, us-west-1, us-west-2, eu-west-1, ap-southeast-1, ap-southeast-2, ap-northeast-1 or sa-east-1)",
		},
		cli.StringFlag{
			Name: "ec2-tag",
			Value: "",
			Usage: "Tag of Docker hosts",
		},
		cli.StringFlag{
			Name: "json-discoverer-path",
			Value: "",
			Usage: "JSON Path for discoverer",
		},
	}
	app.Action = func(c *cli.Context) {
		containerizer := NewContainerizer(c.String("containerizer"), c)
		discoverer := NewDiscoverer(c.String("discoverer"), c)

		handler := HttpHandler{
			Containerizer: containerizer,
			Discoverer: discoverer,
		}

		r := mux.NewRouter()
		r.HandleFunc("/", handler.HandleTop)
		r.HandleFunc("/hosts", handler.HandleHosts)
		r.HandleFunc("/hosts/{id}/containers", handler.HandleHostContainers)
		r.HandleFunc("/containers", handler.HandleContainers)
		http.Handle("/", r)

		addr := c.String("bind") + ":" + strconv.Itoa(c.Int("port"))
		log.Printf("Listening... (%s)", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}

	app.Run(os.Args)
}
