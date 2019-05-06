package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"gopkg.in/mattes/go-expand-tilde.v1"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	Version = "0.2.0"
)

var (
	// Flags.
	debug       = flag.Bool("debug", false, "Run this program as debug mode w/ messages.")
	showColumns = flag.Bool("list-columns", false, "Show all column names and exit.")
	showVersion = flag.Bool("version", false, "Show version and exit.")

	// Paths to databases.
	countryPath     = flag.String("country", "", "Path to GeoIP2/GeoLite2-Country database.")
	cityPath        = flag.String("city", "", "Path to GeoIP2/GeoLite2-City database.")
	asnPath         = flag.String("asn", "", "Path to GeoLite2-ASN database.")
	ispPath         = flag.String("isp", "", "Path to GeoIP2-ISP database.")
	domainPath      = flag.String("domain", "", "Path to GeoIP2-Domain database.")
	contypePath     = flag.String("contype", "", "Path to GeoIP2-Connection-Type database.")
	anonymousipPath = flag.String("anonymousip", "", "Path to GeoIP2-AnonymousIP database.")

	// Output params.
	outputFormat       = flag.String("format", "", "Output format (csv, tsv).")
	outputColumnString = flag.String("output", "", "Output columns separated by comma (,). See '-list-columns' option for more details.")
	escapeComma        = flag.Bool("do-not-escape-comma", true, "Do NOT escape commas in output.")
	escapeDoubleQuote  = flag.Bool("do-not-escape-double-quote", true, "Do NOT escape double quotes in output.")
	skipInvalidIP      = flag.Bool("skip-invalid-ip", false, "Skip invalid IP addresses.")

	// Files.
	conffile = flag.String("conffile", "", "Config file.")
	readfile = flag.String("readfile", "", "Read IP addresses from file.")
)

var (
	// List of paths to default config files. 
	configFiles = [...]string{
		"~/.config/geoipcli.yaml",
		"~/.config/geoipcli.yml",
		"~/.geoipcli.yaml",
		"~/.geoipcli.yml"}
	// This is a master config data. This data are overwritten by configFiles,
	// conffile, and arguments.
	config = &Config{}
	// List of GeoIP database types. This is required to print the database
	// types in order.
	dbtypes = [...]string {
		"country",
		"city",
		"asn",
		"isp",
		"domain",
		"connection_type",
		"anonymousip"}
)

func PrintColumns() {
	// The following records are derived from geoip2-golang library
	// (https://github.com/oschwald/geoip2-golang).
	records := map[string]interface{}{}
	records["country"] = geoip2.Country{}
	records["city"] = geoip2.City{}
	records["asn"] = geoip2.ASN{}
	records["isp"] = geoip2.ISP{}
	records["domain"] = geoip2.Domain{}
	records["connection_type"] = geoip2.ConnectionType{}
	records["anonymousip"] = geoip2.AnonymousIP{}

	fmt.Println("The following columns can be used for output (-output option).")
	fmt.Println("")
	fmt.Println("List of columns:")

	for name, record := range records {
		for key := range Flatten(record) {
			fmt.Println("- " + name + "." + key)
		}
	}

	fmt.Println("")
	fmt.Println("Note:")
	fmt.Println("- The [language] strings needs to be replaced by actual languages such as 'en' and 'ja'.")
	fmt.Println("")
}

func PrintVersion() {
	fmt.Println(Version)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	// Reset log format.
	log.SetFlags(0)

	// Parse command-line arguments.
	flag.Parse()

	// Print current version and exit.
	if *showVersion {
		PrintVersion()
		return
	}

	// Print list of columns and exit.
	if *showColumns {
		PrintColumns()
		return
	}

	// Divide raw string of output columns to a list of column names.
	var outputColumns []string
	if *outputColumnString != "" {
		*outputColumnString = strings.ToLower(*outputColumnString)
		outputColumns = strings.Split(*outputColumnString, ",")
	}

	// Load configs from predefined files (configFiles) and a file specified in
	// the command line, and merge the configs with other arguments in the
	// command-line.

	// Load configs from predefined files.
	for _, filename := range configFiles {
		filepath, err := tilde.Expand(filename)
		if err != nil {
			log.Fatalln("[-] Failed to expand path:", err)
		}

		if !fileExists(filepath) {
			if *debug {
				log.Println("[*] Skip loading default config (not found):", filename)
			}
			continue
		}

		err = LoadConfig(config, filepath)
		if err != nil {
			log.Fatalln("[-] Failed to load default config:", err)
		}

		if *debug {
			log.Println("[+] Load config:", filename)
		}
	}

	// Load configs from a file specified in arguments.
	if *conffile != "" {
		filepath, err := tilde.Expand(*conffile)
		if err != nil {
			log.Fatalln("[-] Failed to expand path:", err)
		}

		err = LoadConfig(config, filepath)
		if err != nil {
			log.Fatalln("[-] Failed to load config:", err)
		}

		if *debug {
			log.Println("[+] Load config:", *conffile)
		}
	}

	// Update paths of databases.
	paths := config.Paths
	if *countryPath != "" {
		paths.Country = *countryPath
	}
	if *cityPath != "" {
		paths.City = *cityPath
	}
	if *asnPath != "" {
		paths.ASN = *asnPath
	}
	if *ispPath != "" {
		paths.ISP = *ispPath
	}
	if *domainPath != "" {
		paths.Domain = *domainPath
	}
	if *contypePath != "" {
		paths.ConnectionType = *contypePath
	}
	if *anonymousipPath != "" {
		paths.AnonymousIP = *anonymousipPath
	}

	// Update output params.
	output := config.Output
	if *outputFormat != "" {
		output.Format = *outputFormat
	}
	if len(outputColumns) > 0 {
		output.Columns = outputColumns
	}
	output.EscapeComma = *escapeComma
	output.EscapeDoubleQuote = *escapeDoubleQuote
	output.SkipInvalidIP = *skipInvalidIP

	// Load GeoIP databases from files.
	// Blank paths are skipped. The program exits if it fails to open GeoIP
	// databases.
	dbpaths := map[string]string{
		"country":         paths.Country,
		"city":            paths.City,
		"asn":             paths.ASN,
		"isp":             paths.ISP,
		"domain":          paths.Domain,
		"connection_type": paths.ConnectionType,
		"anonymousip":     paths.AnonymousIP,
	}
	dbs := map[string]*geoip2.Reader{}

	for _, dbtype := range dbtypes {
		dbpath := dbpaths[dbtype]
		if dbpath == "" {
			continue
		}

		dbpath, err := tilde.Expand(dbpath)
		if err != nil {
			log.Fatalln("[-] Failed to expand path:", err)
		}

		db, err := geoip2.Open(dbpath)
		if err != nil {
			log.Fatalln("[-] Failed to read GeoIP database:", err)
		}

		if *debug {
			log.Printf("[+] Load GeoIP %s database: %s\n", dbtype, dbpath)
		}

		dbs[dbtype] = db
	}

	if len(dbs) == 0 {
		log.Fatalln("[-] No databases.")
	}

	// If output columns are not given, use default list of colums.
	// The list of columns depends on the type of databases.
	if len(output.Columns) == 0 {
		if *debug {
			log.Println("[*] Output columns not found. Use default columns.")
		}

		var defaultColumns []string
		for _, dbtype := range dbtypes {
			if _, ok := dbs[dbtype]; !ok {
				continue
			}

			switch dbtype {
			case "country":
				defaultColumns = append(defaultColumns, "country.country.iso_code")
			case "city":
				defaultColumns = append(defaultColumns, "city.country.iso_code", "city.city.names.en")
			case "asn":
				defaultColumns = append(defaultColumns, "asn.autonomous_system_number", "asn.autonomous_system_organization")
			case "isp":
				defaultColumns = append(defaultColumns, "isp.autonomous_system_number", "isp.autonomous_system_organization")
			case "domain":
				defaultColumns = append(defaultColumns, "domain.domain")
			case "connection_type":
				defaultColumns = append(defaultColumns, "connection_type.connection_type")
			case "anonymousip":
				defaultColumns = append(defaultColumns, "anonymousip.is_anonymous")
			}
		}

		output.Columns = defaultColumns
	}

	if *debug {
		log.Println("[+] Output:", strings.Join(output.Columns, ", "))
	}

	// Select writer and set up configuration of the writer.
	if output.Format == "" {
		output.Format = "csv"
	}
	var writer *Writer
	switch output.Format {
	case "csv":
		writer = NewCSVWriter()
	case "tsv":
		writer = NewTSVWriter()
	default:
		log.Fatalln("[-] Unsupported output format:", output.Format)
	}
	writer.EscapeComma = output.EscapeComma
	writer.EscapeDoubleQuote = output.EscapeDoubleQuote

	// Array to hold the results of lookups.
	results := make([]string, len(output.Columns)+1)

	// Function to look up the IP addresses on the GeoIP databases, and print
	// the lookup results.
	lookupAndPrint := func(address string) {
		address = strings.TrimSpace(address)
		ip := net.ParseIP(address)
		if ip == nil {
			if output.SkipInvalidIP {
				if *debug {
					log.Println("[-] Invalid IP address:", address)
				}
			} else {
				log.Fatalln("[-] Invalid IP address:", address)
			}
		}

		// The first column is the IP address.
		results[0] = address

		// Look up the IP address based on the column name.
		for index, column := range output.Columns {
			var record interface{}
			var err error

			// Parse the column name, select the database, and look up the IP
			// address on the database.
			//
			// e.g. country.country.names.en
			//        ^^
			//      prefix
			labels := strings.Split(column, ".")

			prefix := labels[0]
			switch prefix {
			case "country":
				record, err = dbs["country"].Country(ip)
			case "city":
				record, err = dbs["city"].City(ip)
			case "asn":
				record, err = dbs["asn"].ASN(ip)
			case "isp":
				record, err = dbs["isp"].ISP(ip)
			case "domain":
				record, err = dbs["domain"].Domain(ip)
			case "connection_type":
				record, err = dbs["connection_type"].ConnectionType(ip)
			case "anonymousip":
				record, err = dbs["anonymousip"].AnonymousIP(ip)
			}

			// Extract the result of the lookup from the record, and store it
			// to results array.
			var result string

			if err == nil {
				key := strings.Join(labels[1:], ".")
				flatRecord := Flatten(record)
				switch flatRecord[key].(type) {
				case string:
					if res, ok := flatRecord[key].(string); ok {
						result = res
					}
				case uint:
					if res, ok := flatRecord[key].(uint); ok {
						result = strconv.FormatUint(uint64(res), 10)
					}
				case float64:
					if res, ok := flatRecord[key].(float64); ok {
						result = strconv.FormatFloat(res, 'f', -1, 64)
					}
				case bool:
					if res, ok := flatRecord[key].(bool); ok {
						result = strconv.FormatBool(res)
					}
				case nil:
					result = ""
				default:
					log.Fatalln("[-] Unknown result type:", key, reflect.TypeOf(flatRecord[key]))
				}
			} else {
				result = ""
			}

			results[index+1] = result
		}

		writer.Write(results)
	}

	// Read IP addresses from arguments/stdin/file, look up the IP addresses on
	// the GeoIP databases, and print the results to stdout.
	args := flag.Args()

	if len(args) > 0 {
		// from arguments
		if *debug {
			log.Println("[+] Read IP addresses from arguments.")
		}

		for _, address := range args {
			lookupAndPrint(address)
		}
	} else {
		// from stdin/file
		var fp *os.File

		if *readfile == "" {
			// from stdin
			if *debug {
				log.Println("[+] Read IP addresses from stdin.")
			}

			fp = os.Stdin
		} else {
			// from file
			if *debug {
				log.Println("[+] Read IP addresses from file:", *readfile)
			}

			filename, err := tilde.Expand(*readfile)
			if err != nil {
				log.Fatalln("[-] Failed to expand filename:", err)
			}

			if *debug {
				log.Println("[+] Read IP addresses from:", filename)
			}

			fp, err = os.Open(filename)
			if err != nil {
				log.Fatalln("[-] Failed to open file: ", err)
			}
			defer fp.Close()
		}

		s := bufio.NewScanner(fp)
		for s.Scan() {
			lookupAndPrint(s.Text())
		}

		err := s.Err()
		if err != nil {
			log.Fatalln(err)
		}
	}
}
