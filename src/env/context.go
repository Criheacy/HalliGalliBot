package env

import (
	"fmt"
	"github.com/gorilla/websocket"
	"halligalli/auth"
	"halligalli/game"
	"halligalli/model"
	"sync"
	"time"
)

var lock = &sync.Mutex{}

type Context struct {
	User           model.User
	Token          auth.Token
	Asset          game.Asset
	GameRule       game.Rule
	ReplyMessageId string
	Connection     *websocket.Conn
}

var instance *Context

func OnInit(context *Context) {
	context.GameRule = game.Rule{
		ValidCardNumber:  5,
		FruitNumberToWin: 5,
		DealInterval:     7 * time.Second,
	}
}

func GetContext() *Context {
	// DCL singleton-instance
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = &Context{}
			OnInit(instance)
		} else {
			fmt.Println("Single instance already created.")
		}
	}
	return instance
}
