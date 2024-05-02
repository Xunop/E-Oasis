package config

const (
	defalutLogFile                = "logs.log"
	defaultLogLevel               = "debug"
	defaultLogFileMaxSize         = 20
	defaultLogFileMaxBackups      = 3
	defaultLogFileMaxAge          = 28
	defaultLogCompress            = false
	defaultPort                   = 8080
	defaultHost                   = "0.0.0.0"
	defaultData                   = "/var/opt/e-oasis"
	defaultDSN                    = defaultData + "/e-oasis.db"
	defaultMetaDSN                = defaultData + "/metadata.db"
	defaultMetricsCollector       = false
	defaultMetricsRefreshInterval = 15
	defaultMetricsAllowedNetworks = "127.0.0.1/8"
	defaultMetricsUsername        = ""
	defaultMetricsPassword        = ""
	defaultWorkerPoolSize         = 10
	defaultMaxUploadSize          = 100
	defaultSupportedTypes         = "application/zip"
)

type Option struct {
	Key   string
	Value interface{}
}

// Why use mapstructure instead of json, if use json as field tags, it can't recgnize the field, since the viper use mapstructure.
// see: https://pkg.go.dev/github.com/mitchellh/mapstructure#hdr-Field_Tags
type Options struct {
	// LogFile is the file to write logs to
	LogFile string `mapstructure:"log_file"`
	// LogLevel is the level of logging to show
	LogLevel string `mapstructure:"log_level"`
	// LogFilemaxSize is the maximum size of the log file before it is rotated
	LogFileMaxSize int `mapstructure:"log_file_max_size"`
	// LogFileMaxBackups is the maximum number of log files to keep
	LogFileMaxBackups int `mapstructure:"log_file_max_backups"`
	// LogFileMaxAge is the maximum number of days to keep a log file
	LogFileMaxAge int `mapstructure:"log_file_max_age"`
	// LogCompress is whether or not to compress the log files
	LogCompress bool `mapstructure:"log_compress"`
	// databaseURL is the URL of the database to connect to (sqlite)
	DSN string `mapstructure:"dsn_uri"`
	// metaDSN is the URL of the calibre database to connect to (sqlite)
	MetaDSN string `mapstructure:"meta_dsn_uri"`
	// port is the port to listen on
	Port int `mapstructure:"port"`
	// host is the host to listen on
	Host string `mapstructure:"host"`
	// data is the directory to store data
	Data           string `mapstructure:"data"`
	WorkerPoolSize int    `mapstructure:"worker_pool_size"`
	// MaxUploadSize is the maximum size of the upload, in MiB
	MaxUploadSize int64 `mapstructure:"max_upload_size"`
	// SupportedTypes is the supported types of books
	SupportedTypes []string `mapstructure:"supported_types"`
	// For metrics
	MetricsCollector       bool     `mapstructure:"metrics_collector"`
	MetricsRefreshInterval int      `mapstructure:"metrics_refresh_interval"`
	MetricsAllowedNetworks []string `mapstructure:"metrics_allowed_networks"`
	MetricsUsername        string   `mapstructure:"metrics_username"`
	MetricsPassword        string   `mapstructure:"metrics_password"`
}

func GetDefaultOptions() *Options {
	Opts = &Options{
		LogFile:                defalutLogFile,
		LogLevel:               defaultLogLevel,
		LogFileMaxSize:         defaultLogFileMaxSize,
		LogFileMaxBackups:      defaultLogFileMaxBackups,
		LogFileMaxAge:          defaultLogFileMaxAge,
		LogCompress:            defaultLogCompress,
		DSN:                    defaultDSN,
		MetaDSN:                defaultMetaDSN,
		Port:                   defaultPort,
		Host:                   defaultHost,
		Data:                   defaultData,
		WorkerPoolSize:         defaultWorkerPoolSize,
		SupportedTypes:         []string{defaultSupportedTypes},
		MetricsCollector:       defaultMetricsCollector,
		MetricsRefreshInterval: defaultMetricsRefreshInterval,
		MetricsAllowedNetworks: []string{defaultMetricsAllowedNetworks},
		MetricsUsername:        defaultMetricsUsername,
		MetricsPassword:        defaultMetricsPassword,
	}
	return Opts
}
