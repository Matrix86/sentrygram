package pluginmanager

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	"github.com/wcharczuk/go-chart"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (pm *PluginManager) SetInitialDefines() {
	plugin.Defines = map[string]interface{}{
		"MessageType": 0,
		"ImageType":   1,
		"log": func(s string) interface{} {
			log.Info(s)
			return nil
		},
		"exec": func(cmd string) interface{} {
			return CommandExec(cmd)
		},
		"readFile": func(filename string, binary bool) interface{} {
			dat, err := ioutil.ReadFile(filename)
			if err != nil {
				log.Error("readFile: %s", err)
			}
			if binary {
				return dat
			} else {
				return string(dat)
			}
		},
		"cpuUsage": func() interface{} {
			seconds, _ := time.ParseDuration("1s")
			cpus, err := cpu.Percent(seconds, true)
			if err != nil {
				log.Error("cpuUsage: %s", err)
				return []float64{}
			}
			return cpus
		},
		"getProcesses": func() interface{} {
			processes, err := process.Processes()
			if err != nil {
				log.Error("getProcesses: %s", err)
				return err
			}

			var procs []struct {
				Name string
				Cpu  float64
				Pid  int32
				Mem  float64
			}
			for _, p := range processes {
				name := ""
				name, _ = p.Exe()
				cpu, _ := p.CPUPercent()
				pid := p.Pid
				mem, _ := p.MemoryPercent()
				proc := struct {
					Name string
					Cpu  float64
					Pid  int32
					Mem  float64
				}{Name: name, Cpu: float64(cpu), Pid: pid, Mem: float64(mem)}
				procs = append(procs, proc)
			}
			return procs
		},
		"newBarGraph": func(title string, values []float64, labels []string) interface{} {
			var bars []chart.Value

			if len(values) != len(labels) {
				log.Error("newBarGraph: different length between values and labels")
				return ""
			}
			for i, v := range values {
				bars = append(bars, chart.Value{Value: v, Label: labels[i]})
			}
			sbc := chart.BarChart{
				Title:      title,
				TitleStyle: chart.StyleShow(),
				Background: chart.Style{
					Padding: chart.Box{
						Top: 40,
					},
				},
				Height:   512,
				BarWidth: 60,
				XAxis:    chart.StyleShow(),
				YAxis: chart.YAxis{
					Style: chart.StyleShow(),
				},
				Bars: bars,
			}
			buf := bytes.NewBuffer([]byte{})

			if err := sbc.Render(chart.PNG, buf); err != nil {
				log.Error("newBarGraph: %s", err)
				return ""
			}
			rand.Seed(time.Now().UnixNano())
			rnd := RandString(16)
			if err := ioutil.WriteFile("/tmp/"+rnd, buf.Bytes(), 0644); err != nil {
				log.Error("newBarGraph: %s", err)
				return ""
			}

			return "/tmp/" + rnd
		},
	}
}

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func CommandExec(cmd string) string {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1:]...).Output()
	if err != nil {
		log.Error("CommandExec: %s (%s)", err, cmd)
	}
	return string(out)
}
