package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"runtime"
)

type Token struct {
	AppID       uint64
	AccessToken string
}

func (token Token) GetString() string {
	return fmt.Sprintf("%s %v.%s", "Bot", token.AppID, token.AccessToken)
}

func GetConfigPath(name string) string {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		return fmt.Sprintf("%s/%s", path.Dir(filename), name)
	}
	return ""
}

func LoadTokenFromConfig(file string) error {
	var conf struct {
		AppID uint64 `yaml:"appid"`
		Token string `yaml:"token"`
	}
	content, err := os.ReadFile(GetConfigPath(file))
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return err
	}

	GetContext().Token = Token{
		AppID:       conf.AppID,
		AccessToken: conf.Token,
	}
	return nil
}
