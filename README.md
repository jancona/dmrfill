# dmrfill - Fill codeplugs with repeater data

When I first got my DMR radio, I found the whole process around codeplus very frustrating. Even if your local area repeater group maintains a codeplug, as soon as you travel you're back to trying to find a codeplug with the area repeaters or at least locate repeater and talkgroup data so you can build your own.

`dmrfill` trys to solve this problem by gathering repeater data and adding it an exiting codeplug. So, you start with a minimal codeplug


## How it works

Start with a `QDMR` codeplug YAML file for your radio as a base. Use `dmrfill` to populate it with repeater data gathered from external sources.

Source YAML comes from `QDMR`/dmrconf. Each invocation of `dmrfill` does a query to an external source of repeater data, creates codeplug entries like channels, zones, callgroups, etc. and merges them into the source YAML. `dmrfill` can be run in a pipeline to assemble a codeplug from a variety of sources, like RepeaterBook, RadioID, or user-supplied files (format TBD). The output YAML can be loaded into `QDMR` or piped into `dmrconf`.



Query data sources for repeater info. Sample queries

* Get all DMR repeaters, for a country

## Usage

`ds` Repeater data source, either RADIOID or REPEATERBOOK


### RepeaterBook API

US, Canada, Mexico URL: https://www.repeaterbook.com/api/export.php?state=Maine

Rest of World URL: https://www.repeaterbook.com/api/exportROW.php?country=Germany


### RadioID API

https://radioid.net/api/dmr/repeater/?state=Maine&ipsc_network=NEDECN

Regex to parse `rfinder` field from response JSON: `Time Slot # (\d) [-=] Group Call (\d+)(\s*[-=]\s*([^"<>]*))?`. This regex finds 17623 group call fields in the current RadioID repeater JSON download. The `rfinder` field also contains last update time, e.g. "Last Update: 2024-06-26 21:45:24"

### String Interpolation of Variables into Arguments

For the `-zone` argument, the value can be a string that interpolates values from the repeater being processed along with a maximum length, in order to enable building unique names in the small number og characters available.

Syntax: `Text ${var:length}`

| Variable   | Description |
| ---------- |-------------|
| state      | US/CA State Name (ex. `New Jersey`) |
| state_code | US/CA State Code (ex. `NJ`) |
| city       | City Name (ex. `Newark`)      |
| callsign   | Callsign (ex. `N1ADJ`) |
| frequency  | Frequency (ex. `147.21`) |
| band       | Band (ex. `2m`) |

Example: `-zone '${state_code} ${city:6} ${callsign}'` might produce the output `ME Brunsw N1ADJ`.
