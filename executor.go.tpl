package main
import (
    "fmt"
	"./chains"
    "./functions"
    "./config"
    "./stores"
    "github.com/BurntSushi/toml"
)

func main () {
    from := int32({{.From}})
    until := int32({{.Until}})
    var dep_el chains.ChainEl

    var config config.Main
    if _, err := toml.DecodeFile("graphite-ng.conf", &config); err != nil {
        fmt.Println(err)
        return
    }
    if err := stores.Init(config); err != nil {
        fmt.Println(err)
        return
    }

{{range $i, $e := .Targets}}
{{if $i}}
fmt.Printf("]},\n")
{{else}}
fmt.Printf("[\n")
{{end}}
    dep_el = {{.Cmd}}
    dep_el.Settings <- from
    dep_el.Settings <- until
    fmt.Printf("{\"target\": \"{{.Name}}\", \"datapoints\": [")
    functions.OutPrintStandardJson(dep_el, until)
{{end}}
fmt.Printf("]}\n]")
}
