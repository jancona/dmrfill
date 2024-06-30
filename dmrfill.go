package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type filterFlags []filter

// filter clauses look like 'name=val1[,val2]...'
var filterRegex = regexp.MustCompile(`(\w+)=(\w+(?:,[\w\s]+)*)`)
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
	inFile         string
	outFile        string
	datasource     string
	filters        filterFlags
	zonePattern    string
	glPattern      string
	channelPattern string
	power          string
)

func init() {
	flag.StringVar(&inFile, "in", "", "Input QDMR Codeplug YAML file (default STDIN)")
	flag.StringVar(&outFile, "out", "", "Output QDMR Codeplug YAML file (default STDOUT)")
	flag.StringVar(&datasource, "ds", "", "Repeater data source, either RADIOID or REPEATERBOOK (required)")
	flag.Var(&filters, "f", "Filter clause of the form 'name=val1[,val2]...'")
	flag.StringVar(&zonePattern, "zone", "$state_code $city:6 $callsign", "Pattern for forming zone names")
	flag.StringVar(&glPattern, "gl", "", "Pattern for forming group list names (default zone + ' $time_slot')")
	flag.StringVar(&channelPattern, "ch", "$city:6 $callsign $tg_number $tg_name", "Pattern for forming channel names")
	flag.StringVar(&power, "power", "High", "Channel power setting, one of (Min Low Mid High Max)")
}

const (
	radioID      = "RADIOID"
	repeaterbook = "REPEATERBOOK"
)

func main() {
	var codeplug Codeplug //map[interface{}]interface{}

	yamlReader, yamlWriter := parseArguments()
	defer func() {
		yamlReader.Close()
		yamlWriter.Close()
	}()

	decoder := yaml.NewDecoder(yamlReader)
	err := decoder.Decode(&codeplug)
	if err != nil {
		fatal("Unable to parse YAML input, file: %s: %v\n", inFile, err)
	}
	// pretty.Println(codeplug)
	if datasource == radioID {
		result, err := QueryRadioID(&codeplug, filters)
		if err != nil {
			fatal("error querying RadioID: %v", err)
		}
		for _, repeater := range result.Results {
			rxFreq, err := strconv.ParseFloat(repeater.Frequency, 64)
			if err != nil {
				fmt.Printf("skipping repeater with bad Frequency %s: %v\n", repeater.Frequency, err)
				continue
			}
			offset, err := strconv.ParseFloat(repeater.Offset, 64)
			if err != nil {
				fmt.Printf("skipping repeater with bad Offset %s: %v\n", repeater.Offset, err)
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
			codeplug.Zones = append(codeplug.Zones, zone)
			// create two group lists, one for each timeslot
			tg := TalkGroup{
				TimeSlot: 1,
			}
			gl1 := GroupList{
				ID:   NewID(ToSliceOfIDer(codeplug.GroupLists), "grp"),
				Name: ReplaceArgs(glPattern, repeater, &tg),
			}
			codeplug.GroupLists = append(codeplug.GroupLists, gl1)
			tg.TimeSlot = 2
			gl2 := GroupList{
				ID:   NewID(ToSliceOfIDer(codeplug.GroupLists), "grp"),
				Name: ReplaceArgs(glPattern, repeater, &tg),
			}
			codeplug.GroupLists = append(codeplug.GroupLists, gl2)

			for _, tg := range repeater.TalkGroups {
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
					},
				}
				// add it to the codeplug
				codeplug.Channels = append(codeplug.Channels, ch)
				// and to the zone
				zone.A = append(zone.A, ch.Digital.ID)
			}
		}
	}
	// pretty.Println(codeplug)
	encoder := yaml.NewEncoder(yamlWriter)
	encoder.SetIndent(2)
	err = encoder.Encode(codeplug)
	if err != nil {
		fatal("Error encoding YAML output, file: %s: %v\n", outFile, err)
	}
}

func GetOrCreateContact(tg *TalkGroup, codeplug *Codeplug) *Contact {
	for _, c := range codeplug.Contacts {
		if c.DMR.Number == tg.Number {
			return &c
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
	codeplug.Contacts = append(codeplug.Contacts, c)
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
	fmt.Printf(msg, args...)
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
			fatal("Unable to open input file %s: %v\n", inFile, err)
		}
		yamlReader = yamlFile
	} else {
		yamlReader = os.Stdin
	}

	if outFile != "" {
		yamlFile, err := os.Create(outFile)
		if err != nil {
			fatal("Unable to open output file %s: %v\n", outFile, err)
		}
		yamlWriter = yamlFile
	} else {
		yamlWriter = os.Stdout
	}

	if datasource != radioID && datasource != repeaterbook {
		fatal("ds must be one of RADIOID or REPEATERBOOK\n")
	}

	if zonePattern == "" {
		fatal("zone is required\n")
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
