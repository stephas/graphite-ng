package functions

import (
	"github.com/graphite-ng/graphite-ng/chains"
)

func init() {
	Functions["alias"] = []string{"Alias", "metric", "string"}
}

func Alias(dep_el chains.ChainEl, alias string) (our_el chains.ChainEl) {
	// alias is already set while generating the template. we don't actually need
	// to do anything in this data processing stage.
	return dep_el
}
