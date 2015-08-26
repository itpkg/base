package base_test

import (
	"github.com/itpkg/base"
	"log"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := base.Load("../platform/config.yml")
	if err == nil {
		log.Printf("Configuration: %v", cfg)
		log.Printf(cfg.DbUrl())
	} else {
		t.Errorf("%v", err)
	}
}
