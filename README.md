# dmrfill - Fill codeplugs with repeater data

When I first got my DMR radio, I found the whole process around codeplus very frustrating. Even if your local area repeater group maintains a codeplug, as soon as you travel you're back to trying to find a codeplug with the area repeaters or at least locate repeater and talkgroup data so you can build your own.

`dmrfill` trys to solve this problem by gathering repeater data and adding it an exiting codeplug. So, you start with a minimal codeplug in [QDMR Extensible Codeplug](https://dm3mat.darc.de/qdmr/manual/ch03.html) format and use `dmrfill` to query [RadioID](https://radioid.net/) or [RepeaterBook](https://www.repeaterbook.com/) for repeater data and then create the proper zones, channels, contacts, and group lists. You can do this multiple times for different areas by piping the results through `dmrfill` multiple times, building up a custom codeplug.

## Installation

`dmrfill` runs on Linux and MacOS, because those are the platforms `QDMR` and `dmrconf` run on. To install it, go to the [latest release page](https://github.com/jancona/dmrfill/releases/latest), and download the executable for your operation system and processor architecture. `darwin` means MacOS, so `dmrfill_darwin_amd64` is the choice for older Intel Macs and `dmrfill_darwin_arm64` is for new Apple silicon (M1, M2, M3) machines. You can rename the downloaded file to `dmrfill`, place it in your PATH, perhaps in `/usr/local/bin` or `~/bin`, and make it executable (`chmod +x dmrfill`).

## Example

```
dmrfill -v -in base.codeplug.yaml -ds RADIOID_DMR -f 'band=2m,70cm' -f 'state=Maine' -f 'county=York,Cumberland,Sagadahoc,Oxford,Androscoggin' -zone 'ME W $city:6 $callsign' | dmrfill -v -ds REPEATERBOOK_FM -zone 'ME W Analog' -f 'state=Maine' -f 'county=York,Cumberland,Sagadahoc,Oxford,Androscoggin' -f 'band=2m,70cm' -out maine.codeplug.yaml
```

This example starts with a base codeplug (`base.codeplug.yaml`) containing basic radio setup (radio ID, callsign, defaults, APRS setup, etc.). It then queries RadioID.net for DMR repeaters in the 2m and 70cm bands in the Maine counties of York, Cumberland, Sagadahoc, Oxford, Androscoggin. It places each repeater that matches the criteria in a zone matching the `ME W $city:6 $callsign` pattern, so the W1IMD repeater in Portland is in the zone _ME W Portla W1IMD_.

The `|` character sends the resulting codeplug to a second run of dmrfill. This one queries Repeaterbook for FM repeaters also in the 2m and 70cm bands in the Maine counties of York, Cumberland, Sagadahoc, Oxford, Androscoggin. It places the repeaters that matches the criteria in a zone named _ME W Analog_ and writes the codeplug in the file `maine.codeplug.yaml`. That file can be edited further with `QDMR` or written to the radio with `QDMR` or `dmrconf`.

## More about how it works

You need a `QDMR` codeplug YAML file for your radio as a base. So start at the [excellent `QDMR` website](https://dm3mat.darc.de/qdmr/) to see if your radio is supported and read about installing and getting started. Once you have `QDMR` running, you can read a working codeplug from your radio and delete the channels, callgroups and zones that you don't want to be in every codeplug you create. That is, leave things that should be in all codeplugs, like simplex channels and zones, but remove the rest. The source codeplug YAML is specified using the `-in` argument. If it's not provided, `dmrfill` expects to read it from `stdin`.

Each invocation of `dmrfill` does a query to an external source of repeater data, then creates codeplug entries like channels, zones, group lists and contacts in a customizable way. It then merges them into the source codeplug and outputs the updated YAML.

### Datasources

Currently `dmrfill` has two datasources:

* `REPEATERBOOK_FM` queries for analog FM repeaters from RepeaterBook.
* `RADIOID_DMR` queries for DMR repeaters, including talkgroups, from RadioID.

A datasource must be specified using the `-ds` argument.

### Filters

Each invocation of `dmrfill` should include one or more filters. A filter takes the form `-f 'field=value1[,valueN...]'`, for example `-f 'state=Maine'` or `-f 'county=York,Cumberland,Sagadahoc,Oxford,Androscoggin'`.

Primary filter fields are:

| Field      | Description |
| ---------- |-------------|
| callsign   | Repeater callsign |
| city       | Repeater city |
| state      | State / Province unabbreviated name (Maine not ME) |
| country    | Repeater country unabbreviated name (United States, not US) |
| county     | Repeater county (US only) |
| band       | Frequency band, one of 10m, 6m, 2m, 1.25m, 70cm, 33cm, 23cm |

Multiple filter values can be provided, separated by commas. A repeater that matches any of the values will be included in the codeplug, so `-f 'band=2m,70cm'` will include repeaters in both the 2m and 70cm bands.

### Naming

The `-zone` and `-gl` arguments allow you to specify a pattern for building the DMR zone or group list names. The value is be a string that interpolates values from the repeater being processed along with a maximum length, in order to enable building unique names in the small number of characters available.

Syntax: `Text $var:length`

| Variable   | Description |
| ---------- |-------------|
| state      | US/CA State Name (ex. `New Jersey`) |
| state_code | US/CA State Code (ex. `NJ`) |
| city       | City Name (ex. `Newark`)      |
| callsign   | Callsign (ex. `N1ADJ`) |
| frequency  | Frequency (ex. `147.21`) |
| band       | Band (ex. `2m`) |

Example: `-zone '$state_code $city:6 $callsign'` might produce the output `ME Brunsw N1ADJ`.

In the case of analog FM zone names, all repeaters go into the specified zone, so there is no need for per-repeater values.

### Pipelines

`dmrfill` can accept input from a file (using the `-in` argument) or from `stdin`. It can output to a file (using the `-out` argument) or to `stdout`. So it can be run in a pipeline to assemble a codeplug from a variety of sources. The first invocation uses `-in` to read from a base file, then the output is piped to additional instances of `dmrfill` to add more repeaters. The final instance uses `-out` to write to an output file which can be loaded to the radio using `QDMR` or `dmrconf`.

## Command Line Options

```
  -ch string
    	Pattern for forming DMR channel names (default "$tg_name:8 $tg_number $time_slot $callsign $city")
  -ds string
    	Repeater data source, either RADIOID_DMR or REPEATERBOOK_FM (required)
  -f value
    	Filter clause of the form 'name=val1[,val2]...'
  -gl string
    	Pattern for forming DMR group list names (default zone + ' $time_slot')
  -in string
    	Input QDMR Codeplug YAML file (default STDIN)
  -na
    	Use North American RepeaterBook database. Set it to 'false' to query outside the US, Canada and Mexico. (default true)
  -name_lim int
    	Length limit for generated names (default 16)
  -on_air
    	Only include on-air repeaters (default true)
  -open
    	Only include open repeaters (default true)
  -out string
    	Output QDMR Codeplug YAML file (default STDOUT)
  -power string
    	Channel power setting, one of (Min Low Mid High Max) (default "High")
  -tg
    	Only include DMR repeaters that have talkgroups defined (default true)
  -v	verbose logging
  -vv
    	more verbose logging
  -zone string
    	Pattern for forming DMR zone names, zone name for analog (default "$state_code $city:6 $callsign")
```

## Acknowledgements

### RepeaterBook
[RepeaterBook](https://repeaterbook.com/) is Amateur Radio's most comprehensive, worldwide, FREE repeater directory. Garrett Dow, KD6KPC runs RepeaterBook as a free service and depends on advertising revenue and donations to keep it going, so please visit the [site](https://repeaterbook.com/) or use their mobile app.

### RadioID
[RadioID](https://radioid.net/) issues DMR and NXDN ID's. It also maintains a DMR repeater database that includes talkgroup information. The [site](https://radioid.net/) also includes a repeater map and contact generator tool.

### QDMR
[QDMR](https://dm3mat.darc.de/qdmr/) is the open source tool that inspired `dmrfill`. Bydefining and documenting a codeplug standard file format, along with tools for reading, writing and editing them, QDMR made `dmrfill` possible.

<!--
### Geonames

-->