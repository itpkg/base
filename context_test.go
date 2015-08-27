package base_test

import (
	_ "github.com/lib/pq"
	"testing"

	"github.com/itpkg/base"
)

func TestContextLoad(t *testing.T) {
	ctx := base.Context{}
	err := ctx.Load("../platform/config.yml")
	if err != nil {
		t.Errorf("%v", err)
	}
}