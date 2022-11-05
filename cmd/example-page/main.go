package main

import (
	"github.com/uchr/ToDoInfo/internal/config"
	"github.com/uchr/ToDoInfo/internal/log"
	"github.com/uchr/ToDoInfo/internal/servers"
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
