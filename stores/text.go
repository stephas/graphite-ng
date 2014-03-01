package stores

import (
	"../chains"
	"../config"
	"../metrics"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (t *TextStore) path(name string) string {
	return fmt.Sprintf("%s/%s.txt", t.BasePath, name)
}

type TextStore struct {
	BasePath string
}

func NewTextStore(config config.Main) Store {
	path := config.StoreText.Path
	return TextStore{path}
}

func init() {
	InitFn["text"] = NewTextStore
}

func (t TextStore) Add(metric metrics.Metric) (err error) {
	panic("todo")
}

func (t TextStore) Get(name string) (our_el *chains.ChainEl, err error) {
	var file *os.File
	path := t.path(name)
	if file, err = os.Open(path); err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	datapoints := make([]*metrics.Datapoint, 0)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		ts, _ := strconv.ParseInt(parts[0], 10, 32)
		val, _ := strconv.ParseFloat(parts[1], 64)
		known, _ := strconv.ParseBool(parts[2])
		dp := metrics.NewDatapoint(int32(ts), val, known)
		datapoints = append(datapoints, dp)
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("error reading %s: %s", path, err.Error()))
	}
	metric := metrics.NewMetric(name, datapoints)

	our_el = chains.NewChainEl()
	go func(our_el *chains.ChainEl, metric *metrics.Metric) {
		from := <-our_el.Settings
		until := <-our_el.Settings
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
	}(our_el, metric)
	return our_el, nil
}

func (t TextStore) Has(name string) (found bool, err error) {
	fmt.Println(t.path(name))
	_, err = os.Stat(t.path(name))
	return (err == nil), nil
}
