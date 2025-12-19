package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDatabaseConfig(t *testing.T) {
	cfg := DefaultDatabaseConfig()

	assert.Equal(t, "postgres", cfg.Driver)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "codo", cfg.Name)
	assert.Equal(t, "codo", cfg.User)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, "disable", cfg.SSLMode)
	assert.Equal(t, 25, cfg.MaxOpenConns)
	assert.Equal(t, 5, cfg.MaxIdleConns)
	assert.Equal(t, 300, cfg.ConnMaxLifetime)
}

func TestDatabaseConfig_Validate_Valid(t *testing.T) {
	cfg := DefaultDatabaseConfig()

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestDatabaseConfig_Validate_ValidMySQL(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:       "mysql",
		Host:         "localhost",
		Port:         3306,
		Name:         "mydb",
		User:         "root",
		MaxOpenConns: 10,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestDatabaseConfig_Validate_ValidSQLite(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:       "sqlite",
		Name:         "test.db",
		MaxOpenConns: 1,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestDatabaseConfig_Validate_InvalidDriver(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Driver = "oracle"

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.driver must be one of: postgres, mysql, sqlite")
}

func TestDatabaseConfig_Validate_EmptyDriver(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Driver = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.driver must be one of")
}

func TestDatabaseConfig_Validate_MissingHost_Postgres(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Host = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.host is required")
}

func TestDatabaseConfig_Validate_MissingHost_MySQL(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:       "mysql",
		Host:         "",
		Port:         3306,
		Name:         "mydb",
		MaxOpenConns: 1,
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.host is required for mysql")
}

func TestDatabaseConfig_Validate_SQLiteNoHost(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:       "sqlite",
		Host:         "", // SQLite doesn't need host
		Port:         0,  // SQLite doesn't need port
		Name:         "data.db",
		MaxOpenConns: 1,
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestDatabaseConfig_Validate_InvalidPort_Zero(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Port = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.port must be between 1 and 65535")
}

func TestDatabaseConfig_Validate_InvalidPort_Negative(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Port = -1

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.port must be between 1 and 65535")
}

func TestDatabaseConfig_Validate_InvalidPort_TooHigh(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Port = 70000

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.port must be between 1 and 65535")
}

func TestDatabaseConfig_Validate_MissingName(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.Name = ""

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.name is required")
}

func TestDatabaseConfig_Validate_InvalidMaxOpenConns_Zero(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.MaxOpenConns = 0

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.max_open_conns must be at least 1")
}

func TestDatabaseConfig_Validate_InvalidMaxOpenConns_Negative(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.MaxOpenConns = -5

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.max_open_conns must be at least 1")
}

func TestDatabaseConfig_Validate_InvalidMaxIdleConns_Negative(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.MaxIdleConns = -1

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.max_idle_conns must be non-negative")
}

func TestDatabaseConfig_Validate_ZeroMaxIdleConns(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.MaxIdleConns = 0

	err := cfg.Validate()

	assert.NoError(t, err) // Zero is valid (disables idle connections)
}

func TestDatabaseConfig_Validate_InvalidConnMaxLifetime_Negative(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.ConnMaxLifetime = -1

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.conn_max_lifetime must be non-negative")
}

func TestDatabaseConfig_Validate_ZeroConnMaxLifetime(t *testing.T) {
	cfg := DefaultDatabaseConfig()
	cfg.ConnMaxLifetime = 0

	err := cfg.Validate()

	assert.NoError(t, err) // Zero is valid (no limit)
}

func TestDatabaseConfig_DSN_Postgres(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:   "postgres",
		Host:     "db.example.com",
		Port:     5432,
		Name:     "myapp",
		User:     "appuser",
		Password: "secret",
		SSLMode:  "require",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "host=db.example.com port=5432 user=appuser password=secret dbname=myapp sslmode=require", dsn)
}

func TestDatabaseConfig_DSN_Postgres_NoPassword(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:   "postgres",
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "test",
		Password: "",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "host=localhost port=5432 user=test password= dbname=testdb sslmode=disable", dsn)
}

func TestDatabaseConfig_DSN_MySQL(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:   "mysql",
		Host:     "mysql.example.com",
		Port:     3306,
		Name:     "myapp",
		User:     "appuser",
		Password: "secret",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "appuser:secret@tcp(mysql.example.com:3306)/myapp?parseTime=true", dsn)
}

func TestDatabaseConfig_DSN_MySQL_NoPassword(t *testing.T) {
	cfg := DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Name:     "testdb",
		User:     "root",
		Password: "",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "root:@tcp(localhost:3306)/testdb?parseTime=true", dsn)
}

func TestDatabaseConfig_DSN_SQLite(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "sqlite",
		Name:   "data.db",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "data.db", dsn)
}

func TestDatabaseConfig_DSN_SQLite_WithPath(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "sqlite",
		Name:   "/var/lib/app/data.db",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "/var/lib/app/data.db", dsn)
}

func TestDatabaseConfig_DSN_SQLite_Memory(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "sqlite",
		Name:   ":memory:",
	}

	dsn := cfg.DSN()

	assert.Equal(t, ":memory:", dsn)
}

func TestDatabaseConfig_DSN_UnknownDriver(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "oracle",
		Host:   "localhost",
		Name:   "testdb",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "", dsn)
}

func TestDatabaseConfig_DSN_EmptyDriver(t *testing.T) {
	cfg := DatabaseConfig{
		Driver: "",
		Host:   "localhost",
		Name:   "testdb",
	}

	dsn := cfg.DSN()

	assert.Equal(t, "", dsn)
}

func TestDatabaseConfig_Validate_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		cfg     DatabaseConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid postgres defaults",
			cfg:     DefaultDatabaseConfig(),
			wantErr: false,
		},
		{
			name: "valid mysql",
			cfg: DatabaseConfig{
				Driver:       "mysql",
				Host:         "localhost",
				Port:         3306,
				Name:         "test",
				MaxOpenConns: 1,
			},
			wantErr: false,
		},
		{
			name: "valid sqlite minimal",
			cfg: DatabaseConfig{
				Driver:       "sqlite",
				Name:         ":memory:",
				MaxOpenConns: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid driver",
			cfg: DatabaseConfig{
				Driver:       "mongodb",
				Host:         "localhost",
				Port:         27017,
				Name:         "test",
				MaxOpenConns: 1,
			},
			wantErr: true,
			errMsg:  "database.driver must be one of",
		},
		{
			name: "postgres without host",
			cfg: DatabaseConfig{
				Driver:       "postgres",
				Host:         "",
				Port:         5432,
				Name:         "test",
				MaxOpenConns: 1,
			},
			wantErr: true,
			errMsg:  "database.host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
