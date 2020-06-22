![servicemeow logo](servicemeow_logo.png "servicemeow")

servicemeow is an unoffical ServiceNow CLI powered by the underlying ServiceNow REST API. It uses [cobra](https://github.com/spf13/cobra) to wrap REST calls in easy to use commands.

## :smiley_cat: Overview
servicemeow tries to operate on a `VERB` `NOUN`  `--ADJECTIVE` pattern to form simple to understand commands, for example

```bash
servicemeow get change CHG0000001 --output report
```

should clearly get the change with change number CHG0000001 and output the result as a report

servicemeow ships as a linux binary or a Docker container available from Github packages.

### :tada: Getting started

1. Grab the latest binary release at https://github.com/cosmosdevops/servicemeow/releases and put it on your $PATH
2. Create a .servicemeow.yaml config file  
  
```yaml
servicenow:
  username: <ServiceNow account username>
  password: <ServiceNow account password>
  url: "https://<ServiceNow URL>/api"
```

3. Explore the options!

```
./servicemeow --help
servicemeow is a cli for simplifying interacting with ServiceNow.
It handles both the creation, updating and processing of ServiceNow records
with configuration options suitable for automation. meow.

Usage:
  servicemeow [command]

Available Commands:
  add         Add new records to ServiceNow
  approve     Approve existing records in ServiceNow
  cancel      Cancel the workflow of a ServiceNow record
  close       Close a ServiceNow record
  edit        Edit a ServiceNow record
  get         Get a ServiceNow record
  help        Help about any command
  implement   Move a ServiceNow request into the Implement state
  reject      Reject a ServiceNow record
  schedule    Schedule a ServiceNow record

Flags:
      --config string   config file (default is $HOME/.servicemeow.yaml)
  -h, --help            help for servicemeow
      --nocolor         disable color output

Use "servicemeow [command] --help" for more information about a command.

```

&nbsp;

### :scroll: `Add`ing/`Edit`ing records
Commands which take input files to `add` or `edit` records expect a payload in YAML format.

The keys for fields are dependant on your ServiceNow instance and can be defined in `camelCase`, `snake_case`, `kebab-case` or `Space Seperated`

```yaml
justification: because I want to!
short_Description: "Doing stuff via the UI is for losers"  
Assignment Group: Help Desk
```
&nbsp;

### :watch: Got nothin' but Time
Commands which take a date/time input can be given in plain English (included relatively!) or in `YYYY-MM-DD HH:MM:SS` 

```bash
servicemeow implement change CHG0030334 --start "Tomorrow 4pm" --end "Friday 6pm" 
```

> :exclamation: **WARNING**  Language which is not understood is ignored. This can have unintented consequences with typos as `--start "22nd Decmber` would resolve to "22nd" of the current month (as "Decmber" would be ignored.)"

&nbsp;

### :flags: Flags, Env vars, file configuration

All flags and configuration options can be defined  

* on the command line `--<flagname>` or its short code `-f` (for example`--required <field>`)
* in Environment Variables with a `SM` prefix (for example `SM_REQUIRED=<field>`)
* in the `.servicemeow.yaml` configuration file (`required: <field>`)

in this order of precedence.

&nbsp;

## Warranty

servicemeow is provided on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied, including, without limitation, any warranties or conditions
of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
PARTICULAR PURPOSE.

Especially now, servicemeow is still in *BETA*; use at your own risk.
