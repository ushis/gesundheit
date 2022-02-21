# Gesundheit

Get notifications about unexpected system state from your local Gesundheitsdienst.

## Install

#### Arch Linux & Debian

Arch Linux & Debian users can install gesundheit via the
[gesundheit package repository](https://ushis.github.io/gesundheit/).

#### Linux

Other Linux users can download binaries from the
[releases page](https://github.com/ushis/gesundheit/releases).

#### From Source

You will need a [Go compiler](https://go.dev/), if you want to install gesundheit
from source. Once you have got one, it is as easy as:

```
go install github.com/ushis/gesundheit@latest
```

## Usage

Create a configuration file (e.g. `/etc/gesundheit/gesundheit.toml`).

```toml
# /etc/gesundheit/gesundheit.toml
#
# We are going to log everything to stdout without timestamps. systemd is
# taking care of the rest.
[Log]
Path = "-"
Timestamps = false

# The configuration files for our modules live in
# /etc/gesundheit/modules.d/*.toml
[Modules]
Config = "modules.d/*.toml"
```

Create some check and handler configuration files.

```toml
# /etc/gesundheit/modules.d/check-backup.toml
#
# Our backup system touches /var/lib/backup/backup.stamp after every
# successful backup run. Lets check once every hour that the stamp has
# been touched within the last 25 hours.
[Check]
Module = "mtime"
Description = "Daily Backup"
Interval = "1h"

[Check.Config]
Path = "/var/lib/backup/backup.stamp"
MaxAge = "25h"
```

```toml
# /etc/gesundheit/modules.d/log.toml
#
# We don't trust gesundheit yet. For this reason we are going to use no
# filter and log every single check result.
[Handler]
Module = "log"
```

```toml
# /etc/gesundheit/modules.d/gotify.toml
#
# We need to get notified in case something is off and want to use gotify for
# that. Since people do not want to get spammed with every single check result
# we will filter for result changes (OK -> FAIL and FAIL -> OK). It is also
# important to not disturb them outside of working hours or while they are
# having lunch.
#
# Since this file contains a secret, it is important to set appropriate
# permissions.
#
#   chown root:gesundheit /etc/gesundheit/modules.d/*.toml
#   chmod 0640 /etc/gesundheit/modules.d/*.toml
#
[Handler]
Module = "gotify"

[Handler.Config]
Url = "https://gotify.example.com/"
Token = "secret"

[[Handler.Filter]]
Module = "result-change"

[[Handler.Filter]]
Module = "office-hours"

[Handler.Filter.Config]
Hours = [
  {From = "9:00", To = "13:00"},
  {From = "14:00", To = "17:00"},
]
```

We are ready to go.

```
gesundheit -conf /etc/gesundheit/gesundheit.toml
```

### Checks

<table>
  <thead>
    <tr>
      <th>Module</th>
      <th>Description</th>
      <th>Config</th>
      <th>Config Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="2"><strong>disk-space</strong></td>
      <td rowspan="2">Check available disk space</td>
      <td>MountPoint</td>
      <td>Mount point of the disk to check, e.g. <code>"/"</code></td>
    </tr>
    <tr>
      <td>MinAvailable</td>
      <td>Min available space considered healthy, e.g. <code>"1G"</code></td>
    </tr>
    <tr>
      <td rowspan="5"><strong>http-json</strong></td>
      <td rowspan="5">Check json value in http response</td>
      <td>Method</td>
      <td>HTTP request method, e.g. <code>"GET"</code></td>
    </tr>
    <tr>
      <td>Url</td>
      <td>Url used to request the json document</td>
    </tr>
    <tr>
      <td>Header</td>
      <td>
        HTTP request header, e.g.<br/>
        <code>{Authorization = ["Token secret"]}</code>
      </td>
    </tr>
    <tr>
      <td>Query</td>
      <td>
        <a href="https://stedolan.github.io/jq/">jq</a> compatible query string,
        e.g. <code>".status"</code>
      </td>
    </tr>
    <tr>
      <td>Value</td>
      <td>Expected value</td>
    </tr>
    <tr>
      <td rowspan="4"><strong>http-status</strong></td>
      <td rowspan="4">Check status of http response</td>
      <td>Method</td>
      <td>HTTP request method, e.g. <code>"GET"</code></td>
    </tr>
    <tr>
      <td>Url</td>
      <td>Url to request</td>
    </tr>
    <tr>
      <td>Header</td>
      <td>
        HTTP request header, e.g.<br/>
        <code>{Authorization = ["Token secret"]}</code>
      </td>
    </tr>
    <tr>
      <td>Status</td>
      <td>Status code considered healthy, e.g. <code>200</code></td>
    </tr>
    <tr>
      <td><strong>memory</strong></td>
      <td>Check available memory</td>
      <td>MinAvailable</td>
      <td>Min available memory considered healthy, e.g. <code>"1G"</code></td>
    </tr>
    <tr>
      <td rowspan="2"><strong>mtime</strong></td>
      <td rowspan="2">Check mtime of a file</td>
      <td>Path</td>
      <td>Path to file</td>
    </tr>
    <tr>
      <td>MaxAge</td>
      <td>
        Max age of the file considered healthy,
        e.g. <code>"1h5m10s"</code>
      </td>
    </tr>
  </tbody>
</table>

### Handlers

<table>
  <thead>
    <tr>
      <th>Module</th>
      <th>Description</th>
      <th>Config</th>
      <th>Config Description</th>
      <th>Default</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>log</strong></td>
      <td>Log check results</td>
      <td></td>
      <td></td>
      <td></td>
    </tr>
    <tr>
      <td rowspan="3"><strong>gotify</strong></td>
      <td rowspan="3">Send check results to gotify</td>
      <td>Url</td>
      <td>Url of the gotify server</td>
      <td></td>
    </tr>
    <tr>
      <td>Token</td>
      <td>Gotify application token</td>
      <td></td>
    </tr>
    <tr>
      <td>Priority</td>
      <td>Priority of every gotify message</td>
      <td><code>4</code></td>
    </tr>
  </tbody>
</table>

### Filters

<table>
  <thead>
    <tr>
      <th>Module</th>
      <th>Description</th>
      <th>Config</th>
      <th>Config Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><strong>result-change</strong></td>
      <td>Filter changed check results</td>
      <td></td>
      <td></td>
    </tr>
    <tr>
      <td><strong>office-hours</strong></td>
      <td>Filter events inside given time spans</td>
      <td>Hours</td>
      <td>
        List of time spans, e.g.<br/>
        <code>[{From = "9:00", To = "17:00"}]</code>
      </td>
    </tr>
  </tbody>
</table>
