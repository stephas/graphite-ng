package stores

import (
	"../chains"
	"../config"
	"../metrics"
	"errors"
	"fmt"
)

var InitFn = make(map[string]func(config config.Main) Store)

type Store interface {
	Add(metric metrics.Metric) (err error)
	Get(name string) (our_el *chains.ChainEl, err error)
	Has(name string) (found bool, err error)
	//List(<- chan string)
}

var List = make(map[string]*Store)

func Init(config config.Main) (err error) {
	fmt.Println("stores.Init called")
	for _, key := range config.Stores {
		fmt.Println("adding store", key)
		if constructor, ok := InitFn[key]; ok {
			fmt.Println("ok")
			store := constructor(config)
			List[key] = &store
		} else {
			fmt.Println("not ok")
			return errors.New("no such store: " + key)
		}
	}
	return
}
