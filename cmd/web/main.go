package main

import (
	"fmt"

	"github.com/uchr/ToDoInfo/internal/config"
	"github.com/uchr/ToDoInfo/internal/log"
	"github.com/uchr/ToDoInfo/internal/servers"
	"github.com/uchr/ToDoInfo/internal/todoclient"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Panic("unhandled panic in main", fmt.Errorf("%v", r))
		}
	}()

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
