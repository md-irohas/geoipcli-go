# GeoIP-CLI

The `geoipcli-go` is a simple command-line interface to look up IP addresses on
[MaxMind](https://www.maxmind.com/)'s downloadable GeoIP databases.

I often look up geolocation data of a number of IP addresses in my research,
but existing tools do not satisfy this need.
So, I wrote `geoipcli-go`.


Main functions:
* Search for the geolocation data of IP addresses on downloadable GeoIP databases.
  * You can pass IP addresses through arguments, stdin, or a file.
* Flexible output columns.
* Support of configuration files.


## Installation

### Pre-compiled binaries

Precompiled binaries for macOS / Linux (x86_64) are released.
See [release](https://github.com/md-irohas/geoipcli-go/releases) page.

These binaries are static linked, so you can use them w/o any additional libraries.
Choose one of the binaries depending on your environment,
put it to a directory in the PATH environment,
and run it.


### Compile from source

`geoipcli-go` is written in Go.
So, if you want to build its binary,
you need to prepare the development environment for Go.

If you are ready to build Go programs, type the following commands.

```sh
$ go get github.com/md-irohas/geoipcli-go
```


## Usage

The followings are the options of `geoipcli-go`.
You can also use configuration files instead of these options (See the following example).

```
Usage of ./geoipcli
  -anonymousip string
    	Path to GeoIP2-AnonymousIP database.
  -asn string
    	Path to GeoLite2-ASN database.
  -city string
    	Path to GeoIP2/GeoLite2-City database.
  -conffile string
    	Config file.
  -contype string
    	Path to GeoIP2-ConnectionType database.
  -country string
    	Path to GeoIP2/GeoLite2-Country database.
  -debug
    	Run this program as debug mode w/ debug message.
  -domain string
    	Path to GeoIP2-Domain database.
  -isp string
    	Path to GeoIP2-ISP database.
  -list-columns
    	Show all column names.
  -output string
    	Output columns separated by comma (,). See '--list-columns' option for more details.
  -readfile string
    	Read IP addresses from file.
  -skip-invalid-ip
    	Skip Invalid IP addresses.
  -version
    	Show version and exit.
exit status 2
```


### Download Databases

MaxMind provides two types of downloadable GeoIP databases.

* [GeoIP2](https://www.maxmind.com/en/geoip2-databases)
  * fee-charging databases.
* [GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/)
  * Free but less accurate databases than GeoIP2.


### Example-1: First Step

This example looks up geolocation data of `1.1.1.1` and `8.8.8.8` using
GeoLite2-ASN and GeoLite2-City.

```
$ geoipcli -asn GeoLite2-ASN.mmdb -city GeoLite2-City.mmdb 1.1.1.1 8.8.8.8

1.1.1.1,AU,,13335,Cloudflare<comma> Inc.
8.8.8.8,US,,15169,Google LLC
```

The results are printed as CSV format.
The columns are IP address, country code, city name, AS number, and AS organization name.

The city columns are empty because they are not found in GeoLite2-City.mmdb.
Plus, the comma (,) is replaced by `<commma>` not to break the CSV format.


### Example-2: Flexible Output

You can set the columns to be printed.
The following command shows the column names available in `geoipcli-go`.

```
$ geoipcli -list-columns

The following columns can be used for output (--output option).

List of columns:
- isp.organization
- isp.autonomous_system_number
- isp.autonomous_system_organization
- isp.isp
- domain.domain
- connection_type.connection_type
- anonymousip.is_anonymous_vpn
- anonymousip.is_hosting_provider
- anonymousip.is_public_proxy
- anonymousip.is_tor_exit_node
- anonymousip.is_anonymous
- country.country.iso_code
- country.traits.is_anonymous_proxy
- country.traits.is_satellite_provider
- country.continent.geoname_id
- country.continent.names.[language]
- country.country.is_in_european_union
- country.country.names.[language]
- country.registered_country.is_in_european_union
- country.continent.code
- country.country.geoname_id
- country.represented_country.is_in_european_union
- country.represented_country.iso_code
- country.represented_country.type
- country.registered_country.iso_code
- country.registered_country.names.[language]
- country.represented_country.names.[language]
- country.registered_country.geoname_id
- country.represented_country.geoname_id
- city.represented_country.iso_code
- city.represented_country.type
- city.subdivisions
- city.city.geoname_id
- city.city.names.[language]
- city.location.latitude
- city.country.iso_code
- city.continent.geoname_id
- city.continent.names.[language]
- city.country.geoname_id
- city.continent.code
- city.registered_country.is_in_european_union
- city.represented_country.geoname_id
- city.traits.is_anonymous_proxy
- city.country.names.[language]
- city.location.longitude
- city.registered_country.iso_code
- city.traits.is_satellite_provider
- city.location.time_zone
- city.postal.code
- city.registered_country.names.[language]
- city.represented_country.is_in_european_union
- city.country.is_in_european_union
- city.location.metro_code
- city.represented_country.names.[language]
- city.location.accuracy_radius
- city.registered_country.geoname_id
- asn.autonomous_system_number
- asn.autonomous_system_organization

Note:
- The [language] strings needs to be replaced by actual languages such as 'en' and 'ja'.

```

If you want to get country names in English and Japanese, country code, and a
flag if the country is in EU from GeoLite2-Country.mmdb, type the following command.

```
$ geoipcli -country GeoLite2-Country.mmdb -output country.country.names.en,country.country.names.ja,country.country.iso_code,country.country.is_in_european_union 1.1.1.1 8.8.8.8

1.1.1.1,Australia,オーストラリア,AU,false
8.8.8.8,United States,アメリカ合衆国,US,false
```

The output option often becomes very long,
so I recommend you use a configuration file (See Example-4).


### Example-3: Bulk Search

To look up a lot of IP addresses,
you can pass them through stdin or a file.

```
# from standard input
$ cat IP-list.txt | geoipcli <options>

# from file
$ geoipcli <options> -readfile IP-list.txt
```


### Example-4: Configuration File

`geoipcli` reads configurations from files.

By default, `geoipcli` tries to read configurations from the following paths (if found).

* ~/.config/geoipcli.yaml
* ~/.geoipcli.yaml

The template of the configuration file is [here](https://github.com/md-irohas/geoipcli-go/blob/master/geoipcli.yaml.orig).

Also, you can pass the configuration file from the options (-conffile).

The configuration files will be merged (overwritten),
so I recommend you describe common configurations such as paths to databases in ~/.config/geoipcli.yaml,
and specific configurations such as column names to be printed in a file passed to arguments.


### Misc

By default, `geoipcli` halts when it gets an invalid IP address.
This is useful to prevent you from making errors in analysis.
If you want to skip invalid IP addresses, use `-skip-invalid-ip` option.

If you want to check the behavior of `geoipcli-go`, use `-debug` option.
The details of logs are outputted to stderr.


## Limitations

### Enterprise database

MaxMind provides GeoIP2-Enterprise database which includes data of other types of GeoIP databases.
I want to support it,
but unfortunately, `geoipcli-go` does not fully support the database.


## Alternatives

You might be able to use `geoiplookup` command instead.


## License

MIT License ([link](https://opensource.org/licenses/MIT)).


## Contact

md (E-mail: md.irohas at gmail.com)


