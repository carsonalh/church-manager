package util_test

import (
	"fmt"
	"testing"

	"github.com/carsonalh/churchmanagerbackend/server/util"
)

func TestNewInit(t *testing.T) {
	values := []any{
		1,
		true,
		"In the beginning, God created the heavens and the earth.",
	}

	for _, v := range values {
		t.Run(fmt.Sprintf("NewInit(%v)", v), func(t *testing.T) {
			ni := util.NewPtr(v)
			if ni == nil {
				t.Error("value is nil")
				return
			}
			if *ni != v {
				t.Errorf("pointer value did not match, expected %v and got %v", v, *ni)
			}
		})
	}
}
