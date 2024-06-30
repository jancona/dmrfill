package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Interpolate data into a string

var argsRegex = regexp.MustCompile(`\$(\w+)(?:\:(\d+))?|([^$]+)`)

type RepeaterContext interface {
	GetState() string
	GetCity() string
	GetCallsign() string
	GetFrequency() string
}

func ReplaceArgs(in string, c RepeaterContext, tg *TalkGroup) string {
	var b strings.Builder

	sm := argsRegex.FindAllStringSubmatch(in, -1)
	for _, m := range sm {
		if m[3] != "" {
			// copy literal string
			b.WriteString(m[3])
		} else {
			name := m[1]
			var val string
			if c != nil {
				switch name {
				case "callsign":
					val = c.GetCallsign()
				case "city":
					val = c.GetCity()
				case "frequency":
					val = c.GetFrequency()
				case "state":
					val = c.GetState()
				case "band":
					val = band(ToFloat(c.GetFrequency()))
				case "state_code":
					val = c.GetState()
				}
			}
			if tg != nil {
				switch name {
				case "tg_name":
					val = tg.Name
				case "tg_number":
					val = strconv.Itoa(tg.Number)
				case "time_slot":
					val = strconv.Itoa(tg.TimeSlot)
				}
			}
			l, err := strconv.Atoi(m[2])
			if err == nil && l < len(val) {
				val = val[0:l]
			}
			b.WriteString(val)
		}
	}

	return b.String()
}
func band(freq float64) string {
	switch {
	case freq >= 28.0 && freq <= 29.7:
		return "10m"
	case freq >= 50.0 && freq <= 54.0:
		return "6m"
	case freq >= 144.0 && freq <= 148.0:
		return "2m"
	case freq >= 220.0 && freq <= 225.0:
		return "1.25m"
	case freq >= 420.0 && freq <= 450.0:
		return "70cm"
	case freq >= 902.0 && freq <= 928.0:
		return "33cm"
	case freq >= 1240.0 && freq <= 1325.0:
		return "23cm"
	default:
		return "UNK"
	}
}
func hf(freq float64) string {
	switch {
	case freq <= 30.0:
		return "HF"
	case freq > 30.0 && freq <= 300.0:
		return "VHF"
	case freq > 300.0 && freq <= 3000.0:
		return "UHF"
	case freq > 3000.0 && freq <= 30000.0:
		return "SHF"
	default:
		return "UNK"
	}
}

func ToFloat(freq string) float64 {
	f, err := strconv.ParseFloat(freq, 64)
	if err != nil {
		f = 0.0
	}
	return f
}
