package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const overridePrefix = "OVERRIDE"

// Configuration is the struct representation of the config yaml. Instantiation should occur through the Configure
// function as it creates internal resources.
type Configuration struct {
	Environment string         `mapstructure:"environment"`
	Port        int            `mapstructure:"port"`
	CORS        CORSConfig     `mapstructure:"cors"`
	Database    DatabaseConfig `mapstructure:"database"`
	Auth        AuthConfig     `mapstructure:"auth"`
	Storage     StorageConfig  `mapstructure:"storage"`
}

// StorageConfig holds file storage configuration.
type StorageConfig struct {
	BasePath string `mapstructure:"base_path"`
}

// AuthConfig holds authentication and JWT configuration.
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns a PostgreSQL connection string built from the config fields.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		d.Host, d.User, d.Password, d.DBName, d.Port, d.SSLMode)
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// Configure is the intended way to instantiate a Configuration. This method should be used over direct instantiation
// because it creates internal resources.
func Configure(filepath string) *Configuration {
	viper.SetConfigFile(filepath)
	viper.SetEnvPrefix(overridePrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("reading config: ", err)
	}

	var conf *Configuration

	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalln("unmarshaling config: ", err)
	}

	return conf
}
