package main

import (
	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/servers"
	"ToDoInfo/internal/todoparser"
)

func main() {
	log.Init(log.InfoLevel)
	cfg, err := config.New()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(cfg.ClientId)

	taskProvider := todoparser.New()

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
