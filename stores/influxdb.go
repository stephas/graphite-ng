package stores

import (
	"../chains"
	"../config"
	"../metrics"
	"../util"
	"fmt"
	"github.com/influxdb/influxdb-go"
)

type InfluxdbStore struct {
	client *influxdb.Client
}

func NewInfluxStore(config config.Main) Store {
	c := influxdb.ClientConfig{config.StoreInflux.Host, config.StoreInflux.Username, config.StoreInflux.Password, config.StoreInflux.Database}
	client, err := influxdb.NewClient(&c)
	util.DieIfError(err)
	return InfluxdbStore{client}
}

func init() {
	InitFn["influxdb"] = NewInfluxStore
}

func (i InfluxdbStore) Add(metric metrics.Metric) (err error) {
	panic("todo")
}

func (i InfluxdbStore) Get(name string) (our_el *chains.ChainEl, err error) {

	our_el = chains.NewChainEl()
	go func(our_el *chains.ChainEl) {
		from := <-our_el.Settings
		until := <-our_el.Settings

		query := fmt.Sprintf("select time, value from %s where time > %ds and time < %ds order asc", name, from, until)
		series, err := i.client.Query(query)
		if err != nil {
			panic(err)
		}
		// len(series) can be 0 if there's no datapoints matching the range.
		// so it's up to the caller to make sure the store is supposed to have the data
		// if we don't have enough data to cover the requested timespan, fill with nils
		if len(series) > 0 {
			oldest_dp := int32(series[0].Points[0][0].(float64) / 1000)
			latest_dp := int32(series[len(series)-1].Points[0][0].(float64) / 1000)
			if oldest_dp > from {
				for new_ts := from; new_ts < oldest_dp; new_ts += 60 {
					our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
				}
			}
			for _, values := range series[0].Points {
				ts := int32(values[0].(float64) / 1000)
				val := values[2].(float64)
				dp := metrics.NewDatapoint(ts, val, true)
				our_el.Link <- *dp
			}
			if latest_dp < until {
				for new_ts := latest_dp + 60; new_ts <= until+60; new_ts += 60 {
					our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
				}
			}
		} else {
			for ts := from; ts <= until+60; ts += 60 {
				our_el.Link <- *metrics.NewDatapoint(ts, 0.0, false)
			}
		}
	}(our_el)
	return our_el, nil
}

func (i InfluxdbStore) Has(name string) (found bool, err error) {
	series, err := i.client.Query("select time from " + name + " limit 1;")
	if err != nil {
		panic(err)
	}
	if len(series) > 0 {
		found = true
	}
	return
}
func (i InfluxdbStore) List() (list []string, err error) {
	series, err := i.client.Query("list series")
	if err != nil {
		return
	}
	list = make([]string, len(series))
	for i, s := range series {
		list[i] = s.Name
	}
	return
}
