package collectd

// JSONWriteBody is the full POST body of collectd's write_http format
//
//easyjson:json
type JSONWriteBody []*JSONWriteFormat

// JSONWriteFormat is the format for collectd json datapoints
//
//easyjson:json
type JSONWriteFormat struct {
	Dsnames        []*string  `json:"dsnames"`
	Dstypes        []*string  `json:"dstypes"`
	Host           *string    `json:"host"`
	Interval       *float64   `json:"interval"`
	Plugin         *string    `json:"plugin"`
	PluginInstance *string    `json:"plugin_instance"`
	Time           *float64   `json:"time"`
	TypeS          *string    `json:"type"`
	TypeInstance   *string    `json:"type_instance"`
	Values         []*float64 `json:"values"`
	// events
	Message  *string                `json:"message"`
	Meta     map[string]interface{} `json:"meta"`
	Severity *string                `json:"severity"`
}
