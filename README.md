# Gesundheit

Get notifications about unexpected system state from your local Gesundheitsdienst.

## Install

#### Arch Linux & Debian

Arch Linux & Debian users can install gesundheit via the
[gesundheit package repository](http://repo.honkgong.info/gesundheit/).

#### Linux

Other Linux users can download binaries from the
[releases page](https://github.com/ushis/gesundheit/releases).

#### From Source

You will need a [Go compiler](https://go.dev/), [Node.js](https://nodejs.org/)
and the [Yarn Package Manager](https://yarnpkg.com/), if you want to install
gesundheit from source. Once you have got that, it is as easy as:

```
git clone https://github.com/ushis/gesundheit.git
cd gesundheit
yarn install
yarn build
go build
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

# We love pretty dashboards. Therefore we will enable the web interface.
[Http]
Enabled = true
Listen = "127.0.0.1:8080"

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
Module = "file-mtime"
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
      <td rowspan="4"><strong>dns-record</strong></td>
      <td rowspan="4">Check DNS record</td>
      <td>Address</td>
      <td>DNS server address, e.g. <code>"127.0.0.1:53"</code></td>
    </tr>
    <tr>
      <td>Type</td>
      <td>Record type, e.g. <code>"A"</code></td>
    </tr>
    <tr>
      <td>Name</td>
      <td>Name to lookup, e.g. <code>"example.com"</code></td>
    </tr>
    <tr>
      <td>Value</td>
      <td>Expected value, e.g. <code>"1.1.1.1"</code></td>
    </tr>
    <tr>
      <td rowspan="2"><strong>exec</strong></td>
      <td rowspan="2">Execute check command</td>
      <td>Command</td>
      <td>
        Command to execute, e.g.
        <code>"/usr/lib/nagios-plugins/check_load"</code>
      </td>
    </tr>
    <tr>
      <td>Args</td>
      <td>Command arguments, e.g. <code>["-w", "2", "-c", "3"]</code></td>
    </tr>
    <tr>
      <td rowspan="2"><strong>file-mtime</strong></td>
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
    <tr>
      <td rowspan="2"><strong>file-presence</strong></td>
      <td rowspan="2">Check presence of a file</td>
      <td>Path</td>
      <td>Path to file, e.g. <code>"/run/reboot-required"</code></td>
    </tr>
    <tr>
      <td>Present</td>
      <td>
        Whether or not presence of the file is considered healthy, e.g.
        <code>false</code>
      </td>
    </tr>
    <tr>
      <td><strong>heartbeat</strong></td>
      <td>Always <code>OK</code></td>
      <td colspan="2">Useful in combination with a
        <strong>remote</strong> handler and a <strong>node-alive</strong>
        check on the remote node
      </td>
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
        <a href="https://gjson.dev/">GJSON</a> compatible query string,
        e.g. <code>"status"</code>
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
      <td rowspan="2"><strong>node-alive</strong></td>
      <td rowspan="2">Check last appearance of a (remote) node</td>
      <td>Node</td>
      <td>Node to check, e.g. <code>"proxy-01"</code></td>
    </tr>
    <tr>
      <td>MaxAbsenceTime</td>
      <td>
        Max time since last appearance considered healthy,
        e.g. <code>"1m"</code>.
        Configure <strong>heartbeat</strong> with a low interval on the remote node
        if you need timely notifications about absent nodes.
      </td>
    </tr>
    <tr>
      <td rowspan="3"><strong>tls-cert</strong></td>
      <td rowspan="3">Check tls certificate</td>
      <td>Host</td>
      <td>Host to check, e.g. <code>"example.org"</code></td>
    </tr>
    <tr>
      <td>Port</td>
      <td>Port to connect to, e.g. <code>443</code></td>
    </tr>
    <tr>
      <td>MinTTL</td>
      <td>
        Min time until certificate expiry considered healthy,
        e.g. <code>"24h"</code>
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
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="2"><strong>amqp</strong></td>
      <td rowspan="2">
        Send check results to
        <a href="https://www.rabbitmq.com/">RabbitMQ</a>
      </td>
      <td>Url</td>
      <td>
        Url of the RabbitMQ server,
        e.g. <code>"amqp://guest:guest@localhost:5672"</code>
      </td>
    </tr>
    <tr>
      <td>Exchange</td>
      <td>RabbitMQ exchange name, e.g. <code>"gesundheit"</code></td>
    </tr>
    <tr>
      <td rowspan="3"><strong>gotify</strong></td>
      <td rowspan="3">
        Send check results to <a href="https://gotify.net/">Gotify</a>
      </td>
      <td>Url</td>
      <td>
        Url of the Gotify server,
        e.g. <code>"https://gotify.example.org"</code>
      </td>
    </tr>
    <tr>
      <td>Token</td>
      <td>Gotify application token</td>
    </tr>
    <tr>
      <td>Priority</td>
      <td>Priority of every gotify message</td>
    </tr>
    <tr>
      <td><strong>log</strong></td>
      <td>Log check results</td>
      <td></td>
      <td></td>
    </tr>
    <tr>
      <td rowspan="3"><strong>remote</strong></td>
      <td rowspan="3">Send check results to a remote gesundheit node</td>
      <td>Address</td>
      <td>Address of the remote node, e.g. <code>"gesundheit.example.org:9999"</code></td>
    </tr>
    <tr>
      <td>PrivateKey</td>
      <td>
        Private key of the local node, generated
        with <code>gesundheit genkey</code>
      </td>
    </tr>
    <tr>
      <td>PublicKey</td>
      <td>
        Public key of the remote gesundheit node,
        generated with <code>gesundheit pubkey</code>
      </td>
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
      <td>Filter check results inside given time spans</td>
      <td>Hours</td>
      <td>
        List of time spans, e.g.<br/>
        <code>[{From = "9:00", To = "17:00"}]</code>
      </td>
    </tr>
  </tbody>
</table>


### Inputs

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
      <td rowspan="3"><strong>amqp</strong></td>
      <td rowspan="3">
        Receive check results from
        <a href="https://www.rabbitmq.com/">RabbitMQ</a>
      </td>
      <td>Url</td>
      <td>
        Url of the RabbitMQ server,
        e.g. <code>"amqp://guest:guest@localhost:5672"</code>
      </td>
    </tr>
    <tr>
      <td>Exchange</td>
      <td>RabbitMQ exchange name, e.g. <code>"gesundheit"</code></td>
    </tr>
    <tr>
      <td>Queue</td>
      <td>RabbitMQ queue name, e.g. <code>"notifications"</code></td>
    </tr>
    <tr>
      <td rowspan="3"><strong>remote</strong></td>
      <td rowspan="3">Receive check results from a remote gesundheit node</td>
      <td>Listen</td>
      <td>Address to listen on, e.g. <code>"0.0.0.0:9999"</code></td>
    </tr>
    <tr>
      <td>PrivateKey</td>
      <td>
        Private key of the local node, generated
        with <code>gesundheit genkey</code>
      </td>
    </tr>
    <tr>
      <td>Peers</td>
      <td>
        List of peers, e.g. <code>[{ PublicKey = "xxx" }]</code>
      </td>
    </tr>
  </tbody>
</table>

### Databases

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
      <td rowspan="2"><strong>filesystem</strong></td>
      <td rowspan="2">Simple filesystem baked database suitable for most setups.</td>
      <td>Directory</td>
      <td>Database directory, e.g. <code>"/var/lib/gesundheit"</td>
    </tr>
    <tr>
      <td>VacuumInterval</td>
      <td>
        Interval in which expired entries are deleted from disk,
        e.g. <code>"24h"</code>
      </td>
    </tr>
    <tr>
      <td><strong>memory</strong></td>
      <td>
        In memory database suitable for simple setups and nodes
        without persistence requirements
      </td>
      <td></td>
      <td></td>
    </tr>
    <tr>
      <td rowspan="4"><strong>redis</strong></td>
      <td rowspan="4"><a href="https://redis.io/">Redis</a> adapter</td>
      <td>Address</td>
      <td>Address of the redis server, e.g. <code>127.0.0.1:6379</td>
    </tr>
    <tr>
      <td>DB</td>
      <td>Redis database to use, e.g. <code>0</code></td>
    </tr>
    <tr>
      <td>Username</td>
      <td>
        Redis username, e.g. <code>"gesundheit"</code>
      </td>
    </tr>
    <tr>
      <td>Password</td>
      <td>
        Redis password, e.g. <code>"secret"</code>
      </td>
    </tr>
  </tbody>
</table>
