package main

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"

	"gopkg.in/yaml.v3"
)

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
		key:      m[0][1],
		rawValue: m[0][2],
	}
	m2 := valuesRegex.FindAllStringSubmatch(f.rawValue, -1)
	for _, v := range m2 {
		f.value = append(f.value, v[1])
	}
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
	naRepeaterbookDB   bool
)

func init() {
	flag.StringVar(&inFile, "in", "", "Input QDMR Codeplug YAML file (default STDIN)")
	flag.StringVar(&outFile, "out", "", "Output QDMR Codeplug YAML file (default STDOUT)")
	flag.StringVar(&datasource, "ds", "", "Repeater data source, either RADIOID or REPEATERBOOK (required)")
	flag.Var(&filters, "f", "Filter clause of the form 'name=val1[,val2]...'")
	flag.StringVar(&zonePattern, "zone", "$state_code $city:6 $callsign", "Pattern for forming DMR zone names, zone name for analog")
	flag.StringVar(&glPattern, "gl", "", "Pattern for forming DMR group list names (default zone + ' $time_slot')")
	flag.StringVar(&channelPattern, "ch", "$tg_name $tg_number $callsign $city", "Pattern for forming DMR channel names")
	flag.StringVar(&power, "power", "High", "Channel power setting, one of (Min Low Mid High Max)")
	flag.BoolVar(&talkgroupsRequired, "tg", true, "Only include DMR repeaters that have talkgroups defined")
	flag.BoolVar(&naRepeaterbookDB, "na", true, "Use North American Repeaterbook database. Set it to 'false' to query outside the US, Canada and Mexico.")
}

const (
	radioID      = "RADIOID"
	repeaterbook = "REPEATERBOOK"
)

func main() {
	var codeplug Codeplug

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
		result, err := QueryRadioID(&codeplug, filters)
		if err != nil {
			fatal("error querying RadioID: %v", err)
		}
		for _, repeater := range result.Results {
			rxFreq, err := strconv.ParseFloat(repeater.Frequency, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping repeater with bad Frequency %s: %v\n", repeater.Frequency, err)
				continue
			}
			offset, err := strconv.ParseFloat(repeater.Offset, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping repeater with bad Offset %s: %v\n", repeater.Offset, err)
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

			for _, tg := range repeater.TalkGroups {
				if tg.TimeSlot != 1 && tg.TimeSlot != 2 {
					fmt.Fprintf(os.Stderr, "skipping invalid timeslot: %#v\n", tg)
					continue
				}
				// fmt.Fprintf(os.Stderr, "%#v\n", tg)
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

				//   create a channel for the combo
				channelName := ReplaceArgs(channelPattern, repeater, &tg)

				ch := Channel{
					Digital: Digital{
						ID:          NewID(ToSliceOfIDer(codeplug.Channels), "ch"),
						Name:        channelName,
						RxFrequency: rxFreq,
						TxFrequency: txFreq,
						ColorCode:   repeater.ColorCode,
						TimeSlot:    ts,
						RadioID: DefaultableInt{
							Default: true,
						},
						GroupList: glID,
						Power:     DefaultableString{Value: power},
						Contact:   c.DMR.ID,
						Admit:     "Always",
					},
				}
				// add it to the codeplug
				codeplug.Channels = append(codeplug.Channels, &ch)
				// and to the zone
				zone.A = append(zone.A, ch.Digital.ID)
			}
		}

	case repeaterbook:
		result, err := QueryRepeaterbook(&codeplug, filters)
		if err != nil {
			fatal("error querying Repeaterbook: %v", err)
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
				fmt.Fprintf(os.Stderr, "skipping repeater with bad Frequency %s: %v\n", repeater.Frequency, err)
				continue
			}
			txFreq, err := strconv.ParseFloat(repeater.InputFreq, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping repeater with bad InputFreq %s: %v\n", repeater.InputFreq, err)
				continue
			}
			var rxTone, txTone Tone
			if repeater.PL != "" {
				rxTone.CTCSS, err = strconv.ParseFloat(repeater.PL, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skipping repeater with bad PL %s: %v\n", repeater.PL, err)
					continue
				}
			}
			if repeater.TSQ != "" {
				txTone.CTCSS, err = strconv.ParseFloat(repeater.TSQ, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "skipping repeater with bad TSQ %s: %v\n", repeater.TSQ, err)
					continue
				}
			}
			//   create a channel
			channelName := ReplaceArgs(channelPattern, repeater, nil)

			ch := Channel{
				Analog: Analog{
					ID:          NewID(ToSliceOfIDer(codeplug.Channels), "ch"),
					Name:        channelName,
					RxFrequency: rxFreq,
					TxFrequency: txFreq,
					Admit:       "Always",
					Bandwidth:   "Wide",
					Power:       DefaultableString{Value: power},
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

func fatal(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
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
	case repeaterbook:
		dmrQuery = false
	default:
		fatal("ds must be one of RADIOID or REPEATERBOOK")
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
