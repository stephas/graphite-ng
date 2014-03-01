package functions

import (
	"../chains"
	"fmt"
)

func init() {
	Functions["printStandardJson"] = "OutPrintStandardJson"
}

func OutPrintStandardJson(dep_el chains.ChainEl, until int32) {
	for {
		d := <-dep_el.Link
		fmt.Printf("[%f, %d]", d.Value, d.Ts)
		if d.Ts >= until {
			break
		} else {
			fmt.Printf(", ")
		}
	}

}
