package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import "flag"
import "fmt"
import "os"
import "strings"
import "time"
import "github.com/bitly/go-simplejson"

func main() {
	dataPtr := flag.String("data", "", "the data to write")
	flag.Parse()
	t := time.Now()
	filename := fmt.Sprintf("%d-%02d-%02dT%02d", t.Year(), t.Month(), t.Day(), t.Hour())
	validateEvent(*dataPtr)
	appendToFile(fmt.Sprintf("data/%s", filename), *dataPtr)
}

func validateEvent(data string) {
	if _, err := simplejson.NewJson([]byte(data)); err != nil {
		panic(err)
	}
}

func appendToFile(file string, data string) {
	if _, err := os.Stat(file); err != nil {
		os.Create(file)
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
	    panic(err)
	}

	defer f.Close()
	s := []string{data, "\n"};
	if _, err = f.WriteString(strings.Join(s, "")); err != nil {
	    panic(err)
	}	
}
