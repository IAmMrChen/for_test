package resx

import (
	"fmt"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/mysql"
)

type db struct {
	Main *sqlmer.DbClientEx
}

var Db *db

func InitDb(connectionStrings map[string]string) {
	if Db != nil {
		return
	}

	Db = &db{
		Main: Db.newDbClient(connectionStrings, "Main"),
	}
}

func (*db) newDbClient(connectionStrings map[string]string, name string) *sqlmer.DbClientEx {
	conn, ok := connectionStrings[name]
	if !ok {
		panic(fmt.Errorf("database connection string not found: %s", name))
	}

	client, err := mysql.NewMySqlDbClient(conn)
	if err != nil {
		panic(fmt.Errorf("initializing database %s: %s", name, err))
	}

	return sqlmer.Extend(client)
}
