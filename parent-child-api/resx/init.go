package resx

import "os"

func init() {
	confPath := locateConf("conf.toml")
	InitConf(confPath)
	InitLog()
}

func locateConf(name string) string {
	const maxDepth = 4 // 最多查找 4 层。

	exist := func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}

	// 查找当前目录。
	path := "./" + name
	if exist(path) {
		return path
	}

	// 递归查找。
	path = name
	for i := 1; i < maxDepth; i++ {
		path = "../" + name
		if exist(path) {
			return path
		}
	}

	return path
}
