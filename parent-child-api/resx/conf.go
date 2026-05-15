package resx

import "parent-child-api/ux"

var Conf *config

type config struct {
	// 环境。
	Env string

	// 监听的端口。
	ApiPort int

	// JWT 签名密钥。MVP demo 先使用 HS256，后续可替换为更完整的登录体系。
	JwtSecret string

	// 数据库连接： DB_name -> connection_string
	DB map[string]string
}

func InitConf(path string) {
	Conf = &config{}
	ux.TomlUtil.MustDecodeFromFile(path, Conf)
}
