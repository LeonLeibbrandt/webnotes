package main

import (
	"flag"
	"log"

	"web/config"
	"web/global"
	"web/server"
)

func main() {
	configfn := flag.String("config", "", "Config File Name")
	flag.Parse()
	if *configfn == "" {
		flag.PrintDefaults()
		log.Fatal("No config specified")
	}
	c, err := config.NewConfig(*configfn)
	if err != nil {
		log.Fatal(err)
	}

	g, err := global.NewGlobal(c)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(g)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(s.ListenAndServe())
}
