package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	if rowRepeaterBookDB {
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
	if repeaterBookAPIKey != "" {
		req.Header.Set("X-RB-App-Token", repeaterBookAPIKey)
	}
	logVeryVerbose("RepeaterBook request headers: %s", redactKey(fmt.Sprintf("%v", req.Header)))
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
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		msg := fmt.Sprintf("RepeaterBook request failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			msg += fmt.Sprintf(" (retry after %s seconds)", ra)
		}
		logError("%s", msg)
		os.Exit(1)
	}
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
		f, err := strconv.ParseFloat(string(r.Frequency), 64)
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

// flexString is a string that also accepts JSON numbers, booleans and null.
// RepeaterBook changes field datatypes between string and number without
// notice (see issues #3 and #5), so every result field uses this type.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	switch {
	case string(b) == "null":
		*f = ""
	case len(b) > 0 && b[0] == '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexString(s)
	case len(b) > 0 && (b[0] == '{' || b[0] == '['):
		return fmt.Errorf("unsupported JSON value %s", b)
	default:
		// number or boolean: keep the literal text as-is
		*f = flexString(b)
	}
	return nil
}

type RepeaterBookResults struct {
	Count   int                  `json:"count"`
	Results []RepeaterBookResult `json:"results"`
}
type RepeaterBookResult struct {
	StateID           flexString `json:"State ID"`
	RptrID            flexString `json:"Rptr ID"`
	Frequency         flexString `json:"Frequency"`
	InputFreq         flexString `json:"Input Freq"`
	PL                flexString `json:"PL"`
	TSQ               flexString `json:"TSQ"`
	NearestCity       flexString `json:"Nearest City"`
	Landmark          flexString `json:"Landmark"`
	Region            flexString `json:"Region"`
	County            flexString `json:"County"`
	State             flexString `json:"State"`
	Country           flexString `json:"Country"`
	Lat               flexString `json:"Lat"`
	Long              flexString `json:"Long"`
	Precise           flexString `json:"Precise"`
	Callsign          flexString `json:"Callsign"`
	Use               flexString `json:"Use"`
	OperationalStatus flexString `json:"Operational Status"`
	Ares              flexString `json:"ARES"`
	Races             flexString `json:"RACES"`
	Skywarn           flexString `json:"SKYWARN"`
	Canwarn           flexString `json:"CANWARN"`
	AllStarNode       flexString `json:"AllStar Node"`
	EchoLinkNode      flexString `json:"EchoLink Node"`
	IRLPNode          flexString `json:"IRLP Node"`
	WiresNode         flexString `json:"Wires Node"`
	FMAnalog          flexString `json:"FM Analog"`
	Dmr               flexString `json:"DMR"`
	DMRColorCode      flexString `json:"DMR Color Code"`
	DMRID             flexString `json:"DMR ID"`
	DStar             flexString `json:"D-Star"`
	Nxdn              flexString `json:"NXDN"`
	APCOP25           flexString `json:"APCO P-25"`
	P25NAC            flexString `json:"P-25 NAC"`
	M17               flexString `json:"M17"`
	M17CAN            flexString `json:"M17 CAN"`
	Tetra             flexString `json:"Tetra"`
	TetraMCC          flexString `json:"Tetra MCC"`
	TetraMNC          flexString `json:"Tetra MNC"`
	SystemFusion      flexString `json:"System Fusion"`
	YSFDGIDUplink     flexString `json:"YSF DG ID Uplink"`
	YSFDGISDownlink   flexString `json:"YSF DG IS Downlink"`
	YSFDSC            flexString `json:"YSF DSC"`
	Notes             flexString `json:"Notes"`
	LastUpdate        flexString `json:"Last Update"`
	Band              string
}

func (r RepeaterBookResult) GetCallsign() string {
	return string(r.Callsign)
}
func (r RepeaterBookResult) GetCity() string {
	return string(r.NearestCity)
}
func (r RepeaterBookResult) GetFrequency() string {
	return string(r.Frequency)
}
func (r RepeaterBookResult) GetState() string {
	return string(r.State)
}
