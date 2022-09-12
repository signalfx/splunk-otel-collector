// includes all data models for unmarshaling/parsing request return bodies.

package githubmetricsreceiver

import "encoding/json"

// the days field is the number of commits per day for each day of the week in the form of
// "Days": [ sunday, monday, tuesday, wednesday, friday, saturday ] where each day is just an int with
// the number of commits for a given day. This isn't even the craziest thing this API does....
type commitStats struct {
	Days          []interface{} `json:"days"`
	TotalCommits  int64         `json:"total"`
	WeekStamp     int64         `json:"week"`
}

// The commit activity api returns an array of the form 
// [
//   weekstamp,
//   insertions,
//   deletions,
// ]
// where each field is an unlabled int. 
type commitActivity struct {
    WeekStamp  int64
    Insertions int64
    Deletions  int64
}

func newCommitActivity(body []byte) (*commitActivity, error) {
    // the github code frequency api endpoint returns a list of lists which is (technically) valid
    // json. Unfortunately you can't really cleanly unmarshal this without defining the slice of slice
    // interface like below. You can't convert the resulting array directly but you can access each individual
    // field and then stick that in a struct which is what this function does. note that the api returns
    // every week for the past year so you have to take the most recent one which is (apparently) always the 
    // last one...
    var parse [][]interface{}
    err := json.Unmarshal(body, &parse)
    if err != nil {
        return nil, err
    }

    lastWeek := parse[len(parse)-1]
    return &commitActivity{
        WeekStamp:  int64(lastWeek[0].(float64)),
        Insertions: int64(lastWeek[1].(float64)),
        Deletions:  int64(lastWeek[2].(float64)),
    }, nil
}
