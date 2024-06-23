package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	VERSION  = "0.1.0"
	TERMINAL = "terminal"
	PIPE     = "pipe"
)

func main() {
	var (
		filename        string
		outputdelimiter string
		//inputdelimiter  string
		comma  rune
		help   bool
		sjis   bool
		rev    bool
		ver    bool
		reader *csv.Reader
	)

	flag.StringVar(&filename, "f", "", "File to read")
	//flag.StringVar(&inputdelimiter, "d", ",", "CSV Input Delimiter")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&sjis, "s", false, "Shift-JIS")
	flag.BoolVar(&rev, "r", false, "Reverse")
	flag.BoolVar(&ver, "v", false, "Version")
	flag.Parse()

	// Default : CSV to TSV
	// but when -r supplied : TSV to CSV
	if rev {
		//reader.Comma = '\t'
		comma = '\t'
		outputdelimiter = ","
	} else {
		// Default: CSV to TSV
		//reader.Comma = ','
		comma = ','
		/*
			c := []rune(inputdelimiter)
			fmt.Println(c)
			comma = c[0]
			fmt.Println(comma)
			fmt.Println(string(c))
		*/
		outputdelimiter = "\t"
	}

	// Input from file, not from PIPE
	if term.IsTerminal(int(syscall.Stdin)) {

		// Checking out args
		if help || len(os.Args) <= 1 {
			showhelp()
			os.Exit(0)
		}

		// Show version
		if ver {
			fmt.Printf("Version: %s\n", VERSION)
			os.Exit(0)
		}

		// Open CSV file
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// Print values
		if sjis {
			// SJIS -> UTF-8
			fmt.Println(comma)
			reader := csv.NewReader(transform.NewReader(f, japanese.ShiftJIS.NewDecoder()))
			reader.Comma = comma
			reader.LazyQuotes = true

			// Output
			printcsv(reader, outputdelimiter)
		} else {
			// UTF8 -> UTF-8
			reader = csv.NewReader(f)
			reader.Comma = comma
			reader.LazyQuotes = true

			// Output
			printcsv(reader, outputdelimiter)
		}
	} else {
		// Read data from PIPE
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		// Read stdin
		if sjis {
			// SJIS input
			reader = csv.NewReader(transform.NewReader(strings.NewReader(string(stdin)), japanese.ShiftJIS.NewDecoder()))
		} else {
			// Default, UTF-8 input
			reader = csv.NewReader(strings.NewReader(string(stdin)))
		}

		if rev {
			// TSV input & CSV output, if -r option supplied
			reader.Comma = '\t'
		} else {
			// Default: CSV to TSV,
			reader.Comma = ','
		}

		// Output
		printcsv(reader, outputdelimiter)
	}
}

func printcsv(csvreader *csv.Reader, csvdelimiter string) {
	for {
		line, err := csvreader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(line)
			panic(err)
		}

		for i, item := range line {
			if i < (len(line) - 1) {
				//fmt.Printf("%s: %s%s", i, item, csvdelimiter)  // debug
				fmt.Printf("%s%s", item, csvdelimiter)
			} else if i == (len(line) - 1) {
				fmt.Printf("%s", item)
			}
		}
		fmt.Printf("\n")
	}
}

func showhelp() {
	helpmsg := `# Covert CSV to TSV

## Usage 1 [FILE input]:
csv2tsv -f sample.csv [other flags]

  Flags: 
  -f sample.csv  Specify the file to import 
  -s             Import CSV in Shift-JIS encoding 
  -h             Show help
  -v             Show version

  Example: 
  Show "sample.csv" in TSV format
  csv2tsv -f sample.csv

## Usage 2 [PIPE(|) input]:
cat sample.csv | csv2tsv 

  Flags:
  -s  Shift-JIS encoding input

  > It works with UTF-8 (defautl) or Shift-JIS only.
  > Use iconv, etc. to convert encodings : 
  > iconv -f original_encoding -t utf8 sample.csv | csv2tsv 

## Usage 3 [Reverse input & output]:
By adding -r flag, you can reverse input and output; TSV to CSV. 
It works with both Terminal/PIPE & from FILE.

  Flags:
  -r  Reverse output; from TSV to CSV
  -s  Shift-JIS encoding input 

  Example: 
  Show "my_file.tsv" in CSV format
  cat sample.tsv | csv2tsv -r 
  or 
  csv2tsv -f sample-sjis.tsv -r -s 
`
	fmt.Println(helpmsg)
}
