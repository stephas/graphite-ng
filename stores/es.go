package stores

import (
	"../config"
	"../metrics"
	"errors"
	"fmt"
	"github.com/graphite-ng/graphite-ng/chains"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	"strconv"
)

type Es struct {
	es_host        string
	es_port        int
	es_max_pending int
	in_port        int
}

func NewEs(config config.Main) *Es {
	api.Domain = config.StoreES.Host
	api.Port = string(config.StoreES.Port)
	es := Es{config.StoreES.Host, config.StoreES.Port, config.StoreES.MaxPending, config.StoreES.CarbonPort}
	return &(es.(Store))
}

func init() {
	InitFn["es"] = NewEs
}

func (e *Es) Add(metric metrics.Metric) (err error) {
	panic("todo")
	return nil
}
func (e *Es) Has(name string) (found bool, err error) {
	out, err := core.SearchUri("carbon-es", "datapoint", fmt.Sprintf("metric:%s&size=1", name), "", 0)
	if err != nil {
		return false, errors.New(fmt.Sprintf("error checking ES for %s: %s", name, err.Error()))
	}
	return (out.Hits.Total > 0), nil
}

func (e *Es) Get(name string) (our_el *chains.ChainEl, err error) {
	our_el = chains.NewChainEl()
	go func(our_el *chains.ChainEl) {
		from := <-our_el.Settings
		until := <-our_el.Settings
		qry := map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]string{"metric": name},
				"range": map[string]interface{}{
					"ts": map[string]string{"from": strconv.Itoa(int(from)), "to": strconv.Itoa(int(until))},
				},
			},
		}
		// { "bool": { "must": [ {"term": ... }, {"range": ...}] }}

		// TODO: sorting?
		out, err := core.SearchRequest(true, "carbon-es", "datapoint", qry, "", 0)
		if err != nil {
			panic(fmt.Sprintf("error reading ES for %s: %s", name, err.Error()))

		}
		// if we don't have enough data to cover the requested timespan, fill with nils
		/* if metric.Data[0].Ts > from {
			for new_ts := from; new_ts < metric.Data[0].Ts; new_ts += 60 {
				our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
			}
		}
		for _, d := range metric.Data {
			if d.Ts >= from && until <= until {
				our_el.Link <- *d
			}
		}
		if metric.Data[len(metric.Data)-1].Ts < until {
			for new_ts := metric.Data[len(metric.Data)-1].Ts + 60; new_ts <= until+60; new_ts += 60 {
				our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
			}
		}
		*/

		fmt.Println(out)
	}(our_el)
	return our_el, nil
}
