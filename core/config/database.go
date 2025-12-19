package config

import (
	"fmt"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Name            string `yaml:"name"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // seconds
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		Name:            "codo",
		User:            "codo",
		Password:        "",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}
}

// Validate validates database configuration
func (c *DatabaseConfig) Validate() error {
	validDrivers := map[string]bool{"postgres": true, "mysql": true, "sqlite": true}
	if !validDrivers[c.Driver] {
		return fmt.Errorf("database.driver must be one of: postgres, mysql, sqlite")
	}
	if c.Driver != "sqlite" {
		if c.Host == "" {
			return fmt.Errorf("database.host is required for %s", c.Driver)
		}
		if c.Port < 1 || c.Port > 65535 {
			return fmt.Errorf("database.port must be between 1 and 65535")
		}
	}
	if c.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if c.MaxOpenConns < 1 {
		return fmt.Errorf("database.max_open_conns must be at least 1")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("database.max_idle_conns must be non-negative")
	}
	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("database.conn_max_lifetime must be non-negative")
	}
	return nil
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.User, c.Password, c.Host, c.Port, c.Name)
	case "sqlite":
		return c.Name // Just the file path for SQLite
	default:
		return ""
	}
}
