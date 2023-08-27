package env

import (
	"fmt"
	"github.com/gorilla/websocket"
	"halligalli/auth"
	"halligalli/game"
	"halligalli/model"
	"sync"
)

var lock = &sync.Mutex{}

type Context struct {
	User           model.User
	Token          auth.Token
	Asset          game.Asset
	ChannelId      string
	ReplyMessageId string
	Connection     *websocket.Conn
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
