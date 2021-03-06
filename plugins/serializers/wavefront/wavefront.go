package wavefront

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

// WavefrontSerializer : WavefrontSerializer struct
type WavefrontSerializer struct {
	Prefix         string
	HostTag        string
	SourceOverride []string
}

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
	"=", "-",
)

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", ".")

// Serialize : Serialize based on Wavefront format
func (s *WavefrontSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	out := []byte{}
	metricSeparator := "."
	for fieldName, value := range metric.Fields() {
		var name string

		if fieldName == "value" {
			name = fmt.Sprintf("%s%s", s.Prefix, metric.Name())
		} else {
			name = fmt.Sprintf("%s%s%s%s", s.Prefix, metric.Name(), metricSeparator, fieldName)
		}

		name = sanitizedChars.Replace(name)
		name = pathReplacer.Replace(name)
		timestamp := metric.UnixNano() / 1000000000
		metricValue, buildError := buildValue(value, name)

		if buildError != nil {
			log.Printf("E! Output [wavefront] %s\n", buildError.Error())
			continue
		}

		tagsSlice := buildTags(metric.Tags(), s)
		tags := fmt.Sprint(strings.Join(tagsSlice, " "))
		point := []byte(fmt.Sprintf("%s %s %d %s\n", name, metricValue, timestamp, tags))
		out = append(out, point...)
	}
	return out, nil
}

func buildValue(v interface{}, name string) (string, error) {
	var retv string
	switch p := v.(type) {
	case bool:
		if bool(p) {
			return "1.0", nil
		} else {
			return "0.0", nil
		}
	case int64:
		retv = intToString(int64(p))
	case uint64:
		retv = uintToString(uint64(p))
	case float64:
		retv = floatToString(float64(p))
	default:
		return retv, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	}
	return retv, nil
}

func intToString(inputNum int64) string {
	return strconv.FormatInt(inputNum, 10)
}

func uintToString(inputNum uint64) string {
	return strconv.FormatUint(inputNum, 10)
}

func floatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

func buildTags(mTags map[string]string, s *WavefrontSerializer) []string {
	if s.HostTag == "" {
		s.HostTag = "telegraf_host"
	}
	hosttag := s.HostTag
	sourceTagFound := false

	for _, s := range s.SourceOverride {
		for k, v := range mTags {
			if k == s {
				mTags["source"] = v
				mTags[hosttag] = mTags["host"]
				sourceTagFound = true
				delete(mTags, k)
				break
			}
		}
		if sourceTagFound {
			break
		}
	}

	if !sourceTagFound {
		mTags["source"] = mTags["host"]
	}
	delete(mTags, "host")

	tags := make([]string, len(mTags))
	index := 0
	for k, v := range mTags {
		tags[index] = fmt.Sprintf("%s=\"%s\"", sanitizedChars.Replace(k), tagValueReplacer.Replace(v))
		index++
	}

	sort.Strings(tags)
	return tags
}
