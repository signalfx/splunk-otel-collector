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
	DropwizardTagPathsMap        map[string]string `yaml:"dropwizardTagPathsMap"`
	DefaultTags                  map[string]string `yaml:"defaultTags"`
	DropwizardTimeFormat         string            `yaml:"dropwizardTimeFormat"`
	CollectdSecurityLevel        string            `yaml:"collectdSecurityLevel"`
	DataFormat                   string            `yaml:"dataFormat" default:"influx"`
	DropwizardTagsPath           string            `yaml:"dropwizardTagsPath"`
	JSONNameKey                  string            `yaml:"JSONNameKey"`
	JSONQuery                    string            `yaml:"JSONQuery"`
	JSONTimeKey                  string            `yaml:"JSONTimeKey"`
	JSONTimeFormat               string            `yaml:"JSONTimeFormat"`
	Separator                    string            `yaml:"separator"`
	MetricName                   string            `yaml:"metricName"`
	CollectdAuthFile             string            `yaml:"collectdAuthFile"`
	GrokTimeZone                 string            `yaml:"grokTimezone"`
	CSVTimestampFormat           string            `yaml:"CSVTimestampFormat"`
	CollectdSplit                string            `yaml:"collectdSplit"`
	DropwizardMetricRegistryPath string            `yaml:"dropwizardMetricRegistryPath"`
	DropwizardTimePath           string            `yaml:"dropwizardTimePath"`
	CSVTimestampColumn           string            `yaml:"CSVTimestampColumn"`
	CSVMeasurementColumn         string            `yaml:"CSVMeasurementColumn"`
	DataType                     string            `yaml:"dataType"`
	CSVComment                   string            `yaml:"CSVComment"`
	CSVDelimiter                 string            `yaml:"CSVDelimiter"`
	GrokCustomPatterns           string            `yaml:"grokCustomPatterns"`
	Templates                    []string          `yaml:"templates"`
	CollectdTypesDB              []string          `yaml:"collectdTypesDB"`
	GrokNamedPatterns            []string          `yaml:"grokNamedPatterns"`
	GrokPatterns                 []string          `yaml:"grokPatterns"`
	GrokCustomPatternFiles       []string          `yaml:"grokCustomPatternFiles"`
	CSVColumnNames               []string          `yaml:"CSVColumnNames"`
	CSVColumnTypes               []string          `yaml:"CSVColumnTypes"`
	CSVTagColumns                []string          `yaml:"CSVTagColumns"`
	JSONStringFields             []string          `yaml:"JSONStringFields"`
	TagKeys                      []string          `yaml:"JSONTagKeys"`
	CSVHeaderRowCount            int               `yaml:"CSVHeaderRowCount"`
	CSVSkipRows                  int               `yaml:"CSVSkipRows"`
	CSVSkipColumns               int               `yaml:"CSVSkipColumns"`
	CSVTrimSpace                 bool              `yaml:"CSVTrimSpace"`
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
