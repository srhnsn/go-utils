package database

import (
	"bytes"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/srhnsn/go-utils/log"
)

var DB *sqlx.DB

type DatabaseConfig struct {
	Dsn                string
	MaxIdleConnections uint8
	MaxOpenConnections uint8
}

const defaultMaxIdleConnections = 1
const defaultMaxOpenConnections = 4

func InitDatabase(config DatabaseConfig) {
	var err error

	DB, err = sqlx.Open("mysql", config.Dsn)

	if err != nil {
		log.Error.Fatalf("Could not connect to database: %s", err)
	}

	if config.MaxIdleConnections == 0 {
		DB.SetMaxIdleConns(defaultMaxIdleConnections)
	} else {
		DB.SetMaxIdleConns(int(config.MaxIdleConnections))
	}

	if config.MaxOpenConnections == 0 {
		DB.SetMaxOpenConns(defaultMaxOpenConnections)
	} else {
		DB.SetMaxOpenConns(int(config.MaxOpenConnections))
	}

	DB.MapperFunc(columnMapper)
}

func columnMapper(name string) string {
	runes := []rune(name)

	var buffer bytes.Buffer
	buffer.WriteString(strings.ToLower(string(runes[0])))

	for _, char := range runes[1:] {
		charStr := string(char)

		if charStr == strings.ToUpper(charStr) {
			buffer.WriteString("_")
			buffer.WriteString(strings.ToLower(charStr))
		} else {
			buffer.WriteRune(char)
		}
	}

	return buffer.String()
}
