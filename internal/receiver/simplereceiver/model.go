// Unmarcial request json into a struct for receiver consumption

package githubmetricsreceiver

type CommitStats struct {
	Days          []interface{} `json:"days"`
	TotalCommits  int64         `json:"total"`
	WeekStamp     int64         `json:"week"`
}
