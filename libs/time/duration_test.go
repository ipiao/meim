package time

import (
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestDurationText(t *testing.T) {
	var (
		input  = []byte("10s")
		output = time.Second * 10
		d      Duration
	)
	if err := d.UnmarshalText(input); err != nil {
		t.FailNow()
	}
	if int64(output) != int64(d) {
		t.FailNow()
	}
}

func TestDurationYaml(t *testing.T) {
	var (
		input  = []byte("10s")
		output = time.Second * 10
		d      Duration
	)
	if err := yaml.Unmarshal(input, &d); err != nil {
		t.FailNow()
	}
	if int64(output) != int64(d) {
		t.FailNow()
	}
}
