package main

import (
	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/servers"
)

func main() {
	log.Init(log.InfoLevel)
	cfg, err := config.New()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(cfg.ClientId)

	server, err := servers.New(*cfg, ExampleTasks{})
	if err != nil {
		log.Error(err)
		return
	}

	err = server.Run()
	if err != nil {
		log.Error(err)
		return
	}
}
