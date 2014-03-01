package stores

import (
	"../chains"
	"../config"
	"../metrics"
	"../util"
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
	InitFn["influx"] = NewInfluxStore
}

func (i InfluxdbStore) Add(metric metrics.Metric) (err error) {
	panic("todo")
}

func (t InfluxdbStore) Get(name string) (our_el *chains.ChainEl, err error) {

	our_el = chains.NewChainEl()
	go func(our_el *chains.ChainEl) {
		from := <-our_el.Settings
		until := <-our_el.Settings

		series, err := t.client.Query("select timestamp, value from " + name)
		if err != nil {
			panic(err)
		}
		//if len(series) != 1 {
		//    return nil, errors.New("expected 1 result from influxdb, not " + string(len(series)))
		//}
		datapoints := make([]*metrics.Datapoint, 0)
		for _, values := range series[0].Points {
			ts := values[0]
			val := values[1]
			dp := metrics.NewDatapoint(int32(ts.(int32)), val.(float64), true)
			datapoints = append(datapoints, dp)
		}
		metric := metrics.NewMetric(name, datapoints)

		// if we don't have enough data to cover the requested timespan, fill with nils
		if metric.Data[0].Ts > from {
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
	}(our_el)
	return our_el, nil
}

func (i InfluxdbStore) Has(name string) (found bool, err error) {
	panic("todo")
}
