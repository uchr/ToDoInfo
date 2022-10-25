package main

import (
	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/servers"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Error(err)
		return
	}

	log.Init(log.InfoLevel, cfg.LogFolder)

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
