package infiniband

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	infinibandPath = "/sys/class/infiniband"
	statuses = []string{"link_layer", "phys_state", "rate", "state"}
	counters = []string{"port_rcv_data", "port_xmit_data"}
)

type Infiniband struct {
}

func (ib *Infiniband) SampleConfig() string {
	return ""
}

func (ib *Infiniband) Description() string {
	return "This plugin queries InfiniBand metrics"
}

func (ib *Infiniband) Gather(acc telegraf.Accumulator) error {
	devices, err := lookupInfinibandDevices()
	if err != nil {
		return err
	}
	for _, device := range devices {
		ports, err := lookupInfinibandPorts(device)
		if err != nil {
			return err
		}
		for _, port := range ports {
			err := queryPort(device, port, acc)
			if err != nil {
				return err
			}
		}
	}
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

func queryPort(device string, port string, acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tags["device"] = device
	tags["port"] = port

	// query basic status
	statusDir := filepath.Join(infinibandPath, device, "ports", port)
	for _, status := range statuses {
		statusFile := filepath.Join(statusDir, status)
		statusVal, err := readStringFromFile(statusFile)
		if err != nil {
			return err
		}
		fields[status] = statusVal
	}

	// query counters
	counterDir := filepath.Join(statusDir, "counters")
	for _, counter := range counters {
		counterFile := filepath.Join(counterDir, counter)
		counterVal, err := readUintFromFile(counterFile)
		if err != nil {
			return err
		}
		counterVal *= 4
		fields[counter] = counterVal
	}
	acc.AddFields("infiniband", fields, tags)

	return nil
}

func lookupInfinibandDevices() ([]string, error) {
	devices, err := filepath.Glob(filepath.Join(infinibandPath, "/*"))
	if err != nil {
		return nil, err
	}

	// check number of devices >= 1

	for i, device := range devices {
		devices[i] = filepath.Base(device)
	}

	return devices, nil
}

func lookupInfinibandPorts(device string) ([]string ,error) {
	ports, err := filepath.Glob(filepath.Join(infinibandPath, device, "ports/*"))
	if err != nil {
		return nil, err
	}

	for i, port := range ports {
		ports[i] = filepath.Base(port)
	}

	return ports, nil
}

func init() {
	inputs.Add("infiniband", func() telegraf.Input {
		return &Infiniband{}
	})
}
