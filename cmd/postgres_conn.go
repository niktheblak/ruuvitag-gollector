package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

var passwordRegexp = regexp.MustCompile(`password=\S+\s`)

func CreatePsqlInfoString(prefix string) (psqlInfo string, err error) {
	var (
		host     = viper.GetString(fmt.Sprintf("%s.host", prefix))
		port     = viper.GetInt(fmt.Sprintf("%s.port", prefix))
		username = viper.GetString(fmt.Sprintf("%s.username", prefix))
		password = viper.GetString(fmt.Sprintf("%s.password", prefix))
		database = viper.GetString(fmt.Sprintf("%s.database", prefix))
		table    = viper.GetString(fmt.Sprintf("%s.table", prefix))
		sslmode  = viper.GetString(fmt.Sprintf("%s.sslmode", prefix))
		sslcert  = viper.GetString(fmt.Sprintf("%s.sslcert", prefix))
		sslkey   = viper.GetString(fmt.Sprintf("%s.sslkey", prefix))
	)
	if host == "" {
		err = fmt.Errorf("PostgreSQL host must be specified")
		return
	}
	if database == "" {
		err = fmt.Errorf("PostgreSQL database name must be specified")
		return
	}
	if table == "" {
		err = fmt.Errorf("PostgreSQL table name must be specified")
		return
	}
	if sslmode == "" {
		sslmode = "disable"
	}
	builder := new(strings.Builder)
	builder.WriteString("host=")
	builder.WriteString(host)
	builder.WriteString(" ")
	builder.WriteString("port=")
	builder.WriteString(strconv.Itoa(port))
	builder.WriteString(" ")
	if username != "" {
		builder.WriteString("user=")
		builder.WriteString(username)
		builder.WriteString(" ")
	}
	if password != "" {
		builder.WriteString("password=")
		builder.WriteString(password)
		builder.WriteString(" ")
	}
	builder.WriteString("dbname=")
	builder.WriteString(database)
	builder.WriteString(" ")
	builder.WriteString("sslmode=")
	builder.WriteString(sslmode)
	if sslcert != "" && sslkey != "" {
		builder.WriteString(" ")
		builder.WriteString("sslcert=")
		builder.WriteString(sslcert)
		builder.WriteString(" ")
		builder.WriteString("sslkey=")
		builder.WriteString(sslkey)
	}
	psqlInfo = builder.String()
	return
}

func SanitizePassword(psqlInfo string) string {
	return passwordRegexp.ReplaceAllString(psqlInfo, "password=[redacted] ")
}
