package auth

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"halligalli/common"
	"halligalli/env"
	"os"
	"path"
	"runtime"
)

const ConfigFilePath = "./config.yaml"

func GetTokenString(token common.Token) string {
	return fmt.Sprintf("%s %v.%s", "Bot", token.AppID, token.AccessToken)
}

func GetPath(name string) string {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		return fmt.Sprintf("%s/%s", path.Dir(filename), name)
	}
	return ""
}

func LoadTokenFromConfig() error {
	var conf struct {
		AppID uint64 `yaml:"appid"`
		Token string `yaml:"token"`
	}
	content, err := os.ReadFile(GetPath(ConfigFilePath))
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return err
	}

	env.GetContext().Token = common.Token{
		AppID:       conf.AppID,
		AccessToken: conf.Token,
	}
	return nil
}
