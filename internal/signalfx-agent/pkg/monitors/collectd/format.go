package collectd

// JSONWriteBody is the full POST body of collectd's write_http format
//
//easyjson:json
type JSONWriteBody []*JSONWriteFormat

// JSONWriteFormat is the format for collectd json datapoints
//
//easyjson:json
type JSONWriteFormat struct {
	Time           *float64               `json:"time"`
	Host           *string                `json:"host"`
	Interval       *float64               `json:"interval"`
	Plugin         *string                `json:"plugin"`
	PluginInstance *string                `json:"plugin_instance"`
	TypeS          *string                `json:"type"`
	TypeInstance   *string                `json:"type_instance"`
	Message        *string                `json:"message"`
	Meta           map[string]interface{} `json:"meta"`
	Severity       *string                `json:"severity"`
	Dstypes        []*string              `json:"dstypes"`
	Dsnames        []*string              `json:"dsnames"`
	Values         []*float64             `json:"values"`
}
