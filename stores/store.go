package stores

import (
	"../chains"
	"../config"
	"../metrics"
)

var InitFn map[string]func(config config.Main) //*Store

type Store interface {
	Add(metric metrics.Metric) (err error)
	//List(<- chan string)
	Get(name string) (our_el *chains.ChainEl, err error)
	Has(name string) (found bool, err error)
}
