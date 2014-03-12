package stores

import (
	"../chains"
	"../config"
	"../metrics"
	"bufio"
	"fmt"
	"io/ioutil"
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
	our_el = chains.NewChainEl()

	go func(our_el *chains.ChainEl) {
		var file *os.File
		path := t.path(name)
		if file, err = os.Open(path); err != nil {
			panic(err)
		}
		defer file.Close()
		from := <-our_el.Settings
		until := <-our_el.Settings

		scanner := bufio.NewScanner(file)
		first := true
		// this will be used to fill the potential gap between last datapoint and until,
		// but also if there were no (matching) datapoints in the file at all.
		last_ts := from - 60
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, " ")
			ts, _ := strconv.ParseInt(parts[0], 10, 32)
			val, _ := strconv.ParseFloat(parts[1], 64)
			known, _ := strconv.ParseBool(parts[2])
			dp := metrics.NewDatapoint(int32(ts), val, known)
			if first {
				if from < dp.Ts {
					for new_ts := from; new_ts < dp.Ts; new_ts += 60 {
						our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
					}
				}
			}
			if dp.Ts >= from && dp.Ts <= until {
				our_el.Link <- *dp
				last_ts = dp.Ts
			}
			first = false
		}
		if err := scanner.Err(); err != nil {
			panic(fmt.Sprintf("error reading %s: %s", path, err.Error()))
		}
		if last_ts < until {
			for new_ts := last_ts + 60; new_ts <= until+60; new_ts += 60 {
				our_el.Link <- *metrics.NewDatapoint(new_ts, 0.0, false)
			}
		}
	}(our_el)
	return our_el, nil
}

func (t TextStore) Has(name string) (found bool, err error) {
	_, err = os.Stat(t.path(name))
	return (err == nil), nil
}
func (t TextStore) List() (list []string, err error) {
	file_info, err := ioutil.ReadDir(t.BasePath)
	if err != nil {
		return
	}
	list = make([]string, len(file_info))
	for i, fi := range file_info {
		name := fi.Name()
		list[i] = name[:len(name)-4]
	}
	return
}
