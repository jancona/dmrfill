package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// API Doc: https://radioid.net/database/api
// Example: https://radioid.net/api/dmr/repeater/?state=Maine
const radioIDURL = "https://radioid.net/api/dmr/repeater/"

var talkGroupRegex = regexp.MustCompile(`Time Slot # ?(\d) [-=] Group Call (\d+)(\s*[-=]\s*([^"<>]*))?`)
var lastUpdatedRegex = regexp.MustCompile(`Last Update: (\d+-\d+-\d+ \d+:\d+:\d+)`)

// Supported query parameters
var radioIDQueryParamNames = map[string]struct{}{
	"id":        {}, // DMR Repeater ID
	"callsign":  {}, // Repeater callsign
	"city":      {}, // Repeater city
	"state":     {}, // Repeater state / province
	"country":   {}, // Repeater country
	"frequency": {}, // Repeater frequency
	"trustee":   {}, // Trustee callsign
}
var radioIDResultFields = map[string]string{}

// Put the RadioIDResult JSON field names in a map
func init() {
	st := reflect.TypeOf(RadioIDResult{})
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag != "" {
			radioIDResultFields[tag] = field.Name
		} else {
			radioIDResultFields[strings.ToLower(field.Name)] = field.Name
		}
	}
}

func QueryRadioID(filters filterFlags) (*RadioIDResults, error) {
	var base = radioIDURL

	baseURL, err := url.Parse(base)
	if err != nil {
		panic("Error parsing base URL " + base + ": " + err.Error())
	}

	// Filters to be applied on the results
	var resultFilters filterFlags

	// Query params
	params := url.Values{}
	for _, f := range filters {
		_, ok := radioIDQueryParamNames[f.key]
		if ok {
			for _, v := range f.value {
				params.Add(f.key, v)
			}
		} else {
			_, ok := radioIDResultFields[f.key]
			if ok {
				resultFilters = append(resultFilters, f)
			}
		}
	}
	baseURL.RawQuery = params.Encode()
	logVerbose("RadioID URL %s", baseURL.String())
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cache-Control", "max-age=3600") // Cache results for an hour
	resp, err := cachingHttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-From-Cache") == "1" {
		logVerbose("using cached response")
	}
	logVeryVerbose("response status %s", resp.Status)
	logVeryVerbose("response headers: %#v", resp.Header)
	decoder := json.NewDecoder(resp.Body)
	result := RadioIDResults{
		Results: []RadioIDResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	logVerbose("found %d results", result.Count)
	// Do client filtering
	newResults := []RadioIDResult{}
	for _, r := range result.Results {
		f, err := strconv.ParseFloat(r.Frequency, 64)
		if err == nil {
			r.Band = band(f)
		}
		rv := reflect.ValueOf(r)
		matchesAll := true
		for _, filter := range resultFilters {
			if !MatchesRadioID(filter, rv) {
				logVeryVerbose("Repeater %v doesn't match filter %v", r, filter)
				matchesAll = false
				break
			}
		}
		if matchesAll {
			// Parse LastUpdated
			m := lastUpdatedRegex.FindAllStringSubmatch(r.Details, -1)
			if len(m) > 0 {
				r.LastUpdated, err = time.Parse("2006-01-02 15:04:05", m[0][1])
				if err != nil {
					return nil, fmt.Errorf("error parsing LastUpdated %s: %v", m[0][1], err)
				}
			}
			// Parse talk groups from details field
			var detailsTGs []TalkGroup
			m = talkGroupRegex.FindAllStringSubmatch(r.Details, -1)
			for _, s := range m {
				id, err := strconv.Atoi(s[2])
				if err != nil {
					return nil, fmt.Errorf("error parsing TalkGroup ID %s: %v", s[2], err)
				}
				ts, err := strconv.Atoi(s[1])
				if err != nil {
					return nil, fmt.Errorf("error parsing TalkGroup TimeSlot %s: %v", s[1], err)
				}
				name := s[4]
				if len(name) > nameLength {
					name = name[:nameLength]
				}
				if ts == 1 || ts == 2 {
					detailsTGs = append(detailsTGs, TalkGroup{
						Number:   id,
						TimeSlot: ts,
						Name:     name,
					})
				} else {
					logVerbose("Skipping details talkgroup #%d: %s, bad timeslot %d", id, name, ts)
				}
			}
			logVerbose("Count - detailsTGs: %d", len(detailsTGs))
			// logVerbose("detailsTGs: %#v", detailsTGs)
			r.TalkGroups = detailsTGs
			if !talkgroupsRequired || len(r.TalkGroups) > 0 {
				newResults = append(newResults, r)
			} else {
				logVerbose("Skipping repeater with no talkgroups: %v", r)
			}
		}
	}
	result.Count = len(newResults)
	result.Results = newResults
	logVerbose("%d results after filtering", result.Count)
	// pretty.Println(result)
	return &result, nil
}

func MatchesRadioID(filter filter, rv reflect.Value) bool {
	val := rv.FieldByName(radioIDResultFields[filter.key]).String()
	for _, fv := range filter.value {
		if fv == val {
			return true
		}
	}
	return false
}

type RadioIDResults struct {
	Count   int
	Results []RadioIDResult
}

type RadioIDResult struct {
	Callsign       string `json:"callsign"`        // "KC1FRJ"
	City           string `json:"city"`            // "Presque Isle"
	ColorCode      int    `json:"color_code"`      // 12
	Country        string `json:"country"`         // "United States"
	Details        string `json:"details"`         // "Time Slot #1 - Group Call 759 = SKYWARN\u003Cbr\u003ETime Slot #1 - Group Call 9998 = Parrot*\u003Cbr\u003ETime Slot #1 - Group Call 1 = World Wide*\u003Cbr\u003ETime Slot #1 - Group Call 13 = WW English*\u003Cbr\u003ETime Slot #1 - Group Call 3 = North America\u003Cbr\u003ETime Slot #1 - Group Call 3172 = Northeast\u003Cbr\u003ETime Slot #1 - Group Call 310 = TAC310*\u003Cbr\u003ETime Slot #1 - Group Call 311 = TAC311*\u003Cbr\u003ETime Slot #1 - Group Call 113 = UA English 1*\u003Cbr\u003ETime Slot #1 - Group Call 123 = UA English 2*\u003Cbr\u003ETime Slot #1 - Group Call 8801 = NETAC 1*\u003Cbr\u003E------------------------------------\u003Cbr\u003ETime Slot #2 - Group Call 8802 = NETAC 2*\u003Cbr\u003ETime Slot #2 - Group Call 3181 = New England Wide\u003Cbr\u003ETime Slot #2 - Group Call 8 = Region North\u003Cbr\u003ETime Slot #2 - Group Call 3133 = NH Statewide\u003Cbr\u003ETime Slot #2 - Group Call 3123 = ME Statewide\u003Cbr\u003ETime Slot #1 - Group Call 3029 = New Brunswick\u003Cbr\u003ETime Slot #2 - Group Call 9 = Local Site\u003Cbr\u003E\u003Cbr\u003E* PTT Activated\u003Cbr\u003E\u003Cbr\u003EYou Must Have [ARS] Disabled Within Your Radio\u003Cbr\u003EContact: Dave, KQ1L\u003Cbr\u003EEmail: dhawke@gwi.net\u003Cbr\u003EWebsite: http://nedecn.org"
	Frequency      string `json:"frequency"`       // "145.18000"
	ID             int    `json:"id"`              // 310198
	IPSCNetwork    string `json:"ipsc_network"`    // "NEDECN"
	Offset         string `json:"offset"`          // "-0.600"
	RfinderDetails int    `json:"rfinder_details"` // 0
	State          string `json:"state"`           // "Maine"
	Trustee        string `json:"trustee"`         // "KC1FRJ"
	TSLinked       string `json:"ts_linked"`       // "TS1 TS2"
	LastUpdated    time.Time
	Band           string
	TalkGroups     []TalkGroup
}

func (r RadioIDResult) GetCallsign() string {
	return r.Callsign
}
func (r RadioIDResult) GetCity() string {
	return r.City
}
func (r RadioIDResult) GetFrequency() string {
	return r.Frequency
}
func (r RadioIDResult) GetState() string {
	return r.State
}

type TalkGroup struct {
	Number   int
	TimeSlot int
	Name     string
}
