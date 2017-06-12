package serializers

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/plugins/serializers/wavefront"
)

// SerializerOutput is an interface for output plugins that are able to
// serialize telegraf metrics into arbitrary data formats.
type SerializerOutput interface {
	// SetSerializer sets the serializer function for the interface.
	SetSerializer(serializer Serializer)
}

// Serializer is an interface defining functions that a serializer plugin must
// satisfy.
type Serializer interface {
	// Serialize takes a single telegraf metric and turns it into a byte buffer.
	// separate metrics should be separated by a newline, and there should be
	// a newline at the end of the buffer.
	Serialize(metric telegraf.Metric) ([]byte, error)
}

// Config is a struct that covers the data types needed for all serializer types,
// and can be used to instantiate _any_ of the serializers.
type Config struct {
	// Dataformat can be one of: influx, graphite, or json
	DataFormat string

	// Prefix to add to all measurements, only supports Graphite
	Prefix string

	// Template for converting telegraf metrics into Graphite
	// only supports Graphite
	Template string

	// Timestamp units to use for JSON formatted output
	TimestampUnits time.Duration

	// Point tags to use as the source name for Wavefront (if none found,
	// host will be used). Only supports Wavefront
	SourceOverride []string

	// If the source is overridden, this name will be used for host tag
	// Only supports Wavefront
	HostTag string
}

// NewSerializer a Serializer interface based on the given config.
func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "influx":
		serializer, err = NewInfluxSerializer()
	case "graphite":
		serializer, err = NewGraphiteSerializer(config.Prefix, config.Template)
	case "json":
		serializer, err = NewJsonSerializer(config.TimestampUnits)
	case "wavefront":
		serializer, err = NewWavefrontSerializer(config.Prefix, config.HostTag, config.SourceOverride)
	default:
		err = fmt.Errorf("Invalid data format: %s", config.DataFormat)
	}
	return serializer, err
}

func NewJsonSerializer(timestampUnits time.Duration) (Serializer, error) {
	return &json.JsonSerializer{TimestampUnits: timestampUnits}, nil
}

func NewInfluxSerializer() (Serializer, error) {
	return &influx.InfluxSerializer{}, nil
}

func NewGraphiteSerializer(prefix, template string) (Serializer, error) {
	return &graphite.GraphiteSerializer{
		Prefix:   prefix,
		Template: template,
	}, nil
}

func NewWavefrontSerializer(prefix string, hostTag string, sourceOverride []string) (Serializer, error) {
	return &wavefront.WavefrontSerializer{
		Prefix:         prefix,
		HostTag:        hostTag,
		SourceOverride: sourceOverride,
	}, nil
}
