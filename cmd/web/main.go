package main

import (
	"ToDoInfo/internal/config"
	"ToDoInfo/internal/log"
	"ToDoInfo/internal/login"
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

	token, err := login.Login(cfg.ClientId)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug(token)

	server, err := servers.New(token)
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
