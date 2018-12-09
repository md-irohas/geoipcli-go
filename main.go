package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"os"
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
	records["anonymousip"] = geoip2.AnonymousIP{}
	records["asn"] = geoip2.ASN{}
	records["connection_type"] = geoip2.ConnectionType{}
	records["domain"] = geoip2.Domain{}
	records["isp"] = geoip2.ISP{}

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
	// Paths to GeoIP database files.
	paths := map[string]*string{}
	paths["country"] = flag.String("country", "", "Path to GeoIP2/GeoLite-Country database.")
	paths["city"] = flag.String("city", "", "Path to GeoIP2/GeoLite-City database.")
	paths["asn"] = flag.String("asn", "", "Path to GeoLite-ASN database.")
	paths["isp"] = flag.String("isp", "", "Path to GeoIP2-ISP database.")
	paths["domain"] = flag.String("domain", "", "Path to GeoIP2-Domain database.")
	paths["connection_type"] = flag.String("contype", "", "Path to GeoIP2-ConnectionType database.")
	paths["anonymousip"] = flag.String("anonymousip", "", "Path to GeoIP2-AnonymousIP database.")

	// Output params.
	var outputFormat, outputColumnString string
	// flag.StringVar(&outputFormat, "format", "csv", "Output format [csv/json].")
	flag.StringVar(&outputColumnString, "output", "", "Output columns separated by comma (,). See '--list-columns' option for more details.")

	// TODO: Add support of other output format.
	outputFormat = "csv"

	// Flags
	var showColumns, showVersion, skipInvalidIP bool
	flag.BoolVar(&showColumns, "list-columns", false, "Show all column names.")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit.")
	flag.BoolVar(&skipInvalidIP, "skip", false, "Skip ")
	flag.BoolVar(&Debug, "debug", false, "Run this program as debug mode.")

	// Files
	// conffile := flag.String("conffile", "", "Configuration file.")
	filename := flag.String("readfile", "", "Read IP addresses from file.")

	flag.Parse()

	if showColumns {
		PrintColumns()
		os.Exit(0)
	}

	if showVersion {
		PrintVersion()
		os.Exit(0)
	}

	// Reset log format.
	log.SetFlags(0)

	// Load GeoIP databases from files.
	// Blank paths are skipped, and the program exits if it fails to open GeoIP
	// databases. If the string of the output column is empty, build a list of
	// default output columns.
	useDefaultOutputColumns := (outputColumnString == "")

	dbs := map[string]*geoip2.Reader{}

	for key, path := range paths {
		if *path == "" {
			continue
		}

		db, err := geoip2.Open(*path)
		if err != nil {
			log.Fatal(err)
		}

		dbs[key] = db

		if useDefaultOutputColumns {
			switch key {
			case "country":
				outputColumnString += "country.country.iso_code,"
			case "city":
				outputColumnString += "city.country.iso_code,city.city.names.en,"
			case "asn":
				outputColumnString += "asn.autonomous_system_number,asn.autonomous_system_number,"
			case "isp":
				outputColumnString += "isp.autonomous_system_number,isp.autonomous_system_number,"
			case "domain":
				outputColumnString += "domain.domain,"
			case "connection_type":
				outputColumnString += "connection_type.connection_type,"
			case "anonymousip":
				outputColumnString += "anonymousip.is_anonymous,"
			}
		}
	}

	if useDefaultOutputColumns {
		outputColumnString = strings.Trim(outputColumnString, ",")
	}

	if len(dbs) == 0 {
		log.Fatal("No database.")
	}

	// Divide raw string of output columns to a list of column names, and check
	// the columns names.
	outputColumnString = strings.ToLower(outputColumnString)
	outputColumns := strings.Split(outputColumnString, ",")

	for _, column := range outputColumns {
		labels := strings.Split(column, ".")
		if len(labels) < 2 {
			log.Fatal("Invalid column name: ", column)
		}

		prefix := labels[0]
		switch prefix {
		case "country", "city", "asn", "isp", "domain", "connection_type", "anonymousip":
			if _, ok := dbs[prefix]; !ok {
				log.Fatal("Database corresponding to the column name not found: ", column)
			}
		default:
			log.Fatal("Unknown column name:", column)
		}
	}

	// Select writer.
	var writer *Writer

	switch outputFormat {
	case "csv":
		writer = NewCSVWriter(os.Stdout)
	default:
		log.Fatal("Unsupported output format: ", outputFormat)
	}

	// Array to hold the results of lookups.
	results := make([]string, len(outputColumns)+1)

	// Function to look up and print
	lookupAndPrint := func(address string) {
		address = strings.TrimSpace(address)
		ip := net.ParseIP(address)

		if ip == nil {
			if skipInvalidIP {
			} else {
				log.Fatal("Invalid IP address: ", address)
			}
		}

		results[0] = address

		for index, column := range outputColumns {
			var record interface{}
			var err error

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

			result := ""

			if err == nil {
				key := strings.Join(labels[1:], ".")
				flatRecord := Flatten(record)
				if r, ok := flatRecord[key].(string); ok {
					result = r
				}
			}

			results[index+1] = result
		}

		writer.Write(results)
	}

	// Read IP addresses from arguments or file (stdin/file), look up the IP
	// addresses from GeoIP databases, and print the results to stdout.
	args := flag.Args()

	if len(args) > 0 {
		for _, address := range args {
			lookupAndPrint(address)
		}
	} else {
		var fp *os.File
		var err error

		if *filename != "" {
			fp, err = os.Open(*filename)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
		} else {
			fp = os.Stdin
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
