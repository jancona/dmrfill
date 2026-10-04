package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRepeaterBookResultMixedTypes(t *testing.T) {
	// RepeaterBook switches fields between strings and numbers without notice.
	data := `{"count":1,"results":[{
		"State ID":"47","Rptr ID":1234,"Frequency":146.94,"Input Freq":"146.34000",
		"PL":"100.0","TSQ":null,"Nearest City":"Nashville","State":"Tennessee",
		"Precise":"1","DMR":"No","DMR Color Code":"","DMR ID":3112345,
		"M17 CAN":7,"Tetra MCC":310,"YSF DSC":true}]}`
	var res RepeaterBookResults
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r := res.Results[0]
	for _, c := range []struct {
		name string
		got  flexString
		want string
	}{
		{"RptrID", r.RptrID, "1234"},
		{"Frequency", r.Frequency, "146.94"},
		{"InputFreq", r.InputFreq, "146.34000"},
		{"TSQ", r.TSQ, ""},
		{"Precise", r.Precise, "1"},
		{"DMRColorCode", r.DMRColorCode, ""},
		{"DMRID", r.DMRID, "3112345"},
		{"M17CAN", r.M17CAN, "7"},
		{"TetraMCC", r.TetraMCC, "310"},
		{"YSFDSC", r.YSFDSC, "true"},
	} {
		if string(c.got) != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if !MatchesRepeaterBook(filter{key: "nearest city", value: []string{"Nashville"}}, reflect.ValueOf(r)) {
		t.Error("filter on Nearest City did not match")
	}
}

func TestFlexStringRejectsObjects(t *testing.T) {
	var f flexString
	if err := json.Unmarshal([]byte(`{"a":1}`), &f); err == nil {
		t.Error("expected error for object value")
	}
}
