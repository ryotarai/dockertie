package main

import (
	"os"
	"net/http"
	"log"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
)

func main() {
	app := cli.NewApp()
	app.Name = "dockerport"
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
			Usage: "Discoverer (valid options: ec2)",
		},
		cli.StringFlag{
			Name: "bind",
			Value: "",
			Usage: "IP Address to bind",
		},
		cli.StringFlag{
			Name: "port",
			Value: "8080",
			Usage: "HTTP Port",
		},
	}
	app.Action = func(c *cli.Context) {
		containerizer := NewContainerizer(c.String("containerizer"))
		discoverer := NewDiscoverer(c.String("discoverer"))

		handler := HttpHandler{
			Containerizer: &containerizer,
			Discoverer: &discoverer,
		}

		r := mux.NewRouter()
		r.HandleFunc("/", handler.HandleTop)
		http.Handle("/", r)

		addr := c.String("bind") + ":" + c.String("port")
		log.Printf("Listening... (%s)", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}

	app.Run(os.Args)
}
