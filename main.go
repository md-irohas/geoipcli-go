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
	VERSION = "0.1.0"
)

var (
	Debug = false
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

	fmt.Println("The following columns can be used for output (--output option).")
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
	fmt.Println(VERSION)
}

func main() {
	// Reset log format.
	log.SetFlags(0)

	// Paths to GeoIP database files.
	var countryPath, cityPath, asnPath, ispPath, domainPath, contypePath, anonymousipPath string
	//, enterprisePath string
	flag.StringVar(&countryPath, "country", "", "Path to GeoIP2/GeoLite-Country database.")
	flag.StringVar(&cityPath, "city", "", "Path to GeoIP2/GeoLite-City database.")
	flag.StringVar(&asnPath, "asn", "", "Path to GeoLite-ASN database.")
	flag.StringVar(&ispPath, "isp", "", "Path to GeoIP2-ISP database.")
	flag.StringVar(&domainPath, "domain", "", "Path to GeoIP2-Domain database.")
	flag.StringVar(&contypePath, "contype", "", "Path to GeoIP2-ConnectionType database.")
	flag.StringVar(&anonymousipPath, "anonymousip", "", "Path to GeoIP2-AnonymousIP database.")
	// flag.StringVar(&enterprisePath, "enterprise", "", "Path to GeoIP2-Enterprise database.")

	// Output params.
	var outputFormat, outputColumnString string
	// Current implementation supports only csv format as output. The following
	// lines will be fixed.
	// flag.StringVar(&outputFormat, "format", "", "Output format.")
	outputFormat = "csv"
	flag.StringVar(&outputColumnString, "output", "", "Output columns separated by comma (,). See '--list-columns' option for more details.")

	// Flags
	var showColumns, showVersion, skipInvalidIP bool
	flag.BoolVar(&showColumns, "list-columns", false, "Show all column names.")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit.")
	flag.BoolVar(&skipInvalidIP, "skip-invalid-ip", false, "Skip Invalid IP addresses.")
	flag.BoolVar(&Debug, "debug", false, "Run this program as debug mode w/ debug message.")

	// Files
	conffile := flag.String("conffile", "", "Config file.")
	readfile := flag.String("readfile", "", "Read IP addresses from file.")

	flag.Parse()

	if showVersion {
		PrintVersion()
		os.Exit(0)
	}

	if showColumns {
		PrintColumns()
		os.Exit(0)
	}

	// Divide raw string of output columns to a list of column names.
	var outputColumns []string
	if outputColumnString != "" {
		outputColumnString = strings.ToLower(outputColumnString)
		outputColumns = strings.Split(outputColumnString, ",")
	}

	// Load default configurations from predefined files, load configurations
	// from the specified file, and overwrite the arguments to the Config
	// variable.
	LoadDefaultConfigs()

	if *conffile != "" {
		filename, err := tilde.Expand(*conffile)
		if err != nil {
			log.Fatal("[-] Failed to expand filename:", err)
		}

		LoadConfig(filename)
	}

	paths := Config.Paths
	if countryPath != "" {
		paths.Country = countryPath
	}
	if cityPath != "" {
		paths.City = cityPath
	}
	if asnPath != "" {
		paths.ASN = asnPath
	}
	if ispPath != "" {
		paths.ISP = ispPath
	}
	if domainPath != "" {
		paths.Domain = domainPath
	}
	if contypePath != "" {
		paths.ConnectionType = contypePath
	}
	if anonymousipPath != "" {
		paths.AnonymousIP = anonymousipPath
	}
	// if enterprisePath != "" {
	// 	paths.Enterprise = enterprisePath
	// }

	output := Config.Output
	if outputFormat != "" {
		output.Format = outputFormat
	}
	if len(outputColumns) > 0 {
		output.Columns = outputColumns
	}
	if skipInvalidIP {
		output.SkipInvalidIP = skipInvalidIP
	}

	// If output columns are not given, use default list of colums.
	useDefaultOutputColumns := (len(output.Columns) == 0)
	if useDefaultOutputColumns {
		if Debug {
			log.Println("[*] Output columns not found. Use default columns.")
		}
	}

	// Load GeoIP databases from files. Blank paths are skipped, and the
	// program exits if it fails to open GeoIP databases. If the string of the
	// output column is empty, build a list of default output columns.
	dbpaths := map[string]string{
		"country":         paths.Country,
		"city":            paths.City,
		"asn":             paths.ASN,
		"isp":             paths.ISP,
		"domain":          paths.Domain,
		"connection_type": paths.ConnectionType,
		"anonymousip":     paths.AnonymousIP,
		// "enterprise":      paths.Enterprise,
	}

	var defaultColumns []string
	dbs := map[string]*geoip2.Reader{}

	for key, dbpath := range dbpaths {
		if dbpath == "" {
			continue
		}

		dbpath, err := tilde.Expand(dbpath)
		if err != nil {
			log.Fatal("[-] Failed to expand dbpath:", err)
		}

		db, err := geoip2.Open(dbpath)
		if err != nil {
			log.Fatal("[-] Failed to read GeoIP database:", err)
		}

		dbs[key] = db

		if useDefaultOutputColumns {
			switch key {
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
	}

	if useDefaultOutputColumns {
		output.Columns = defaultColumns
	}

	if len(dbs) == 0 {
		log.Fatal("[-] No databases.")
	}
	if len(output.Columns) == 0 {
		// Unreachable code.
		log.Fatal("[-] No output columns.")
	}

	for _, column := range output.Columns {
		labels := strings.Split(column, ".")
		if len(labels) < 2 {
			log.Fatal("[-] Invalid column name: ", column)
		}

		prefix := labels[0]
		switch prefix {
		case "country", "city", "asn", "isp", "domain", "connection_type", "anonymousip":
			if _, ok := dbs[prefix]; !ok {
				log.Fatal("[-] Database corresponding to the column name not found: ", column)
			}
		default:
			log.Fatal("[-] Unknown column name:", column)
		}
	}

	if Debug {
		log.Println("[+] Output:", strings.Join(output.Columns, ", "))
	}

	// Select writer.
	var writer *Writer

	switch output.Format {
	case "csv":
		writer = NewCSVWriter(os.Stdout)
	default:
		log.Fatal("[-] Unsupported output format: ", output.Format)
	}

	// Array to hold the results of lookups.
	results := make([]string, len(output.Columns)+1)

	// Function to look up the IP addresses on the GeoIP databases,
	// and print the lookup results.
	lookupAndPrint := func(address string) {
		address = strings.TrimSpace(address)
		ip := net.ParseIP(address)
		if ip == nil {
			if output.SkipInvalidIP {
				if Debug {
					log.Println("[-] Invalid IP address:", address)
				}
			} else {
				log.Fatalln("[-] Invalid IP address:", address)
			}
		}

		// The first columns is the IP address.
		results[0] = address

		// Look up the IP address based on the column name.
		for index, column := range output.Columns {
			var record interface{}
			var err error

			// Parse the column name, select the database, and look up the IP
			// address on the database. The GeoIP2-Enterprise database is a
			// all-in-one database, so if available, the database is used to
			// look up.
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
					log.Fatal("Unknown result type.", key, reflect.TypeOf(flatRecord[key]))
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
		for _, address := range args {
			lookupAndPrint(address)
		}
	} else {
		// from stdin/file
		var fp *os.File

		if *readfile == "" {
			// from stdin
			fp = os.Stdin
		} else {
			// from file
			filename, err := tilde.Expand(*readfile)
			if err != nil {
				log.Fatal("[-] Failed to expand filename:", err)
			}

			if Debug {
				log.Println("[+] Read IP addresses from:", filename)
			}

			fp, err = os.Open(filename)
			if err != nil {
				log.Fatal("[-] Failed to open file: ", err)
			}
			defer fp.Close()
		}

		s := bufio.NewScanner(fp)
		for s.Scan() {
			lookupAndPrint(s.Text())
		}

		if err := s.Err(); err != nil {
			log.Fatal(err)
		}
	}
}
