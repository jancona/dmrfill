package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// API Doc: https://www.repeaterbook.com/wiki/doku.php?id=api
// Examples: https://www.repeaterbook.com/api/export.php?state=Maine&county=Cumberland
//           https://www.repeaterbook.com/api/export.php?qtype=prox&dunit=km&lat=44.551&lng=-69.632&dist=40

const repeaterbookNA = "https://www.repeaterbook.com/api/export.php"
const repeaterbookROW = "https://www.repeaterbook.com/api/exportROW.php"

// Supported query parameters
var repeaterbookQueryParamNames = map[string]struct{}{
	"callsign":  {}, // Repeater callsign
	"city":      {}, // Repeater city
	"landmark":  {}, //
	"state":     {}, // State / Province
	"country":   {}, // Repeater country
	"county":    {}, // Repeater county
	"frequency": {}, // Repeater frequency
	"mode":      {}, // Repeater operating mode (analog, DMR, NXDN, P25, tetra)
	"emcomm":    {}, // ARES, RACES, SKYWARN, CANWARN
	"stype":     {}, // Service type. Only required when searching for GMRS repeaters. ex: stype=gmrs
}
var repeaterbookResultFields = map[string]string{}

// Put the RadioIDResult JSON field names in a map
func init() {
	st := reflect.TypeOf(RepeaterbookResult{})
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag != "" {
			repeaterbookResultFields[tag] = field.Name
		} else {
			repeaterbookResultFields[strings.ToLower(field.Name)] = field.Name
		}
	}
}

func QueryRepeaterbook(cp *Codeplug, filters filterFlags) (*RepeaterbookResults, error) {
	var base = repeaterbookNA
	if !naRepeaterbookDB {
		base = repeaterbookROW
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing base URL %s: %v", base, err)
		os.Exit(1)
	}

	// Filters to be applied on the results
	var resultFilters filterFlags

	// Query params
	params := url.Values{}
	for _, f := range filters {
		_, ok := repeaterbookQueryParamNames[f.key]
		if ok {
			for _, v := range f.value {
				params.Add(f.key, v)
			}
		} else {
			_, ok := repeaterbookResultFields[f.key]
			if ok {
				resultFilters = append(resultFilters, f)
			}
		}
	}
	baseURL.RawQuery = params.Encode()
	fmt.Fprintln(os.Stderr, baseURL.String())

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HTTP request %s: %v", baseURL.String(), err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "dmrfill/0.1 github.com/jancona/dmrfill https://www.qrz.com/db/N1ADJ")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing HTTP request %s: %v", baseURL.String(), err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	// fmt.Fprintf(os.Stderr, "response status %d\n", resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	result := RepeaterbookResults{
		Results: []RepeaterbookResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON response: %v", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "Found", result.Count, "results")
	// Do client filtering
	newResults := []RepeaterbookResult{}
	for _, r := range result.Results {
		f, err := strconv.ParseFloat(r.Frequency, 64)
		if err == nil {
			r.Band = band(f)
		}
		rv := reflect.ValueOf(r)
		matchesAll := true
		for _, filter := range resultFilters {
			if !Matches(filter, rv) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			newResults = append(newResults, r)
		}
	}
	result.Count = len(newResults)
	result.Results = newResults
	fmt.Fprintln(os.Stderr, result.Count, "results after filtering")
	// pretty.Println(result)
	return &result, nil
}

type RepeaterbookResults struct {
	Count   int                  `json:"count"`
	Results []RepeaterbookResult `json:"results"`
}
type RepeaterbookResult struct {
	StateID           string `json:"State ID"`
	RptrID            string `json:"Rptr ID"`
	Frequency         string `json:"Frequency"`
	InputFreq         string `json:"Input Freq"`
	PL                string `json:"PL"`
	TSQ               string `json:"TSQ"`
	NearestCity       string `json:"Nearest City"`
	Landmark          string `json:"Landmark"`
	County            string `json:"County"`
	State             string `json:"State"`
	Country           string `json:"Country"`
	Lat               string `json:"Lat"`
	Long              string `json:"Long"`
	Precise           int    `json:"Precise"`
	Callsign          string `json:"Callsign"`
	Use               string `json:"Use"`
	OperationalStatus string `json:"Operational Status"`
	Ares              string `json:"ARES"`
	Races             string `json:"RACES"`
	Skywarn           string `json:"SKYWARN"`
	Canwarn           string `json:"CANWARN"`
	AllStarNode       string `json:"AllStar Node"`
	EchoLinkNode      string `json:"EchoLink Node"`
	IRLPNode          string `json:"IRLP Node"`
	WiresNode         string `json:"Wires Node"`
	FMAnalog          string `json:"FM Analog"`
	Dmr               string `json:"DMR"`
	DMRColorCode      string `json:"DMR Color Code"`
	DMRID             string `json:"DMR ID"`
	DStar             string `json:"D-Star"`
	Nxdn              string `json:"NXDN"`
	APCOP25           string `json:"APCO P-25"`
	P25NAC            string `json:"P-25 NAC"`
	M17               string `json:"M17"`
	M17CAN            string `json:"M17 CAN"`
	Tetra             string `json:"Tetra"`
	TetraMCC          string `json:"Tetra MCC"`
	TetraMNC          string `json:"Tetra MNC"`
	SystemFusion      string `json:"System Fusion"`
	YSFDGIDUplink     string `json:"YSF DG ID Uplink"`
	YSFDGISDownlink   string `json:"YSF DG IS Downlink"`
	YSFDSC            string `json:"YSF DSC"`
	Notes             string `json:"Notes"`
	LastUpdate        string `json:"Last Update"`
	Band              string
}

func (r RepeaterbookResult) GetCallsign() string {
	return r.Callsign
}
func (r RepeaterbookResult) GetCity() string {
	return r.NearestCity
}
func (r RepeaterbookResult) GetFrequency() string {
	return r.Frequency
}
func (r RepeaterbookResult) GetState() string {
	return r.State
}
