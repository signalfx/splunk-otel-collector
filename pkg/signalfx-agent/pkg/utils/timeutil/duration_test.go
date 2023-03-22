package timeutil

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/signalfx/defaults"

	"gopkg.in/yaml.v2"
)

func TestSecondsOrDuration_UnmarshalYAML(t *testing.T) {
	type args struct {
		yml string
	}
	tests := []struct {
		name    string
		d       time.Duration
		args    args
		wantErr bool
	}{
		{"5 seconds as integer", 5 * time.Second, args{"time: 5"}, false},
		{"10 seconds as string", 10 * time.Second, args{"time: '10'"}, false},
		{"15 seconds as duration", 15 * time.Second, args{"time: '15s'"}, false},
		{"invalid syntax", 0, args{"time: invalid"}, true},
		{"empty string", 0, args{"time:''"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out map[string]Duration
			if err := yaml.Unmarshal([]byte(tt.args.yml), &out); (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if out["time"].AsDuration() != tt.d {
				t.Fatalf("got %v, wanted %v", out["time"], tt.d)
			}
		})
	}
}

func TestSecondsOrDuration_Defaults(t *testing.T) {
	type args struct {
		yml string
	}
	tests := []struct {
		name    string
		d       time.Duration
		args    args
		wantErr bool
	}{
		{"5 seconds as integer", 5 * time.Second, args{"{}"}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out struct {
				Interval Duration `yaml:"interval" default:"5s"`
			}
			if err := yaml.Unmarshal([]byte(tt.args.yml), &out); (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := defaults.Set(&out); err != nil {
				t.Fatal(err)
			}
			if out.Interval.AsDuration() != tt.d {
				t.Errorf("got %v, wanted %v", out.Interval, tt.d)
			}
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	type args struct {
		json string
	}
	tests := []struct {
		name    string
		d       time.Duration
		args    args
		wantErr bool
	}{
		{"5 seconds as integer", 5 * time.Second, args{`{"time": 5}`}, false},
		{"10 seconds as string", 10 * time.Second, args{`{"time": "10"}`}, false},
		{"15 seconds as duration", 15 * time.Second, args{`{"time": "15s"}`}, false},
		{"invalid syntax", 0, args{`{"time": "invalid"}`}, true},
		{"empty string", 0, args{"time:''"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out map[string]Duration
			if err := json.Unmarshal([]byte(tt.args.json), &out); (err != nil) != tt.wantErr {
				t.Fatalf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if out["time"].AsDuration() != tt.d {
				t.Fatalf("got %v, wanted %v", out["time"], tt.d)
			}
		})
	}
}
