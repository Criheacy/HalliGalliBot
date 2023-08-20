package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

var lock = &sync.Mutex{}

type Context struct {
	User       User
	Token      Token
	Asset      Asset
	Connection *websocket.Conn
}

var instance *Context

func GetContext() *Context {
	// DCL singleton-instance
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = &Context{}
		} else {
			fmt.Println("Single instance already created.")
		}
	}
	return instance
}
