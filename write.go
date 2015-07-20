package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import "flag"
import "fmt"
import "os"
import "time"
import "github.com/bitly/go-simplejson"

func main() {
        dataPtr := flag.String("data", "", "the data to write")
        flag.Parse()
	data := *dataPtr
	if actionJson, err := simplejson.NewJson([]byte(data)); err != nil {
                panic(err)
        } else {
		ts, _ := actionJson.Get("_ts").Int()
		t := time.Unix(int64(ts), 0)
		file := fmt.Sprintf("data/%d-%02d-%02dT%02d", t.Year(), t.Month(), t.Day(), t.Hour())
		if _, err := os.Stat(file); err != nil {
			_, err := os.Create(file)
			if err != nil {
				panic(err)
			}
		}

		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
		    panic(err)
		}

		defer f.Close()
		if _, err = f.WriteString(data + "\n"); err != nil {
		    panic(err)
		}	
	}
}
