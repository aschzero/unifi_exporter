package unifiexporter

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mdlayher/unifi"
)

func TestAlarmCollector(t *testing.T) {
	var tests = []struct {
		desc    string
		input   string
		sites   []*unifi.Site
		matches []*regexp.Regexp
	}{
		{
			desc: "one alarm, one site",
			input: strings.TrimSpace(`
{
	"data": [
		{
			"_id": "abcdef",
			"ap": "de:ad:be:ef:de:ad",
			"ap_name": "foo",
			"archived": false,
			"datetime": "2018-09-08T18:56:12Z",
			"key": "EVT_AP_Lost_Contact",
			"msg": "AP[de:ad:be:ef:de:ad] was disconnected",
			"subsystem": "wlan",
			"time": 1536432972388
		}
	]
}
`),
			matches: []*regexp.Regexp{
				regexp.MustCompile(`unifi_alarms_total{site="Default"} 1`),

				regexp.MustCompile(`unifi_alarms{id="abcdef",key="EVT_AP_Lost_Contact",mac="de:ad:be:ef:de:ad",message="AP\[de:ad:be:ef:de:ad\] was disconnected",name="foo",site="Default",subsytem="wlan"} 1`),
			},
			sites: []*unifi.Site{{
				Name:        "default",
				Description: "Default",
			}},
		},
	}

	for i, tt := range tests {
		t.Logf("[%02d] test %q", i, tt.desc)

		out := testAlarmCollector(t, []byte(tt.input), tt.sites)

		for j, m := range tt.matches {
			t.Logf("\t[%02d:%02d] match: %s", i, j, m.String())

			if !m.Match(out) {
				t.Fatal("\toutput failed to match regex")
			}
		}
	}
}

func testAlarmCollector(t *testing.T, input []byte, sites []*unifi.Site) []byte {
	c, done := testUniFiClient(t, input)
	defer done()

	collector := NewAlarmCollector(
		c,
		sites,
	)

	return testCollector(t, collector)
}
