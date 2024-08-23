package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// API Doc: https://api.brandmeister.network/docs/#/Device/device.getTalkgroups
// Example: https://api.brandmeister.network/v2/device/313327/talkgroup
const brandmeisterGetTalkgroupsURL = "https://api.brandmeister.network/v2/device/%d/talkgroup"
const brandmeisterGetTalkGroupNamessURL = "https://api.brandmeister.network/v2/talkgroup"

var talkgroupNames map[string]string

func getTalkgroupNames() map[string]string {
	tgNames := make(map[string]string)
	logVeryVerbose("getTalkgroupNames()")
	req, err := http.NewRequest("GET", brandmeisterGetTalkGroupNamessURL, nil)
	if err != nil {
		logError("error creating talkgroup names request: %#v", err)
		return nil
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cache-Control", "max-age=3600") // Cache results for an hour
	resp, err := cachingHttpClient.Do(req)
	if err != nil {
		logError("error getting talkgroup names: %#v", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-From-Cache") == "1" {
		logVerbose("using cached response")
	}
	logVeryVerbose("response status %s", resp.Status)
	logVeryVerbose("response headers: %#v", resp.Header)
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tgNames)
	if err != nil {
		logError("error decoding talkgroup names: %#v", err)
		return nil
	}
	logVeryVerbose("rgetTalkgroupNames results: %#v", tgNames)
	return tgNames
}

func QueryBrandmeisterTalkgroups(repeater RadioIDResult) ([]TalkGroup, error) {
	logVeryVerbose("QueryBrandmeisterTalkgroups")
	if talkgroupNames == nil {
		talkgroupNames = getTalkgroupNames()
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(brandmeisterGetTalkgroupsURL, repeater.ID), nil)
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
	var bmResult []BrandmeisterResult
	err = decoder.Decode(&bmResult)
	if err != nil {
		return nil, err
	}
	var results []TalkGroup
	for _, r := range bmResult {
		tg, err := strconv.Atoi(r.TalkGroup)
		if err != nil {
			return nil, fmt.Errorf("error parsing TalkGroup ID %s: %v", r.TalkGroup, err)
		}
		ts, err := strconv.Atoi(r.Slot)
		if err != nil {
			return nil, fmt.Errorf("error parsing TalkGroup TimeSlot %s: %v", r.Slot, err)
		}
		var name string
		if tg == repeater.ID {
			name = repeater.Callsign + " Local"
		} else {
			name = talkgroupNames[r.TalkGroup]
		}
		results = append(results, TalkGroup{
			Number:   tg,
			TimeSlot: ts,
			Name:     name,
		})
	}
	logVeryVerbose("QueryBrandmeisterTalkgroups results: %#v", results)

	return results, nil
}

type BrandmeisterResult struct {
	TalkGroup  string `json:"talkgroup"`  // "3133"
	Slot       string `json:"slot"`       // "2"
	RepeaterID string `json:"repeaterid"` // "313327"
}
