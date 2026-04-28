package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	t.Run("Missing JWT Secret", func(t *testing.T) {
		os.Clearenv()
		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET must be set")
	})

	t.Run("Valid Config", func(t *testing.T) {
		os.Clearenv()
		require.NoError(t, os.Setenv("JWT_SECRET", "very-long-secret-key-at-least-32-chars"))
		require.NoError(t, os.Setenv("APP_PORT", "9090"))
		require.NoError(t, os.Setenv("DB_HOST", "db-host"))

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "9090", cfg.App.Port)
		assert.Equal(t, "db-host", cfg.DB.Host)
		assert.Equal(t, "very-long-secret-key-at-least-32-chars", cfg.JWT.Secret)
	})
}

func TestDBConfig_DSN(t *testing.T) {
	c := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "pass",
		Name:     "db",
		SSLMode:  "disable",
	}
	assert.Equal(t, "host=localhost port=5432 user=user password=pass dbname=db sslmode=disable", c.DSN())
	assert.Equal(t, "host=localhost port=5432 user=user password=**** dbname=db sslmode=disable", c.MaskedDSN())
}

func TestGetEnvHelpers(t *testing.T) {
	require.NoError(t, os.Setenv("TEST_INT", "100"))
	assert.Equal(t, 100, getEnvInt("TEST_INT", 10))
	assert.Equal(t, 10, getEnvInt("NON_EXIST", 10))

	require.NoError(t, os.Setenv("TEST_BOOL", "true"))
	assert.Equal(t, true, getEnvBool("TEST_BOOL", false))
	assert.Equal(t, false, getEnvBool("NON_EXIST", false))
}
