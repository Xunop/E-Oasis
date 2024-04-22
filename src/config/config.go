package config

type Config struct {
	// databaseURL is the URL of the database to connect to(sqlite)
	DsnURI string
	// port is the port to listen on
	Port int
	// host is the host to listen on
	Host string
	// data is the directory to store data
	Data string
	// version is the version of the application
	Version string
	// logFile is the file to write logs to
	logFile string
	// logLevel is the level of logging to show
	logLevel string
	// logFilemaxSize is the maximum size of the log file before it is rotated
	logFileMaxSize int
	// logFileMaxBackups is the maximum number of log files to keep
	logFileMaxBackups int
	// logFileMaxAge is the maximum number of days to keep a log file
	logFileMaxAge int
	// logCompress is whether or not to compress the log files
	logCompress bool
}

func init() {
	// init is called before main

}

func NewConfig() *Config {
	// return default configuration
	return &Config{
		logFile:           "e-oasis.log",
		logLevel:          "info",
		logFileMaxSize:    20,
		logFileMaxBackups: 3,
		logFileMaxAge:     30,
		logCompress:       false,
		DsnURI:            "sqlite://db.sqlite",
		Port:              8080,
		Data:              "/var/opt/e-oasis",
		Version:		   "0.0.1",
	}
}
