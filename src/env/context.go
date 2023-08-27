package env

import (
	"fmt"
	"github.com/gorilla/websocket"
	"halligalli/common"
	"halligalli/model"
	"sync"
	"time"
)

var lock = &sync.Mutex{}

type Context struct {
	User           model.User
	Token          common.Token
	Asset          common.Asset
	GameRule       common.Rule
	ReplyMessageId string
	Connection     *websocket.Conn
}

var instance *Context

func OnInit(context *Context) {
	context.GameRule = common.Rule{
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
