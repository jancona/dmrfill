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

var talkGroupRegex = regexp.MustCompile(`Time Slot # (\d) [-=] Group Call (\d+)(\s*[-=]\s*([^"<>]*))?`)
var lastUpdatedRegex = regexp.MustCompile(`Last Update: (\d+-\d+-\d+ \d+:\d+:\d+)`)

// Supported query parameters
var queryParamNames = map[string]struct{}{
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
		}
	}
}

func QueryRadioID(cp *Codeplug, filters filterFlags) (*RadioIDResults, error) {
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
		_, ok := queryParamNames[f.key]
		if ok {
			fmt.Printf("f: %#v\n", f)
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
	fmt.Println(baseURL.String())
	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "dmrfill/0.1 github.com/jancona/dmrfill https://www.qrz.com/db/N1ADJ")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	fmt.Printf("response status %s\n", resp.Status)
	decoder := json.NewDecoder(resp.Body)
	result := RadioIDResults{
		Results: []RadioIDResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	if len(resultFilters) > 0 {
		// Do client filtering
		newResults := []RadioIDResult{}
		for _, r := range result.Results {
			rv := reflect.ValueOf(r)
			matchesAll := true
			for _, filter := range resultFilters {
				if !Matches(filter, rv) {
					matchesAll = false
					break
				}
			}
			if matchesAll {
				// Parse LastUpdated
				m := lastUpdatedRegex.FindAllStringSubmatch(r.Rfinder, -1)
				if len(m) > 0 {
					r.LastUpdated, err = time.Parse("2006-01-02 15:04:05", m[0][1])
					if err != nil {
						return nil, fmt.Errorf("error parsing LastUpdated %s: %v", m[0][1], err)
					}
				}
				// Parse talk groups
				m = talkGroupRegex.FindAllStringSubmatch(r.Rfinder, -1)
				for _, s := range m {
					id, err := strconv.Atoi(s[2])
					if err != nil {
						return nil, fmt.Errorf("error parsing TalkGroup ID %s: %v", s[2], err)
					}
					ts, err := strconv.Atoi(s[1])
					if err != nil {
						return nil, fmt.Errorf("error parsing TalkGroup TimeSlot %s: %v", s[1], err)
					}
					r.TalkGroups = append(r.TalkGroups, TalkGroup{
						Number:   id,
						TimeSlot: ts,
						Name:     s[4],
					})
				}
				newResults = append(newResults, r)
			}
		}
		result.Count = len(newResults)
		result.Results = newResults
	}
	// pretty.Println(result)
	return &result, nil
}

func Matches(filter filter, rv reflect.Value) bool {
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
	Callsign    string `json:"callsign"`     // "KC1FRJ"
	City        string `json:"city"`         // "Presque Isle"
	ColorCode   int    `json:"color_code"`   // 12
	Country     string `json:"country"`      // "United States"
	Details     string `json:"details"`      // "Time Slot #1 - Group Call 759 = SKYWARN<br>Time Slot #1 - Group Call 9998 = Parrot*<br>Time Slot #1 - Group Call 1 = World Wide*<br>Time Slot #1 - Group Call 13 = WW English*<br>Time Slot #1 - Group Call 3 = North America<br>Time Slot #1 - Group Call 3172 = Northeast<br>Time Slot #1 - Group Call 1 = World Wide*<br>Time Slot #1 - Group Call 13 = WW English*<br>Time Slot #1 - Group Call 3 = North America<br>Time Slot #1 - Group Call 3172 = Northeast<br>Time Slot #1 - Group Call 310 = TAC310*<br>Time Slot #1 - Group Call 311 = TAC311*<br>Time Slot #1 - Group Call 113 = UA English 1*<br>Time Slot #1 - Group Call 123 = UA English 2*<br>Time Slot #1 - Group Call 8801 = NETAC 1*<br>------------------------------------<br>Time Slot #2 - Group Call 8802 = NETAC 2*<br>Time Slot #2 - Group Call 3181 = New England Wide<br>Time Slot #2 - Group Call 8 = Region North<br>Time Slot #2 - Group Call 3133 = New Hampshire*<br>Time Slot #2 - Group Call 3123 = ME Statewide<br>Time Slot #1 - Group Call 3029 = New Brunswick<br>Time Slot #2 - Group Call 9 = Local Site<br><br>* PTT Activated<br><br>You Must Have [ARS] Disabled Within Your Radio<br> <br>Contact: Dave, KQ1L<br>Email: dhawke@gwi.net<br>Website: http://nedecn.org<br>"
	Frequency   string `json:"frequency"`    // "145.18000"
	ID          int    `json:"id"`           // 310198
	IPSCNetwork string `json:"ipsc_network"` // "NEDECN"
	Offset      string `json:"offset"`       // "-0.600"
	Rfinder     string `json:"rfinder"`      // "IPSC Network:NEDECN - Color Code:12 - Assigned:Peer - TS Linked:TS1 TS2 - Operator:KC1FRJ<br>Time Slot # 1 - Group Call 1 = World Wide*<br>Time Slot # 1 - Group Call 3 = North America<br>Time Slot # 1 - Group Call 13 = WW English*<br>Time Slot # 1 - Group Call 113 = UA English 1*<br>Time Slot # 1 - Group Call 123 = UA English 2*<br>Time Slot # 1 - Group Call 310 = TAC310*<br>Time Slot # 1 - Group Call 311 = TAC311*<br>Time Slot # 1 - Group Call 759 = SKYWARN<br>Time Slot # 1 - Group Call 3029 = New Brunswick<br>Time Slot # 1 - Group Call 3172 = Northeast<br>Time Slot # 1 - Group Call 8801 = NETAC 1*<br>Time Slot # 1 - Group Call 9998 = Parrot*<br>Time Slot # 2 - Group Call 8 = Region North<br>Time Slot # 2 - Group Call 9 = Local Site<br>Time Slot # 2 - Group Call 3123 = ME Statewide<br>Time Slot # 2 - Group Call 3133 = New Hampshire*<br>Time Slot # 2 - Group Call 3181 = New England Wide<br>Time Slot # 2 - Group Call 8802 = NETAC 2*<br><br>Last Update: 2024-06-24 21:06:51"
	State       string `json:"state"`        // "Maine"
	Trustee     string `json:"trustee"`      // "KC1FRJ"
	TSLinked    string `json:"ts_linked"`    // "TS1 TS2"
	LastUpdated time.Time
	TalkGroups  []TalkGroup
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
