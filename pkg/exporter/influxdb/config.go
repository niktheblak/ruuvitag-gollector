package influxdb

type Config struct {
	Addr        string
	Org         string
	Bucket      string
	Database    string
	Measurement string
	Token       string
	Username    string
	Password    string
}
