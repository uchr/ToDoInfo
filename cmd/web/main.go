package main

import (
	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/servers"
	"ToDoInfo/internal/todoclient"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Error(err)
		return
	}

	log.Init(log.InfoLevel, cfg.LogFolder)

	taskProvider := todoclient.New()

	server, err := servers.New(*cfg, taskProvider)
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
