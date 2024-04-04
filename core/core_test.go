package core

import (
	"testing"

	"github.com/lucasmbaia/power-actions/config"
)

func Test_Run(t *testing.T) {
	config.LoadSingletons()
	Run("../invoke-go-script-on-tag.yml")
}
