package config

import (
	"log"
	"os"
	"strconv"
)

type MainConfig struct {
	BindAddress            string            `yaml:"bindAddress"`
	BindPort               int               `yaml:"bindPort"`
	ApiIps                 []string          `yaml:"apiIpRanges"`
	MonitoringIpRanges     []string          `yaml:"monitoringIpRanges"`
	PrivateIps             []string          `yaml:"privateIpRanges"`
	QueryTokenKey          string            `yaml:"queryTokenKey"`
	MediaMtxUrlBase        string            `yaml:"mediamtxApiBase"`
	MediaMtxUrlBasePublish string            `yaml:"mediamtxApiBasePublish"`
	ForwardAuth            ForwardAuthConfig `yaml:"forwardAuth"`
	Database               DatabaseConfig    `yaml:"database"`
}

type ForwardAuthConfig struct {
	UriHeader string `yaml:"uriHeader"`
	IpHeader  string `yaml:"ipHeader"`
	BasePath  string `yaml:"basePath"`
}

type DatabaseConfig struct {
	Hostname                string `yaml:"hostname"`
	Port                    int    `yaml:"port"`
	Database                string `yaml:"database"`
	Username                string `yaml:"username"`
	Password                string `yaml:"password"`
	PollInterval            int    `yaml:"pollInterval"`
	CacheDuration           int    `yaml:"cacheDuration"`
	ConnectionTrackDuration int    `yaml:"connectionTrackDuration"`
}

func NewMainConfig() MainConfig {
	return MainConfig{
		BindAddress:        "",
		BindPort:           8080,
		ApiIps:             []string{"127.0.0.0/8", "::1/128"},
		MonitoringIpRanges: []string{"127.0.0.0/8", "::1/128"},
		PrivateIps: []string{"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12", "192.168.0.0/16", "198.18.0.0/15",
			"::1/128", "fc00::/7", "fe80::/64"},
		QueryTokenKey:          "token",
		MediaMtxUrlBase:        "http://localhost:9997",
		MediaMtxUrlBasePublish: "http://localhost:9997",
		ForwardAuth: ForwardAuthConfig{
			UriHeader: "X-Forwarded-Uri",
			IpHeader:  "X-Forwarded-For",
			BasePath:  "/thumbnails",
		},
		Database: DatabaseConfig{
			Hostname:                "localhost",
			Port:                    5432,
			Database:                "mediamtxauth",
			Username:                "mediamtxauth",
			Password:                "",
			PollInterval:            15,
			CacheDuration:           300,
			ConnectionTrackDuration: 60,
		},
	}
}

func (m *MainConfig) envInit() {
	readEnvString("BIND_ADDRESS", &m.BindAddress)
	readEnvInt("BIND_PORT", &m.BindPort)
	// Database
	readEnvString("DB_HOSTNAME", &m.Database.Hostname)
	readEnvInt("DB_PORT", &m.Database.Port)
	readEnvString("DB_DATABASE", &m.Database.Database)
	readEnvString("DB_USERNAME", &m.Database.Username)
	readEnvString("DB_PASSWORD", &m.Database.Password)
}

func readEnvString(env string, res *string) {
	val, exist := os.LookupEnv(env)
	if exist && len(val) > 0 {
		*res = val
	}
}

func readEnvInt(env string, res *int) {
	val, exist := os.LookupEnv(env)
	if exist && len(val) > 0 {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("Invalid int %v\n", val)
			return
		}
		*res = parsed
	}
}
