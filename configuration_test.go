package base_test

import (
	"testing"

	"github.com/itpkg/base"
)

func TestLoadConfig(t *testing.T) {
	app := base.Application{}
	err := app.Load("../platform/config.yml", false)
	if err != nil {
		t.Errorf("%v", err)
	}
}
