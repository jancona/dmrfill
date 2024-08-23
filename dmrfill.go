package main

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"gopkg.in/yaml.v3"
)

const userAgent = "dmrfill/0.1 github.com/jancona/dmrfill n1adj@anconafamily.com"

var cachingHttpClient *http.Client

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fatal("error accessing home directory: %v", err)
	}
	cacheDir := homeDir + "/.cache/dmrfill"
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		fatal("error creating cache directory: %v", err)
	}
	cache := diskcache.New(cacheDir)
	t := httpcache.NewTransport(cache)
	t.MarkCachedResponses = true

	cachingHttpClient = t.Client()
}

type filterFlags []filter

// filter clauses look like 'name=val1[,val2]...'
var filterRegex = regexp.MustCompile(`(\w+)=([^,]+(?:,[\w\s]+)*)`)
var valuesRegex = regexp.MustCompile(`,?([^,]+)`)

type filter struct {
	key      string
	rawValue string
	value    []string
}

func (ff *filterFlags) String() string {
	return fmt.Sprintf("%#v", *ff)
}

func (ff *filterFlags) Set(value string) error {
	m := filterRegex.FindAllStringSubmatch(value, -1)
	if len(m) == 0 {
		return errors.New("invalid filter expression '" + value + "'")
	}
	f := filter{
		key:      strings.ReplaceAll(m[0][1], "_", " "),
		rawValue: m[0][2],
	}
	m2 := valuesRegex.FindAllStringSubmatch(f.rawValue, -1)
	for _, v := range m2 {
		f.value = append(f.value, v[1])
	}
	// logVeryVerbose("filter.Set(%s) %#v", value, f)
	*ff = append(*ff, f)
	return nil
}

var (
	inFile             string
	outFile            string
	datasource         string
	filters            filterFlags
	zonePattern        string
	glPattern          string
	channelPattern     string
	power              string
	talkgroupsRequired bool
	dmrQuery           bool
	naRepeaterBookDB   bool
	nameLength         int
	open               bool
	onAir              bool
	location           string
	radius             float64
	radiusUnits        string
	verbose            bool
	veryVerbose        bool
)

func init() {
	flag.StringVar(&inFile, "in", "", "Input QDMR Codeplug YAML file (default STDIN)")
	flag.StringVar(&outFile, "out", "", "Output QDMR Codeplug YAML file (default STDOUT)")
	flag.StringVar(&datasource, "ds", "", "Repeater data source, either RADIOID_DMR or REPEATERBOOK_FM (required)")
	flag.Var(&filters, "f", "Filter clause of the form 'name=val1[,val2]...'")
	flag.StringVar(&zonePattern, "zone", "$state_code $city:6 $callsign", "Pattern for forming DMR zone names, zone name for analog")
	flag.StringVar(&glPattern, "gl", "", "Pattern for forming DMR group list names (default zone + ' $time_slot')")
	flag.StringVar(&channelPattern, "ch", "$tg_name:8 $tg_number $time_slot $callsign $city", "Pattern for forming DMR channel names")
	flag.StringVar(&power, "power", "High", "Channel power setting, one of ('Min' 'Low' 'Mid' 'High' 'Max')")
	flag.BoolVar(&talkgroupsRequired, "tg", true, "Only include DMR repeaters that have talkgroups defined")
	flag.BoolVar(&naRepeaterBookDB, "na", true, "Use North American RepeaterBook database. Set it to 'false' to query outside the US, Canada and Mexico.")
	flag.IntVar(&nameLength, "name_lim", 16, "Length limit for generated names")
	flag.BoolVar(&open, "open", true, "Only include open repeaters")
	flag.BoolVar(&onAir, "on_air", true, "Only include on-air repeaters")
	flag.StringVar(&location, "loc", "", "Center location for proximity search, e.g. 'Bangor, ME', 'München'")
	flag.Float64Var(&radius, "radius", 25, "Radius for proximity search")
	flag.StringVar(&radiusUnits, "units", "miles", "Distance units for proximity search, one of ('miles' 'km')")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&veryVerbose, "vv", false, "more verbose logging")
}

const (
	radioID      = "RADIOID_DMR"
	repeaterBook = "REPEATERBOOK_FM"
	// brandmeister      = "BRANDMEISTER_DMR"
)

func main() {
	var codeplug Codeplug
	for i, a := range os.Args {
		if i == 0 {
			logVerbose("%s", a)
		} else {
			logVerbose("    %s", a)
		}
	}

	yamlReader, yamlWriter := parseArguments()
	defer func() {
		yamlReader.Close()
		yamlWriter.Close()
	}()

	decoder := yaml.NewDecoder(yamlReader)
	err := decoder.Decode(&codeplug)
	if err != nil {
		fatal("Unable to parse YAML input, file: %s: %v", inFile, err)
	}
	// pretty.Println(codeplug)
	switch datasource {
	case radioID:
		rbFilters := make(filterFlags, len(filters)+1)
		copy(rbFilters, filters)
		rbFilters.Set("mode=dmr")
		repeaterList, err := QueryRepeaterBook(filters)
		if err != nil {
			fatal("error querying RepeaterBook: %v", err)
		}
		var b strings.Builder
		b.WriteString("id=")
		first := true
		for _, r := range repeaterList.Results {
			if r.DMRID == "" {
				logVerbose("skipping repeater %s %s with empty DMRID", r.Callsign, r.Frequency)
				continue
			} else {
				id, err := strconv.Atoi(r.DMRID)
				if err != nil || id == 0 {
					logVerbose("skipping repeater %s %s with invalid DMRID: %v", r.Callsign, r.Frequency, err)
					continue
				}
			}
			if !first {
				b.WriteString(",")
			}
			b.WriteString(r.DMRID)
			first = false
		}
		ridFilters := make(filterFlags, len(filters)+1)
		copy(rbFilters, filters)
		ridFilters.Set(b.String())

		result, err := QueryRadioID(ridFilters)
		if err != nil {
			fatal("error querying RadioID: %v", err)
		}
		logVeryVerbose("RadioID results %#v", result)

		for _, repeater := range result.Results {
			logVeryVerbose("Processing repeater %#v", repeater)
			rxFreq, err := strconv.ParseFloat(repeater.Frequency, 64)
			if err != nil {
				logError("skipping repeater with bad Frequency %s: %v", repeater.Frequency, err)
				continue
			}
			offset, err := strconv.ParseFloat(repeater.Offset, 64)
			if err != nil {
				logError("skipping repeater with bad Offset %s: %v", repeater.Offset, err)
				continue
			}
			txFreq := rxFreq + offset
			zoneName := ReplaceArgs(zonePattern, repeater, nil)
			// create a Zone
			zone := Zone{
				ID:   NewID(ToSliceOfIDer(codeplug.Zones), "zone"),
				Name: zoneName,
			}
			// add it to the codeplug
			codeplug.Zones = append(codeplug.Zones, &zone)
			// create two group lists, one for each timeslot
			tg := TalkGroup{
				TimeSlot: 1,
			}
			gl1 := GroupList{
				ID:   NewID(ToSliceOfIDer(codeplug.GroupLists), "grp"),
				Name: ReplaceArgs(glPattern, repeater, &tg),
			}
			codeplug.GroupLists = append(codeplug.GroupLists, &gl1)
			tg.TimeSlot = 2
			gl2 := GroupList{
				ID:   NewID(ToSliceOfIDer(codeplug.GroupLists), "grp"),
				Name: ReplaceArgs(glPattern, repeater, &tg),
			}
			codeplug.GroupLists = append(codeplug.GroupLists, &gl2)
			net := strings.ToLower(repeater.IPSCNetwork)
			var bmTalkgroups []TalkGroup
			if strings.Contains(net, "brandmeister") || strings.Contains(net, "bm") {
				// Call Brandmeister API to lookup talkgroups
				bmTalkgroups, err = QueryBrandmeisterTalkgroups(repeater)
				if err != nil {
					logError("error getting Brandmeister talkgroups for repeater %#v: %v", repeater, err)
				}
			}
			logVeryVerbose("radioID talkgroups: %#v", repeater.TalkGroups)
			logVeryVerbose("bmTalkgroups: %#v", bmTalkgroups)
			if len(bmTalkgroups) > len(repeater.TalkGroups) {
				repeater.TalkGroups = bmTalkgroups
			}

			logVeryVerbose("repeater.TalkGroups: %#v", repeater.TalkGroups)
			for _, tg := range repeater.TalkGroups {
				if tg.TimeSlot != 1 && tg.TimeSlot != 2 {
					logError("skipping invalid timeslot: %#v", tg)
					continue
				}
				// logVerbose("%#v", tg)
				// for each repeater-talkgroup combo,
				//   if no contact exists for the talkgroup,
				//		 create it and add it to the proper group list
				ts := "TS" + strconv.Itoa(tg.TimeSlot)
				var glID string
				c := GetOrCreateContact(&tg, &codeplug)
				if tg.TimeSlot == 1 {
					gl1.Contacts = append(gl1.Contacts, c.DMR.ID)
					glID = gl1.ID
				} else {
					gl2.Contacts = append(gl2.Contacts, c.DMR.ID)
					glID = gl2.ID
				}
				// Always use the contact name as the TG name. That way names are
				// consistent and users can control the name that appears by editing
				// the contact name, which is unique for each TG number.
				tg.Name = c.DMR.Name

				//   create a channel for the combo
				channelName := ReplaceArgs(channelPattern, repeater, &tg)

				ch := Channel{
					Digital: Digital{
						ID:          NewID(ToSliceOfIDer(codeplug.Channels), "ch"),
						Name:        channelName,
						RxFrequency: fmt.Sprintf("%f MHz", rxFreq),
						TxFrequency: fmt.Sprintf("%f MHz", txFreq),
						ColorCode:   repeater.ColorCode,
						TimeSlot:    ts,
						GroupList:   glID,
						Power:       DefaultableString{Value: power, HasValue: true},
						Contact:     c.DMR.ID,
						Admit:       "Always",
					},
				}
				// add it to the codeplug
				codeplug.Channels = append(codeplug.Channels, &ch)
				// and to the zone
				zone.A = append(zone.A, ch.Digital.ID)
			}
		}

	case repeaterBook:
		filters.Set("mode=analog")
		result, err := QueryRepeaterBook(filters)
		if err != nil {
			fatal("error querying RepeaterBook: %v", err)
		}
		// create a Zone for the repeaters queried
		zone := Zone{
			ID:   NewID(ToSliceOfIDer(codeplug.Zones), "zone"),
			Name: zonePattern,
		}
		// add it to the codeplug
		codeplug.Zones = append(codeplug.Zones, &zone)
		for _, repeater := range result.Results {
			rxFreq, err := strconv.ParseFloat(repeater.Frequency, 64)
			if err != nil {
				logError("skipping repeater with bad Frequency %s: %v", repeater.Frequency, err)
				continue
			}
			txFreq, err := strconv.ParseFloat(repeater.InputFreq, 64)
			if err != nil {
				logError("skipping repeater with bad InputFreq %s: %v", repeater.InputFreq, err)
				continue
			}
			var rxTone, txTone Tone
			if repeater.TSQ != "" {
				err = rxTone.Set(repeater.TSQ)
				if err != nil {
					logError("skipping repeater with bad TSQ %s: %v", repeater.TSQ, err)
					continue
				}
			}
			if repeater.PL != "" {
				err = txTone.Set(repeater.PL)
				if err != nil {
					logError("skipping repeater with bad PL %s: %v", repeater.PL, err)
					continue
				}
			}
			//   create a channel
			channelName := ReplaceArgs(channelPattern, repeater, nil)

			ch := Channel{
				Analog: Analog{
					ID:          NewID(ToSliceOfIDer(codeplug.Channels), "ch"),
					Name:        channelName,
					RxFrequency: fmt.Sprintf("%f MHz", rxFreq),
					TxFrequency: fmt.Sprintf("%f MHz", txFreq),
					Admit:       "Always",
					Bandwidth:   "Wide",
					Power:       DefaultableString{Value: power, HasValue: true},
					RxTone:      rxTone,
					TxTone:      txTone,
				},
			}
			// add it to the codeplug
			codeplug.Channels = append(codeplug.Channels, &ch)
			// and to the zone
			zone.A = append(zone.A, ch.Analog.ID)
		}
	}
	for _, z := range codeplug.Zones {
		slices.SortStableFunc(z.A, func(a, b string) int {
			return cmp.Compare(getChannelName(a, codeplug), getChannelName(b, codeplug))
		})

	}
	slices.SortStableFunc(codeplug.Zones, func(a, b *Zone) int {
		return cmp.Compare(a.Name, b.Name)
	})
	// pretty.Println(codeplug)
	encoder := yaml.NewEncoder(yamlWriter)
	encoder.SetIndent(2)
	err = encoder.Encode(codeplug)
	if err != nil {
		fatal("Error encoding YAML output, file: %s: %v", outFile, err)
	}
}
func getChannelName(id string, codeplug Codeplug) string {
	for _, ch := range codeplug.Channels {
		if ch.Analog.ID == id {
			return ch.Analog.Name
		} else if ch.Digital.ID == id {
			return ch.Digital.Name
		}
	}
	return ""
}

func GetOrCreateContact(tg *TalkGroup, codeplug *Codeplug) *Contact {
	for _, c := range codeplug.Contacts {
		if c.DMR.ID != "" && c.DMR.Number == tg.Number {
			return c
		}
	}
	c := Contact{
		DMR: DMR{
			ID:     NewID(ToSliceOfIDer(codeplug.Contacts), "cont"),
			Name:   tg.Name,
			Number: tg.Number,
			Type:   "GroupCall",
		},
	}
	codeplug.Contacts = append(codeplug.Contacts, &c)
	return &c
}

var idRegex = regexp.MustCompile(`([a-zA-Z]+)(\d+)`)

func NewID(ids []IDer, defaultPrefix string) string {
	if len(ids) == 0 {
		return defaultPrefix + "1"
	}
	var prefix string
	var lastNumber int
	for _, id := range ids {
		m := idRegex.FindAllStringSubmatch(id.GetID(), -1)
		if len(m) > 0 {
			prefix = m[0][1]
			// Perhaps we should detect if prefix changes?
			n, err := strconv.Atoi(m[0][2])
			if err == nil && n > lastNumber {
				lastNumber = n
			}

		}
	}
	return fmt.Sprint(prefix, lastNumber+1)
}

func parseArguments() (io.ReadCloser, io.WriteCloser) {
	flag.Parse()
	var (
		yamlReader io.ReadCloser
		yamlWriter io.WriteCloser
	)

	if inFile != "" {
		yamlFile, err := os.Open(inFile)
		if err != nil {
			fatal("Unable to open input file %s: %v", inFile, err)
		}
		yamlReader = yamlFile
	} else {
		yamlReader = os.Stdin
	}

	if outFile != "" {
		yamlFile, err := os.Create(outFile)
		if err != nil {
			fatal("Unable to open output file %s: %v", outFile, err)
		}
		yamlWriter = yamlFile
	} else {
		yamlWriter = os.Stdout
	}

	switch datasource {
	case radioID:
		dmrQuery = true
	case repeaterBook:
		dmrQuery = false
	default:
		fatal("ds must be one of RADIOID_DMR or REPEATERBOOK_FM")
	}

	if !dmrQuery {
		if zonePattern == flag.Lookup("zone").DefValue {
			fatal("zone is required for analog datasources")
		}
		if channelPattern == flag.Lookup("ch").DefValue {
			channelPattern = "$callsign $city"
		}
	}

	if glPattern == "" {
		glPattern = zonePattern + " $time_slot"
	}

	switch power {
	case "Min", "Low", "Mid", "High", "Max":
		// good
	default:
		fatal("power must be one of (Min Low Mid High Max)")
	}

	if radius <= 0.0 {
		fatal("radius must be greater than zero")
	}

	switch radiusUnits {
	case "miles", "km":
		// good
	default:
		fatal("units must be one of (miles km)")
	}

	return yamlReader, yamlWriter
}

type IDer interface {
	GetID() string
}

func ToSliceOfIDer[T IDer](s []T) []IDer {
	result := make([]IDer, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

func fatal(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
	os.Exit(1)
}

func logError(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

func logInfo(f string, args ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

func logVerbose(f string, args ...any) {
	if verbose || veryVerbose {
		fmt.Fprintf(os.Stderr, f+"\n", args...)
	}
}

func logVeryVerbose(f string, args ...any) {
	if veryVerbose {
		fmt.Fprintf(os.Stderr, f+"\n", args...)
	}
}
