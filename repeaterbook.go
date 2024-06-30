package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/kr/pretty"
)

// API Doc: https://www.repeaterbook.com/wiki/doku.php?id=api
// Examples: https://www.repeaterbook.com/api/export.php?state=Maine&county=Cumberland
//           https://www.repeaterbook.com/api/export.php?qtype=prox&dunit=km&lat=44.551&lng=-69.632&dist=40

const repeaterbookNA = "https://www.repeaterbook.com/api/export.php"
const repeaterbookROW = "https://www.repeaterbook.com/api/exportROW.php"

func QueryRepeaterbook() {
	var base = repeaterbookNA
	// TODO: use repeaterbookROW outside North America

	baseURL, err := url.Parse(base)
	if err != nil {
		slog.Error("Error parsing base URL", "URL", base, "error", err)
		os.Exit(1)
	}

	// Query params
	params := url.Values{}
	params.Add("state", "Maine")
	params.Add("county", "Cumberland")
	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		slog.Error("Error creating HTTP request", "baseURL", baseURL.String(), "error", err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "dmrfill/0.1 github.com/jancona/dmrfill https://www.qrz.com/db/N1ADJ")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Error executing HTTP request", "baseURL", baseURL.String(), "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	slog.Info("response status", "status", resp.StatusCode)
	decoder := json.NewDecoder(resp.Body)
	result := RepeaterbookResults{
		Results: []RepeaterbookResult{},
	}
	err = decoder.Decode(&result)
	if err != nil {
		slog.Error("Error parsing JSON response", "error", err)
		os.Exit(1)
	}
	pretty.Println(result)
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
	Pl                string `json:"PL"`
	Tsq               string `json:"TSQ"`
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
}
