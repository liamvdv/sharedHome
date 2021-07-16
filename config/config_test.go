package config_test

import (
	"testing"

	"github.com/liamvdv/sharedHome/config"
)

func TestLoadAndStoreConfigFile(t *testing.T) {
	defer func() {
		err := config.Delete(config.D_ConfigFolder)
		if err != nil {
			t.Error(err)
		}
	}()

	config.InitVars()
	c, err := config.LoadConfigFile()
	if err != nil {
		t.Error(err)
	}
	if config.StoreConfigFile(c) != nil {
		t.Error(err)
	}

	c, err = config.LoadConfigFile()
	if err != nil {
		t.Error(err)
	}
	if config.StoreConfigFile(c) != nil {
		t.Error(err)
	}
}
