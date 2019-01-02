package cpu_speed

import (
	"io/ioutil"
	"strings"
	"strconv"
	"path/filepath"
	"runtime"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	cpuSpeedPath = "/sys/devices/system/cpu"
	perCpuIntStatuses = []string{"scaling_cur_freq",
	  "scaling_min_freq",
	  "scaling_max_freq",
        }
	perCpuStringStatuses = []string{
	  "scaling_governor",
	  "scaling_driver",
	  "energy_performance_preference",
        }
	globalStatuses = []string{"no_turbo",
	  "max_perf_pct",
	  "min_perf_pct",
        }
)

type CpuSpeed struct {
	numberOfCpus int
}

func (cs *CpuSpeed) SampleConfig() string {
	return ""
}

func (cs *CpuSpeed) Description() string {
	return "This plugin queries CPU frequency and governers"
}

func (cs *CpuSpeed) Gather(acc telegraf.Accumulator) error {
	// collect per-CPU stats
	for i := 0; i < cs.numberOfCpus; i++ {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		cpuId := strconv.Itoa(i)
		tags["cpu"] = cpuId

		onlinePath := filepath.Join(cpuSpeedPath, "cpu" + cpuId, "online")
		onlineVal, err := readUintFromFile(onlinePath)
		if err != nil {
			return err
		}
		fields["online"] = onlineVal
		cpufreqPath := filepath.Join(cpuSpeedPath, "cpu" + cpuId, "cpufreq")
		for _, status := range perCpuIntStatuses {
			statusVal, err := readUintFromFile(filepath.Join(cpufreqPath, status))
			if err != nil {
				return err
			}
			fields[status] = statusVal
		}
		for _, status := range perCpuStringStatuses {
			statusVal, err := readStringFromFile(filepath.Join(cpufreqPath, status))
			if err != nil {
				return err
			}
			fields[status] = statusVal
		}
		acc.AddFields("cpu_speed", fields, tags)
	}
	// collect global stats
	globalPath := filepath.Join(cpuSpeedPath, "intel_pstate")
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tags["cpu"] = "overall"
	for _, status := range globalStatuses {
		statusVal, err := readUintFromFile(filepath.Join(globalPath, status))
		if err != nil {
			return err
		}
		fields[status] = statusVal
	}
	acc.AddFields("cpu_speed", fields, tags)

	return nil
}

func readStringFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func readUintFromFile(path string) (uint64, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func init() {
	inputs.Add("cpu_speed", func() telegraf.Input {
		return &CpuSpeed{runtime.NumCPU()}
	})
}
