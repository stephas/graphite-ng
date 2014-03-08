package main

import (
	"./config"
	"./functions"
	"./stack"
	"./stores"
	"./timespec"
	"bytes"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"
	"unicode/utf8"
)

type Target struct {
	Name string
	Cmd  string
}

type Params struct {
	From    int32
	Until   int32
	Targets []Target
}

// FieldsFuncWithDelim is like strings.FieldsFunc except it also returns
// the delimiters
func FieldsFuncWithDelim(s string, f func(r rune) bool) []string {
	var l []string
	var sp int
	for i, r := range s {
		if f(r) {
			if sp < i {
				l = append(l, s[sp:i])
			}
			l = append(l, string(r))
			sp = i + utf8.RuneLen(r)
		}
	}
	if sp < len(s) {
		l = append(l, s[sp:])
	}
	return l
}

// generateCommand parses an input target such as
// "alias(foo(bar baz unit=Mb/s ip=127.0.0.1 qux,12,foo2(5.0, somestr)), my alias name)"
// into the correct golang code, with intermidate tokens like:
// ["alias" "(" "foo" "(" "bar baz unit=Mb/s ip=127.0.0.1 qux" "," "12" "," "foo2"
// "(" "5.0" "," " somestr" ")" ")" "," " my alias name" ")"]
func generateTarget(target_str string) (target Target, err error) {
	tokens := FieldsFuncWithDelim(target_str, func(r rune) bool {
		return r == '(' || r == ')' || r == ','
	})
	target.Name = target_str
	cmd := ""
	in_fn := ""
	arg_no := 0
	prior_arg_no := new(stack.Stack)
	prior_in_fn := new(stack.Stack)
	for i, token := range tokens {
		next := ""
		if i < len(tokens)-1 {
			next = tokens[i+1]
		}
		if next == "(" {
   		// a function is starting
			if in_fn != "" {
				prior_in_fn.Push(in_fn)
			}
			if arg_no != 0 {
				prior_arg_no.Push(arg_no)
			}
			in_fn = token
			arg_no = 0
			if _, ok := functions.Functions[in_fn]; !ok {
				return target, errors.New(fmt.Sprintf("ERROR: invalid syntax. did not recognize function '%s'", in_fn))
			}
			cmd += "functions." + functions.Functions[token][0]
		} else if token == ")" {
			// a function is ending
			// do we need to do any actions right now for certain functions?
			if in_fn == "alias" {
				target.Name = tokens[i-1]
			}
			cmd += ")"
			fn := prior_in_fn.Pop()
			if fn == nil {
				in_fn = ""
			} else {
				in_fn = fn.(string)
			}
			an := prior_arg_no.Pop()
			if an == nil {
				arg_no = 0
			} else {
				arg_no = an.(int)
			}
		} else if token == "(" {
			cmd += "(\n"
		} else if token == "," {
			cmd += ",\n"
			// token is an argument
		} else {
			arg_no += 1
			arg_type := "metric"
			if arg_no < len(functions.Functions[in_fn]) {
				arg_type = functions.Functions[in_fn][arg_no]
			}
			if arg_type == "metric" {
				cmd += "ReadMetric(\"" + token + "\")"
			} else if arg_type == "string" {
				cmd += "\"" + token + "\""
			} else {
				cmd += token
			}
		}
	}
	target.Cmd = cmd
	return
}
func renderJson(targets_list []string, from int32, until int32) string {
	targets := make([]Target, 0)
	for _, target_str := range targets_list {
		target, err := generateTarget(target_str)
		if err != nil {
			return err.Error()
		}
		targets = append(targets, target)
	}
	params := Params{from, until, targets}
	t, err := template.ParseFiles("executor.go.tpl")
	if err != nil {
		panic(err)
	}
	fname := fmt.Sprintf("executor-%d.go", rand.Int())
	fo, err := os.Create(fname)
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	fmt.Println("writing to template", params)
	t.Execute(fo, params)
	// TODO: timeout, display errors, etc
	fmt.Printf("executing: go run %s data.go\n", fname)
	cmd_exec := exec.Command("go", "run", fname, "data.go")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd_exec.Stdout = &stdout
	cmd_exec.Stderr = &stderr
	err = cmd_exec.Run()
	if stderr.Len() > 0 {
		fmt.Println("stdout:", stdout.String())
		fmt.Println("sterr:", stderr.String())
	}
	if err != nil {
		fmt.Println("error:", err)
		return stdout.String() + "\nERRORS:" + stderr.String() + "\n" + err.Error()
	}
	return stdout.String() + "\n" + stderr.String()
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	until := int32(time.Now().Unix())
	from := until - 24*60*60
	r.ParseForm()
	from_list := r.Form["from"]
	if len(from_list) > 0 {
		t, err := timespec.GetTimeStamp(from_list[0])
		if err != nil {
			fmt.Fprintf(w, "Error: invalid 'from' spec: "+from_list[0])
			return
		}
		from = int32(t.Unix())
	}
	until_list := r.Form["until"]
	if len(until_list) > 0 {
		t, err := timespec.GetTimeStamp(until_list[0])
		if err != nil {
			fmt.Fprintf(w, "Error: invalid 'until' spec: "+until_list[0])
			return
		}
		until = int32(t.Unix())
	}
	targets_list := r.Form["target"]
	for _, target := range targets_list {
		if target == "" {
			fmt.Fprintf(w, "invalid request: one or more empty targets")
			return
		}
	}
	if len(targets_list) < 1 {
		fmt.Fprintf(w, "invalid request: no targets requested")
	} else {
		fmt.Fprintf(w, renderJson(targets_list, from, until))
	}
}
func MetricsListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "[")
	prev := false
	for _, store := range stores.List {
		list, err := (*store).List()
		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			for _, metric := range list {
				if prev {
					fmt.Fprintf(w, fmt.Sprintf(",\n\"%s\"", metric))
				} else {
					fmt.Fprintf(w, fmt.Sprintf("\n\"%s\"", metric))
				}
				prev = true
			}
		}
	}
	fmt.Fprintf(w, "]")
}

func main() {
	var config config.Main
	if _, err := toml.DecodeFile("graphite-ng.conf", &config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("registered functions:")
	for k, v := range functions.Functions {
		fmt.Printf("%-20s -> %s\n", k, v)
	}
	fmt.Println("initializing stores")
	if err := stores.Init(config); err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/render", renderHandler)
	http.HandleFunc("/render/", renderHandler)
	http.HandleFunc("/metrics/index.json", MetricsListHandler)
	fmt.Println("listening on", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, nil)
}
