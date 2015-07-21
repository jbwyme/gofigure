package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import "encoding/json"
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
		actionArr, err := actionJson.Array()
		if err == nil {
			for _, action := range actionArr {
				jsonString, err := json.Marshal(action)
				if err != nil {
					fmt.Println(err)
				}
		
				writeAction(string(jsonString))
			} 
		} else {
			action, err := actionJson.MarshalJSON()
			if err == nil {
				writeAction(string(action))
			} else {
				panic(err)
			}
		}
	}
}

func writeAction(action string) {
		actionJson, err := simplejson.NewJson([]byte(action))
		ts, _ := actionJson.Get("_ts").Int()
		t := time.Unix(int64(ts), 0)
		file := fmt.Sprintf("data/%s", GenerateFileName(t))
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
		if _, err = f.WriteString(action + "\n"); err != nil {
		    panic(err)
		}	
}
