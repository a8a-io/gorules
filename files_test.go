package gorules_test

import (
	"testing"

	"github.com/a8a-io/gorules"
)

func TestRules(t *testing.T) {
	gorules.LoadRulesFromFolder("./rules/")
}
