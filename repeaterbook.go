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

const repeaterBookNA = "https://www.repeaterbook.com/api/export.php"
const repeaterBookROW = "https://www.repeaterbook.com/api/exportROW.php"

const kmPerMile = 1.609344

// Supported query parameters
var repeaterBookQueryParamNames = map[string]struct{}{
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

// Supported proximity query parameters
var repeaterBookProxQueryParamNames = map[string]struct{}{
	"qtype": {}, // Proximity search "prox"
	"lat":   {}, // Proximity search latitude
	"lng":   {}, // Proximity search longitude
	"dist":  {}, // Proximity search distance
	"dunit": {}, // Proximity search distance units (km, ?)
}

// RepeaterBookResult JSON field names
var repeaterBookResultFields = map[string]string{}

// Put the RepeaterBookResult JSON field names in a map
func init() {
	st := reflect.TypeOf(RepeaterBookResult{})
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag != "" {
			repeaterBookResultFields[strings.ToLower(tag)] = field.Name
		} else {
			repeaterBookResultFields[strings.ToLower(field.Name)] = field.Name
		}
	}
}

func QueryRepeaterBook(filters filterFlags) (*RepeaterBookResults, error) {
	var base = repeaterBookNA
	if !naRepeaterBookDB {
		base = repeaterBookROW
	}
	if open {
		filters.Set("use=OPEN")
	}
	if onAir {
		filters.Set("operational_status=On-air")
	}
	if location != "" {
		// Do a proximity search
		gResult, err := QueryGeonames(location)
		if err != nil {
			logError("Error reverse geocoding location '%s': %v", location, err)
			os.Exit(1)
		}
		if gResult.TotalResultsCount < 1 {
			logError("No location found for '%s'", location)
			os.Exit(1)
		}
		if radiusUnits == "miles" {
			radius = radius * kmPerMile
		}
		filters.Set("qtype=prox")
		filters.Set("dunit=km")
		filters.Set(fmt.Sprintf("dist=%f", radius))
		filters.Set(fmt.Sprintf("lat=%s", gResult.Geonames[0].Lat))
		filters.Set(fmt.Sprintf("lng=%s", gResult.Geonames[0].Lng))
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		logError("Error parsing base URL %s: %v", base, err)
		os.Exit(1)
	}

	// Filters to be applied on the results
	var resultFilters filterFlags

	// Query params
	params := url.Values{}
	for _, f := range filters {
		if location == "" {
			_, ok := repeaterBookQueryParamNames[f.key]
			if ok && len(f.value) == 1 {
				// RepeaterBook doesn't OR multiple filter parameters, it just uses the last one
				// for _, v := range f.value {
				params.Add(f.key, f.value[0])
				// }
			} else {
				_, ok := repeaterBookResultFields[f.key]
				if ok {
					resultFilters = append(resultFilters, f)
				}
			}
		} else {
			_, ok := repeaterBookProxQueryParamNames[f.key]
			if ok && len(f.value) == 1 {
				// RepeaterBook doesn't OR multiple filter parameters, it just uses the last one
				// for _, v := range f.value {
				params.Add(f.key, f.value[0])
				// }
			} else {
				_, ok := repeaterBookResultFields[f.key]
				if ok {
					resultFilters = append(resultFilters, f)
				}
			}
		}
	}
	baseURL.RawQuery = params.Encode()
	logVerbose("RepeaterBook URL %s", baseURL.String())

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		logError("Error creating HTTP request %s: %v", baseURL.String(), err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cache-Control", "max-age=3600") // Cache results for an hour
	resp, err := cachingHttpClient.Do(req)
	if err != nil {
		logError("Error executing HTTP request %s: %v", baseURL.String(), err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-From-Cache") == "1" {
		logVerbose("using cached response")
	}
	logVeryVerbose("response status %d", resp.StatusCode)
	logVeryVerbose("response headers: %#v", resp.Header)
	decoder := json.NewDecoder(resp.Body)
	result := RepeaterBookResults{
		Results: []RepeaterBookResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		logError("Error parsing JSON response: %v", err)
		os.Exit(1)
	}
	logVerbose("found %d results", result.Count)
	// Do client filtering
	newResults := []RepeaterBookResult{}
	for _, r := range result.Results {
		f, err := strconv.ParseFloat(r.Frequency, 64)
		if err == nil {
			r.Band = band(f)
		}
		rv := reflect.ValueOf(r)
		matchesAll := true
		for _, filter := range resultFilters {
			if !MatchesRepeaterBook(filter, rv) {
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
	logVerbose("%d results after filtering", result.Count)
	// pretty.Println(result)
	return &result, nil
}

func MatchesRepeaterBook(filter filter, rv reflect.Value) bool {
	val := rv.FieldByName(repeaterBookResultFields[filter.key]).String()
	logVeryVerbose("filter %#v, val: %s", filter, val)
	for _, fv := range filter.value {
		if fv == val {
			return true
		}
	}
	return false
}

type RepeaterBookResults struct {
	Count   int                  `json:"count"`
	Results []RepeaterBookResult `json:"results"`
}
type RepeaterBookResult struct {
	StateID           string `json:"State ID"`
	RptrID            string `json:"Rptr ID"`
	Frequency         string `json:"Frequency"`
	InputFreq         string `json:"Input Freq"`
	PL                string `json:"PL"`
	TSQ               string `json:"TSQ"`
	NearestCity       string `json:"Nearest City"`
	Landmark          string `json:"Landmark"`
	Region            string `json:"Region"`
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

func (r RepeaterBookResult) GetCallsign() string {
	return r.Callsign
}
func (r RepeaterBookResult) GetCity() string {
	return r.NearestCity
}
func (r RepeaterBookResult) GetFrequency() string {
	return r.Frequency
}
func (r RepeaterBookResult) GetState() string {
	return r.State
}
