package main

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// API Doc: https://www.geonames.org/export/ws-overview.html
// Example: http://api.geonames.org/searchJSON?q=harfords%20point,%20me&maxRows=10&username=
const geonamesURL = "http://api.geonames.org/searchJSON"

func QueryGeonames(query string) (*GeonamesResults, error) {
	var base = geonamesURL

	baseURL, err := url.Parse(base)
	if err != nil {
		panic("Error parsing base URL " + base + ": " + err.Error())
	}
	params := url.Values{}
	params.Add("q", query)
	params.Add("maxRows", "1")
	params.Add("username", "dmrfill")
	baseURL.RawQuery = params.Encode()
	logVerbose("Geonames URL %s", baseURL.String())
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
	result := GeonamesResults{
		Geonames: []GeonamesResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	logVerbose("found %d results", result.TotalResultsCount)
	// pretty.Println(result)
	return &result, nil
}

type GeonamesResults struct {
	TotalResultsCount int
	Geonames          []GeonamesResult
}

type GeonamesResult struct {
	GeonameId   int    `json:"geonameId"`   // 4966529
	Lat         string `json:"lat"`         // "45.49477"
	Lng         string `json:"lng"`         // "-69.6145"
	Name        string `json:"name"`        // "Harfords Point"
	ToponymName string `json:"toponymName"` // "Harfords Point"
	AdminName1  string `json:"adminName1"`  // "Maine"
	AdminCode1  string `json:"adminCode1"`  // "ME"
	CountryId   string `json:"countryId"`   // "6252001"
	CountryCode string `json:"countryCode"` // "US"
	CountryName string `json:"countryName"` // "United States"
	Fcl         string `json:"fcl"`         // "T"
	FclName     string `json:"fclName"`     // "mountain,hill,rock,... "
	Population  int    `json:"population"`  // 0
	FcodeName   string `json:"fcodeName"`   // "cape"
	Fcode       string `json:"fcode"`       // "CAPE"
}
