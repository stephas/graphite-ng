package stores

import (
	"../chains"
	"../metrics"
)

var List []interface{}

type Store interface {
	Add(metrics.Metric) (err error)
	//List(<- chan string)
	Get(string) (our_el *chains.ChainEl, err error)
	Has(name string) (found bool, err error)
}
