# dmrfill - Fill codeplugs with repeater data

When I first got my DMR radio, I found the whole process around codeplus very frustrating. Even if your local area repeater group maintains a codeplug, as soon as you travel you're back to trying to find a codeplug with the area repeaters or at least locate repeater and talkgroup data so you can build your own.

`dmrfill` trys to solve this problem by gathering repeater data and adding it an exiting codeplug. So, you start with a minimal codeplug in [QDMR Extensible Codeplug](https://dm3mat.darc.de/qdmr/manual/ch03.html) format and use `dmrfill` to query [RadioID](https://radioid.net/) or [RepeaterBook](https://www.repeaterbook.com/) for repeater data and then create the proper zones, channels, contacts, and group lists. You can do this multiple times for different areas by piping the results through `dmrfill` multiple times, building up a custom codeplug.

## Example

```
dmrfill -v -in base.codeplug.yaml -ds RADIOID_DMR -f 'band=2m,70cm' -f 'state=Maine' -f 'county=York,Cumberland,Sagadahoc,Oxford,Androscoggin' -zone 'ME W $city:6 $callsign' | dmrfill -v -ds REPEATERBOOK_FM -zone 'ME W Analog' -f 'state=Maine' -f 'county=York,Cumberland,Sagadahoc,Oxford,Androscoggin' -f 'band=2m,70cm' -out maine.codeplug.yaml
```
This example starts with a base codeplug (`base.codeplug.yaml`) containing basic radio setup (radio ID, callsign, defaults, APRS setup, etc.). It then queries RadioID.net for DMR repeaters in the 2m and 70cm bands in the Maine counties of York, Cumberland, Sagadahoc, Oxford, Androscoggin. It places each repeater that matches the criteria in a zone matching the `ME W $city:6 $callsign` pattern, so the W1IMD repeater in Portland is in the zone _ME W Portla W1IMD_.

The `|` character sends the resulting codeplug to a second run of dmrfill. This one queries Repeaterbook for FM repeaters also in the 2m and 70cm bands in the Maine counties of York, Cumberland, Sagadahoc, Oxford, Androscoggin. It places the repeaters that matches the criteria in a zone named _ME W Analog_ and writes the codeplug in the file `maine.codeplug.yaml`. That file can be edited further with `QDMR` or written to the radio with `QDMR` or `dmrconf`.

## How it works

Start with a `QDMR` codeplug YAML file for your radio as a base. Use `dmrfill` to populate it with repeater data gathered from external sources.

The source YAML can be read from the radio using `QDMR` or `dmrconf`. Each invocation of `dmrfill` does a query to an external source of repeater data, creates codeplug entries like channels, zones, callgroups, etc. and merges them into the source YAML. `dmrfill` can be run in a pipeline to assemble a codeplug from a variety of sources. Currently, it supports [RepeaterBook](https://repeaterbook.com/) and [RadioID](https://radioid.net/). The output YAML can be loaded to the radio using `QDMR` or `dmrconf`.

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

### String Interpolation of Variables into Arguments

For the `-zone` and `-gl` arguments, the value can be a string that interpolates values from the repeater being processed along with a maximum length, in order to enable building unique names in the small number og characters available.

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
