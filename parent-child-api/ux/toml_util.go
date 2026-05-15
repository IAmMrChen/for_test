package ux

import (
	"os"

	"github.com/BurntSushi/toml"
)

var TomlUtil tomlUtil

type tomlUtil struct{}

// toml.Decode 的 panic 版。
func (x tomlUtil) MustDecode(data string, v any) (toml.MetaData, error) {
	metadata, err := toml.Decode(string(data), v)
	if err != nil {
		panic(err)
	}

	return metadata, nil
}

// 用于从指定路径读取文件，并通过 toml.Decode 反序列化到指定对象里。
func (x tomlUtil) MustDecodeFromFile(filePath string, v any) (toml.MetaData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return x.MustDecode(string(data), v)
}
