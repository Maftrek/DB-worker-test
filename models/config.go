package models

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	configPath = "config/config.toml"
)

type duration time.Duration

// Config struct
type Config struct {
	ServerOpt     ServerOpt `toml:"ServerOpt"`
	HashSum       []byte
	NatsServer    NatsServer    `toml:"NatsServer"`
	SQLDataBase   SQLDataBase   `toml:"SQLDataBase"`
	NoSQLDataBase NoSQLDataBase `toml:"NoSQLDataBase"`
}

type SQLDataBase struct {
	Server          string   `toml:"Server"`
	Database        string   `toml:"Database"`
	ApplicationName string   `toml:"ApplicationName"`
	MaxIdleConns    int      `toml:"MaxIdleConns"`
	MaxOpenConns    int      `toml:"MaxOpenConns"`
	ConnMaxLifetime duration `toml:"ConnMaxLifetime"`
	Port            int      `toml:"Port"`
	UserID          string
	Password        string
}

type NoSQLDataBase struct {
	Database string `toml:"Database"`
	Server   string `toml:"Server"`
	Port     int    `toml:"Port"`
	UserID   string
	Password string
}

func (d *duration) UnmarshalText(text []byte) error {
	temp, err := time.ParseDuration(string(text))
	*d = duration(temp)
	return err
}

// ServerOpt struct
type ServerOpt struct {
	ReadTimeout  duration
	WriteTimeout duration
	IdleTimeout  duration
}

// LoadConfig from path
func LoadConfig(c *Config) {
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		return
	}

	c.SQLDataBase.UserID = "role_1"
	c.SQLDataBase.Password = "1"

	c.NoSQLDataBase.UserID = "admin"
	c.NoSQLDataBase.Password = "123"

	c.HashSum = GetHashSum()
}

func getCredential(path string) string {
	c, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf(`{"err": %s}`, err.Error())
	}
	return strings.TrimSpace(string(c))
}

// GetHashSum of config file
func GetHashSum() []byte {
	paths := []string{
		configPath,
	}
	h := sha256.New()

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		if _, err = io.Copy(h, f); err != nil {
			return nil
		}
	}

	return h.Sum(nil)
}
