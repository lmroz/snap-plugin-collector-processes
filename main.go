/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"io/ioutil"
	"os"
	"unicode"

	// "github.com/intelsdi-x/snap-plugin-collector-processes/processes"
	// "github.com/intelsdi-x/snap/control/plugin"

	"fmt"
	. "github.com/ahmetb/go-linq"
	"github.com/davecgh/go-spew/spew"
	"path/filepath"
	"strconv"
	"strings"
)

var statLabels = []string{
	"pid", "comm", "state", "ppid", "pgrp", "session", "tty_nr", "tpgid", "flags",
	"minflt", "cminflt", "majflt", "cmajflt", "utime", "stime", "cutime", "cstime",
	"priority", "nice", "num_threads", "itrealvalue", "starttime", "vsize", "rss",
	"rsslim", "startcode", "endcode", "startstack", "kstkesp", "kstkeip", "signal",
	"blocked", "sigignore", "sigcatch", "wchan", "nswap", "cnswap", "exit_signal", "processor",
	"rt_priority", "policy", "delayacct_blkio_ticks", "guest_time", "cguest_time", "start_data",
	"end_data", "start_brk", "arg_start", "arg_end", "env_start", "env_end", "exit_code",
}

var stateMap = map[string]string{
	"R": "running",
	"S": "sleeping",
	"D": "waiting",
	"Z": "zombie",
	"T": "traced",
	"W": "paging",
}

func parseInt(s string) (interface{}, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return nil, err
			}
			return v, nil
		} else {
			return nil, err
		}
	}
	return v, nil
}

func readStat(file string) map[string]interface{} {
	procFile, err := ioutil.ReadFile(filepath.Join(file, "stat"))
	if err != nil {
		panic(err)
	}
	fields := strings.Fields(string(procFile))
	res := map[string]interface{}{}
	for i, v := range fields {
		if i >= len(statLabels) {
			break
		}
		var val interface{}
		if i == 2 {
			if v2, ok := stateMap[v]; ok {
				res[statLabels[i]] = v2
			} else {
				res[statLabels[i]] = v
			}
			continue
		}
		if i == 1 {
			val = v
		} else {
			val, err = parseInt(v)
			if err != nil {
				panic(err)
			}

		}
		res[statLabels[i]] = val
	}
	res["count"] = 1
	return res
}

func foo(bar []map[string]interface{}, key, elems []string) {
	type kev struct {
		key, elem string
		value     interface{}
	}

	nop := func(s string) string { return s }

	req := From(bar).SelectManyT(func(v map[string]interface{}) Query {
		keyLookup := make([]string, len(key))
		for i, k := range key {
			keyLookup[i] = fmt.Sprint(v[k])
		}
		return From(elems).GroupJoinT(From(v), nop, func(kv KeyValue) string {
			return kv.Key.(string)
		}, func(el string, kvs []KeyValue) string {
			return strings.Join(keyLookup, "#") + "/" + fmt.Sprint(el) + ":" + fmt.Sprint(kvs)
		})
	}).Results()
	spew.Dump(req)
}

func main() {
	procFiles, err := ioutil.ReadDir("/proc")
	if err != nil {
		panic(err)
	}
	files := From(procFiles).WhereT(func(f os.FileInfo) bool {
		if !f.IsDir() {
			return false
		}
		return FromString(f.Name()).AllT(unicode.IsDigit)
	}).SelectT(os.FileInfo.Name).Results()
	spew.Dump(files)

	fmt.Println("----------------")
	ppp1 := readStat("/proc/1077")
	ppp2 := readStat("/proc/1071")
	ppp := []map[string]interface{}{ppp1, ppp2}
	foo(ppp, []string{"state", "pid"}, []string{"count", "comm"})
	// qqq := From(ppp).SelectManyT(func(v map[string]interface{}) Query {
	// 	return From(v).SelectT(func(kv KeyValue) string {
	// 		return fmt.Sprintln(v["state"], kv.Key, ":", kv.Value)
	// 	})
	// }).

	// 	// ).SelectT(func (g Group) Query {
	// 	// 	type accum map[string][]interface{}
	// 	// 	return From(g.Group).AggregateWithSeedT(accum{}, func(a accum, map[string]interface{} ){
	// 	// 		for k,v :=
	// 	// 	})

	// 	Results()
	// spew.Dump(qqq)

	// return
	// procPlg := processes.New()
	// if procPlg == nil {
	// 	panic("Failed to initialize plugin\n")
	// }

	// plugin.Start(
	// 	processes.Meta(),
	// 	procPlg,
	// 	os.Args[1],
	// )
}
