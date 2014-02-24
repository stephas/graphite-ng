package main

import (
	"./chains"
	"./stores"
	"fmt"
)

func ReadMetric(name string) (our_el chains.ChainEl) {
	var found bool
	var err error
	for _, s := range stores.List {
		store := s.(stores.Store)
		found, err = store.Has(name)
		if err != nil {
			panic(fmt.Sprintf("Error checking store %s for %s: %s", store, name, err))
		}
		if found {
			our_el, err := store.Get(name)
			if err == nil {
				return *our_el
			} else {
				panic(fmt.Sprintf("Error reading store %s for %s: %s", store, name, err))
			}
		}
	}
	panic("Could not find metric " + name + " in any of the stores")
}
