package scriptloader

import (
	"embed"
	"fmt"

	"github.com/go-redis/redis/v8"
)

//go:embed *.lua
var luaFS embed.FS

var (
	reserveScript *redis.Script
	cancelScript  *redis.Script
)

func LoadReserveScript() (*redis.Script, error) {
	if reserveScript != nil {
		return reserveScript, nil
	}
	data, err := luaFS.ReadFile("reserve.lua")
	if err != nil {
		return nil, fmt.Errorf("cannot load Lua script: %w", err)
	}
	reserveScript = redis.NewScript(string(data))
	return reserveScript, nil
}

func LoadCancelScript() (*redis.Script, error) {
	if cancelScript != nil {
		return cancelScript, nil
	}
	data, err := luaFS.ReadFile("cancel.lua")
	if err != nil {
		return nil, fmt.Errorf("cannot load Lua script: %w", err)
	}
	cancelScript = redis.NewScript(string(data))
	return cancelScript, nil
}
