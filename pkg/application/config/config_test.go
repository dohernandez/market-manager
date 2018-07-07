package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertConfig(t *testing.T, globalconfig *Specification) {
	assert.Equal(t, "debug", globalconfig.LogLevel)
	assert.Equal(t, true, globalconfig.Debug)
}

func TestLoadEnv(t *testing.T) {
	// Global Config
	setGlobalConfigEnv()

	globalConfig, err := LoadEnv()
	assert.Nil(t, err)

	assertConfig(t, globalConfig)
}

func setGlobalConfigEnv() {
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DEBUG", "true")

}
