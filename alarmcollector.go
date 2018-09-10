package unifiexporter

import (
	"log"

	"github.com/mdlayher/unifi"
	"github.com/prometheus/client_golang/prometheus"
)

// AlarmCollector is a Prometheus collector for Unifi alarms
type AlarmCollector struct {
	AlarmsTotal *prometheus.Desc
	Alarms      *prometheus.Desc

	c     *unifi.Client
	sites []*unifi.Site
}

// Verify that the Exporter implements the collector interface.
var _ collector = &AlarmCollector{}

// NewAlarmCollector creates a new AlarmCollector
func NewAlarmCollector(c *unifi.Client, sites []*unifi.Site) *AlarmCollector {
	const (
		subsystem = "alarms"
	)

	var (
		labelsSiteOnly = []string{"site"}
		labelsAlarms   = []string{"site", "id", "name", "mac", "key", "message", "subsytem"}
	)

	return &AlarmCollector{
		AlarmsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"Total number of active alarms",
			labelsSiteOnly,
			nil,
		),

		Alarms: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", subsystem),
			"Number of active alarms",
			labelsAlarms,
			nil,
		),

		c:     c,
		sites: sites,
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *AlarmCollector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.AlarmsTotal,
		c.Alarms,
	}

	for _, d := range ds {
		ch <- d
	}
}

// collect begins a metrics collection task for all alarms
func (c *AlarmCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	for _, s := range c.sites {
		alarms, err := c.c.Alarms(s.Name)
		if err != nil {
			return c.Alarms, err
		}

		ch <- prometheus.MustNewConstMetric(
			c.AlarmsTotal,
			prometheus.GaugeValue,
			float64(len(alarms)),
			s.Description,
		)

		c.collectAlarms(ch, s.Description, alarms)
	}

	return nil, nil
}

// collectAlarms collects the details for each alarm
func (c *AlarmCollector) collectAlarms(ch chan<- prometheus.Metric, siteLabel string, alarms []*unifi.Alarm) {
	for _, alarm := range alarms {
		labels := []string{
			siteLabel,
			alarm.ID,
			alarm.APName,
			alarm.APMAC.String(),
			alarm.Key,
			alarm.Message,
			alarm.Subsystem,
		}

		ch <- prometheus.MustNewConstMetric(
			c.Alarms,
			prometheus.GaugeValue,
			float64(len(alarms)),
			labels...,
		)
	}
}

// Collect is the same as CollectError, but ignores any errors which occur.
// Collect exists to satisfy the prometheus.Collector interface.
func (c *AlarmCollector) Collect(ch chan<- prometheus.Metric) {
	_ = c.CollectError(ch)
}

// CollectError sends the metric values for each metric pertaining to the global
// cluster usage over to the provided prometheus Metric channel, returning any
// errors which occur.
func (c *AlarmCollector) CollectError(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		ch <- prometheus.NewInvalidMetric(desc, err)
		log.Printf("[ERROR] failed collecting device metric %v: %v", desc, err)
		return err
	}

	return nil
}
