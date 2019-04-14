package collector

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type synoCollector struct{}

func init() {
	registerCollector("syno", defaultEnabled, NewSynoCollector)
}

func NewSynoCollector() (Collector, error) {
	return &synoCollector{}, nil
}

var (
	deepSleepSupport = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "deep_sleep_support"),
		"TODO",
		[]string{"device"},
		nil,
	)

	diskSerial = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "disk_serial"),
		"TODO",
		[]string{"device", "serial"},
		nil,
	)

	idleTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "idle_time"),
		"TODO",
		[]string{"device"},
		nil,
	)

	pwrResetCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "pwr_reset_count"),
		"TODO",
		[]string{"device"},
		nil,
	)

	spindown = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "spindown"),
		"TODO",
		[]string{"device"},
		nil,
	)

	standbySyncing = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "syno", "standby_syncing"),
		"TODO",
		[]string{"device"},
		nil,
	)
)

func (c *synoCollector) Update(ch chan<- prometheus.Metric) error {
	for _, dev := range listBlockDevs() {
		ch <- prometheus.MustNewConstMetric(
			deepSleepSupport,
			prometheus.GaugeValue,
			float64(intFromFile(dev, "syno_deep_sleep_support")),
			dev,
		)

		ch <- prometheus.MustNewConstMetric(
			diskSerial,
			prometheus.GaugeValue,
			1.0,
			dev,
			stringFromFile(dev, "syno_disk_serial"),
		)

		ch <- prometheus.MustNewConstMetric(
			idleTime,
			prometheus.GaugeValue,
			float64(intFromFile(dev, "syno_idle_time")),
			dev,
		)

		ch <- prometheus.MustNewConstMetric(
			pwrResetCount,
			prometheus.GaugeValue,
			float64(intFromFile(dev, "syno_pwr_reset_count")),
			dev,
		)

		ch <- prometheus.MustNewConstMetric(
			spindown,
			prometheus.GaugeValue,
			float64(intFromFile(dev, "syno_spindown")),
			dev,
		)

		ch <- prometheus.MustNewConstMetric(
			standbySyncing,
			prometheus.GaugeValue,
			float64(intFromFile(dev, "syno_standby_syncing")),
			dev,
		)

	}

	return nil
}

func stringFromFile(dev, name string) string {
	fn := sysFilePath(filepath.Join("block", dev, "device", name))
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Debugf("Error reading %s: %s", fn, err)
		return ""
	}
	return strings.TrimSpace(string(buf))
}

func intFromFile(dev, name string) int {
	sv := stringFromFile(dev, name)
	iv, err := strconv.Atoi(sv)
	if err != nil {
		log.Debugf("Error converting value for %s / %s from %s: %s", dev, name, sv, err)
		return 0
	}
	return iv
}

func synoBlockDevsIdleGt(idleTime int) bool {
	max := 0
	for _, dev := range listBlockDevs() {
		if iv := intFromFile(dev, "syno_idle_time"); iv > max {
			max = iv
		}
	}
	return max > idleTime
}

func synoBlockDevIdleLt(idleTime int) bool {
	for _, dev := range listBlockDevs() {
		if iv := intFromFile(dev, "syno_idle_time"); iv < idleTime {
			return true
		}
	}
	return false
}

func listBlockDevs() []string {
	files, err := ioutil.ReadDir(sysFilePath("block"))
	if err != nil {
		return nil
	}
	blockDevs := make([]string, 0, len(files))
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "sd") {
			continue
		}
		blockDevs = append(blockDevs, f.Name())
	}
	return blockDevs
}
