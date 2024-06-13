package psql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	passwordRegexp   = regexp.MustCompile(`password=\S+\s`)
	whitespaceRegexp = regexp.MustCompile(`\s+`)
)

// RemovePassword removes password from a psqlInfo style string
func RemovePassword(psqlInfo string) string {
	return passwordRegexp.ReplaceAllString(psqlInfo, "password=[redacted] ")
}

// TrimQuery replaces all whitespace (newlines and repeated spaces) from a string with one space
func TrimQuery(q string) string {
	return strings.TrimSpace(whitespaceRegexp.ReplaceAllString(q, " "))
}

func CreatePsqlInfoString(vpr *viper.Viper, prefix string) (psqlInfo string, err error) {
	var (
		host     = vpr.GetString(fmt.Sprintf("%s.host", prefix))
		port     = vpr.GetInt(fmt.Sprintf("%s.port", prefix))
		username = vpr.GetString(fmt.Sprintf("%s.username", prefix))
		password = vpr.GetString(fmt.Sprintf("%s.password", prefix))
		database = vpr.GetString(fmt.Sprintf("%s.database", prefix))
		table    = vpr.GetString(fmt.Sprintf("%s.table", prefix))
		sslmode  = vpr.GetString(fmt.Sprintf("%s.sslmode", prefix))
		sslcert  = vpr.GetString(fmt.Sprintf("%s.sslcert", prefix))
		sslkey   = vpr.GetString(fmt.Sprintf("%s.sslkey", prefix))
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

func AddPsqlFlags(fs *pflag.FlagSet, vpr *viper.Viper, prefix string) {
	fs.String(fmt.Sprintf("%s.host", prefix), "", "database host or IP")
	fs.Int(fmt.Sprintf("%s.port", prefix), 0, "database port")
	fs.String(fmt.Sprintf("%s.username", prefix), "", "database username")
	fs.String(fmt.Sprintf("%s.password", prefix), "", "database password")
	fs.String(fmt.Sprintf("%s.database", prefix), "", "database name")
	fs.String(fmt.Sprintf("%s.table", prefix), "", "table name")
	fs.String(fmt.Sprintf("%s.sslmode", prefix), "", "SSL mode")
	fs.String(fmt.Sprintf("%s.sslcert", prefix), "", "path to SSL certificate file")
	fs.String(fmt.Sprintf("%s.sslkey", prefix), "", "path to SSL key file")
	fs.String(fmt.Sprintf("%s.column.time", prefix), "", "time column name")
	fs.String(fmt.Sprintf("%s.type", prefix), "", "database type, postgres or timescaledb")

	vpr.SetDefault(fmt.Sprintf("%s.port", prefix), "5432")
	vpr.SetDefault(fmt.Sprintf("%s.sslmode", prefix), "disable")
	vpr.SetDefault(fmt.Sprintf("%s.column.time", prefix), "time")
	vpr.SetDefault(fmt.Sprintf("%s.type", prefix), "postgres")
}
