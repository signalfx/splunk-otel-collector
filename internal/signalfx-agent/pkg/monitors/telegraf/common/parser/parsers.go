package parser

import (
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/ulule/deepcopier"
)

// Config implements Telegraf parsers.Config, but with SignalFx Smart Agent struct tags
// and a methods for returning a Telegraf parsers.Config struct and a Telegraf parsers.Parser.
// Please refer to Telegraf's documentation for more information about the different parsers
// and their specific configurations
type Config struct {
	// dataFormat specifies a data format to parse: `json`, `value`, `influx`, `graphite`, `value`, `nagios`,
	// `collectd`, `dropwizard`, `wavefront`, `grok`, `csv`, or `logfmt`.
	DataFormat string `yaml:"dataFormat" default:"influx"`

	// defaultTags are tags that will be added to all metrics. (`json`, `value`, `graphite`, `collectd`, `dropwizard`,
	// `wavefront`, `grok`, `csv` and `logfmt` only)
	DefaultTags map[string]string `yaml:"defaultTags"`

	// metricName applies to (`json` and `value`). This will be the name of the measurement.
	MetricName string `yaml:"metricName"`

	// value

	// dataType specifies the value type to parse the value to: `integer`, `float`,
	// `long`, `string`, or `boolean`. (`value` only)
	DataType string `yaml:"dataType"`

	// json

	// A list of tag names to fetch from JSON data. (`json` only)
	TagKeys []string `yaml:"JSONTagKeys"`
	// A list of fields in JSON to extract and use as string fields. (json only)
	JSONStringFields []string `yaml:"JSONStringFields"`
	// A path used to extract the metric name in JSON data.  (`json` only)
	JSONNameKey string `yaml:"JSONNameKey"`
	// A gjson path for json parser. (`json` only)
	JSONQuery string `yaml:"JSONQuery"`
	// The name of the timestamp key. (`json` only)
	JSONTimeKey string `yaml:"JSONTimeKey"`
	// Specifies the timestamp format. (`json` only)
	JSONTimeFormat string `yaml:"JSONTimeFormat"`

	// graphite

	// Separator for Graphite data. (`graphite` only).
	Separator string `yaml:"separator"`
	// A list of templates for Graphite data. (`graphite` only).
	Templates []string `yaml:"templates"`

	// collectd

	// The path to the collectd authentication file (`collectd` only)
	CollectdAuthFile string `yaml:"collectdAuthFile"`
	// Specifies the security level: `none` (default), `sign`, or
	// `encrypt`. (`collectd only`)
	CollectdSecurityLevel string `yaml:"collectdSecurityLevel"`
	// A list of paths to collectd TypesDB files. (`collectd` only)
	CollectdTypesDB []string `yaml:"collectdTypesDB"`
	// Indicates whether to separate or join multivalue metrics. (`collectd` only)
	CollectdSplit string `yaml:"collectdSplit"`

	// dropwizard

	// An optional gjson path used to locate a metric registry inside of JSON data.
	// The default behavior is to consider the entire JSON document. (`dropwizard` only)
	DropwizardMetricRegistryPath string `yaml:"dropwizardMetricRegistryPath"`
	// An optional gjson path used to identify the drop wizard metric timestamp. (`dropwizard` only)
	DropwizardTimePath string `yaml:"dropwizardTimePath"`
	// The format used for parsing the drop wizard metric timestamp.
	// The default format is time.RFC3339. (`dropwizard` only)
	DropwizardTimeFormat string `yaml:"dropwizardTimeFormat"`
	// An optional gjson path used to locate drop wizard tags. (`dropwizard` only)
	DropwizardTagsPath string `yaml:"dropwizardTagsPath"`
	// A map of gjson tag names and gjson paths used to extract tag values from the JSON document.
	// This is only used if `dropwizardTagsPath` is not specified. (`dropwizard` only)
	DropwizardTagPathsMap map[string]string `yaml:"dropwizardTagPathsMap"`

	// grok

	// A list of patterns to match. (`grok` only)
	GrokPatterns []string `yaml:"grokPatterns"`
	// A list of named grok patterns to match.  (`grok` only)
	GrokNamedPatterns []string `yaml:"grokNamedPatterns"`
	// Custom grok patterns. (`grok` only)
	GrokCustomPatterns string `yaml:"grokCustomPatterns"`
	// List of paths to custom grok pattern files. (`grok` only)
	GrokCustomPatternFiles []string `yaml:"grokCustomPatternFiles"`
	// Specifies the timezone.  The default is UTC time.  Other options are `Local` for the
	// local time on the machine, `UTC`, and `Canada/Eastern` (unix style timezones).  (`grok` only)
	GrokTimeZone string `yaml:"grokTimezone"`

	//csv

	// The delimiter used between fields in the csv. (`csv` only)
	CSVDelimiter string `yaml:"CSVDelimiter"`
	// The character used to mark rows as comments. (`csv` only)
	CSVComment string `yaml:"CSVComment"`
	// Indicates whether to trim leading white from fields. (`csv` only)
	CSVTrimSpace bool `yaml:"CSVTrimSpace"`
	// List of custom column names.  All columns must have names.  Unnamed columns are ignored.
	// This configuration must be set when `CSVHeaderRowCount` is 0. (`csv` only)
	CSVColumnNames []string `yaml:"CSVColumnNames"`
	// List of types to assign to columns.  Acceptable values are `int`,
	// `float`, `bool`, or `string` (`csv` only).
	CSVColumnTypes []string `yaml:"CSVColumnTypes"`
	// List of columns that should be added as tags.  Unspecified columns will be added as fields. (`csv` only)
	CSVTagColumns []string `yaml:"CSVTagColumns"`
	//  The name of the column to extract the metric name from (`csv` only)
	CSVMeasurementColumn string `yaml:"CSVMeasurementColumn"`
	// The name of the column to extract the metric timestamp from.
	// `CSVTimestampFormat` must be set when using this option.  (`csv` only)
	CSVTimestampColumn string `yaml:"CSVTimestampColumn"`
	//  The format to use for extracting timestamps. (`csv` only)
	CSVTimestampFormat string `yaml:"CSVTimestampFormat"`
	// The number of rows that are headers.  By default no rows are treated as headers.  (`csv` only)
	CSVHeaderRowCount int `yaml:"CSVHeaderRowCount"`
	// The number of rows to ignore before looking for headers. (`csv` only)
	CSVSkipRows int `yaml:"CSVSkipRows"`
	// The number of columns to ignore before parsing data on a given row. (`csv` only)
	CSVSkipColumns int `yaml:"CSVSkipColumns"`
}

// GetTelegrafConfig returns the configuration as a Telegraf *parsers.Config
func (c *Config) GetTelegrafConfig() (config *parsers.Config, err error) {
	config = &parsers.Config{}
	// copy top level struct fields to the
	err = deepcopier.Copy(c).To(config)
	return config, err
}

// GetTelegrafParser returns a pointer to a telegraf *parsers.Parser
func (c *Config) GetTelegrafParser() (parser parsers.Parser, err error) {
	var config *parsers.Config
	if config, err = c.GetTelegrafConfig(); err != nil {
		return parser, err
	}
	return parsers.NewParser(config)
}
