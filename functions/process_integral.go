package functions

import (
	"github.com/graphite-ng/graphite-ng/chains"
	"github.com/graphite-ng/graphite-ng/metrics"
)

func init() {
	Functions["integral"] = []string{"ProcessIntegral", "metric"}
}
func ProcessIntegral(dep_el chains.ChainEl) (our_el chains.ChainEl) {
	our_el = *chains.NewChainEl()
	go func(our_el chains.ChainEl, dep_el chains.ChainEl) {
		from := <-our_el.Settings
		until := <-our_el.Settings
		dep_el.Settings <- from - 60
		dep_el.Settings <- until
		sum := float64(0)
		d := <-dep_el.Link
		last_ts := d.Ts

		for {
			d = <-dep_el.Link
			if d.Known {
				sum += d.Value * float64(d.Ts-last_ts)
			}
			our_el.Link <- *metrics.NewDatapoint(d.Ts, sum, true)
			last_ts = d.Ts
			if d.Ts >= until {
				return
			}
		}
	}(our_el, dep_el)
	return
}
