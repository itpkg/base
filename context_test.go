package base_test

import (
	_ "github.com/lib/pq"
	"testing"

	"github.com/itpkg/base"
)

func TestContextLoad(t *testing.T) {
	ctx := base.Context{}
	err := ctx.Load("test", "../platform/config.yml", true)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = ctx.Load("test", "../platform/config.yml", false)
	if err != nil {
		t.Errorf("%v", err)
	}

}
