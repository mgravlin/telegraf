package wavefront

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestBuildTags(t *testing.T) {
	var tagtests = []struct {
		ptIn    map[string]string
		outTags []string
	}{
		{
			map[string]string{"one": "two", "three": "four", "host": "testHost"},
			[]string{"one=\"two\"", "source=\"testHost\"", "three=\"four\""},
		},
		{
			map[string]string{"aaa": "bbb", "host": "testHost"},
			[]string{"aaa=\"bbb\"", "source=\"testHost\""},
		},
		{
			map[string]string{"bbb": "789", "aaa": "123", "host": "testHost"},
			[]string{"aaa=\"123\"", "bbb=\"789\"", "source=\"testHost\""},
		},
		{
			map[string]string{"host": "aaa", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "source=\"aaa\""},
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "source=\"i-0123456789\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"instance-id": "i-0123456789", "host": "aaa", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "source=\"i-0123456789\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "hostname": "ccc", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "hostname=\"ccc\"", "source=\"i-0123456789\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"instanceid": "i-0123456789", "host": "aaa", "snmp_host": "ccc", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "snmp_host=\"ccc\"", "source=\"i-0123456789\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"host": "aaa", "snmp_host": "ccc", "dc": "bbb"},
			[]string{"dc=\"bbb\"", "source=\"ccc\"", "telegraf_host=\"aaa\""},
		},
		{
			map[string]string{"Sp%ci@l Chars": "\"g*t repl#ced", "host": "testHost"},
			[]string{"Sp-ci-l-Chars=\"\\\"g-t repl#ced\"", "source=\"testHost\""},
		},
	}
	for _, tt := range tagtests {
		tags := buildTags(tt.ptIn)
		if !reflect.DeepEqual(tags, tt.outTags) {
			t.Errorf("\nexpected\t%+v\nreceived\t%+v\n", tt.outTags, tags)
		}
	}
}

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("cpu.usage.idle 91.500000 %d cpu=\"cpu0\" source=\"realHost\"", now.UnixNano()/1000000000)}
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(91),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := WavefrontSerializer{}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("cpu.usage.idle 91 %d cpu=\"cpu0\" source=\"realHost\"", now.UnixNano()/1000000000)}
	assert.Equal(t, expS, mS)
}

func TestSerializeMetricPrefix(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(91),
	}
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

	s := WavefrontSerializer{Prefix: "telegraf"}
	buf, _ := s.Serialize(m)
	mS := strings.Split(strings.TrimSpace(string(buf)), "\n")
	assert.NoError(t, err)

	expS := []string{fmt.Sprintf("telegraf.cpu.usage.idle 91 %d cpu=\"cpu0\" source=\"realHost\"", now.UnixNano()/1000000000)}
	assert.Equal(t, expS, mS)
}
