package influxdb

type Config struct {
	Addr        string
	Token       string
	Database    string
	Measurement string
	Username    string
	Password    string
}
