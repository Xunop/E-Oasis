package version

var version = "0.0.1"

func GetCurrentVersion() string {
    return version
}

func SetVersion(v string) {
    version = v
}
