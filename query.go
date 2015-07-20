package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import "bufio"
import "encoding/json"
import "flag"
import "fmt"
import "log"
import "os"
import "path/filepath"
import "time"
import "github.com/bitly/go-simplejson"

func processEvent(event string) {
	if event_json, err := simplejson.NewJson([]byte(event)); err == nil {
		user_id, _ := event_json.Get("user_id").String()
		item_count, _ := event_json.Get("total_item_count").Int();
		amount, _ := event_json.Get("amount").Float64();
		_, ok := results[user_id]
		if !ok {
			results[user_id] = new(UserResult)
		}
		user_result := results[user_id]
		user_result.ItemCount += item_count
		user_result.TotalSpent += amount
	}
}

type UserResult struct {
	ItemCount int
	TotalSpent float64
}

var results map[string]*UserResult = make(map[string]*UserResult)

func main() {
    startPtr := flag.Int64("start", 0, "Start date (in seconds)")
    endPtr := flag.Int64("end", time.Now().Unix(), "End date (in seconds)")
    flag.Parse()
    startTm := time.Unix(*startPtr, 0)
    endTm := time.Unix(*endPtr, 0)
    startFile := GenerateFileName(startTm)
    endFile := GenerateFileName(endTm)
    dirname := "data" + string(filepath.Separator)

     d, err := os.Open(dirname)
     if err != nil {
         fmt.Println(err)
         os.Exit(1)
     }

     defer d.Close()

     files, err := d.Readdir(-1)
     if err != nil {
         fmt.Println(err)
         os.Exit(1)
     }

    for _, f := range files {
        if startFile <= f.Name() && f.Name() <= endFile {
            if file, err := os.Open(dirname + f.Name()); err == nil {
                scanner := bufio.NewScanner(file)
                for scanner.Scan() {
                    processEvent(scanner.Text())
                }

                if err = scanner.Err(); err != nil {
                    log.Fatal(err)
                }

                file.Close()
            } else {
                log.Fatal(err)
            }
        }
    }
    
    for id, user := range results {
        if user.ItemCount < 10 || user.TotalSpent < 300 {
            delete(results, id)
        }
    }

    resultsJson, err := json.Marshal(results)
    if err == nil {
        fmt.Println(string(resultsJson))
    } else {
	    panic(err)
    }
}

