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
import "github.com/bitly/go-simplejson"

func processEvent(event string) {
	if event_json, err := simplejson.NewJson([]byte(event)); err == nil {
		user_id, _ := event_json.Get("user_id").String()
		item_count, _ := event_json.Get("item_count").Int();
		amount, _ := event_json.Get("amount").Float64();
		_, ok := results[user_id]
		if !ok {
			results[user_id] = new(UserResult)
		}
		user_result := results[user_id]
		user_result.ItemCount += item_count
		user_result.TotalSpent += amount
        } else {
		panic(err)
	}
}

type UserResult struct {
	ItemCount int
	TotalSpent float64
}

var results map[string]*UserResult = make(map[string]*UserResult)

func main() {
    //queryPtr := flag.String("query", "", "the query to run")
    flag.Parse()
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
	  if file, err := os.Open(dirname + f.Name()); err == nil {

	    defer file.Close()

	    scanner := bufio.NewScanner(file)
	    for scanner.Scan() {
		processEvent(scanner.Text())
	    }

	    // check for errors
	    if err = scanner.Err(); err != nil {
	      log.Fatal(err)
	    }

	  } else {
	    log.Fatal(err)
	  }
     }

     resultsJson, err := json.Marshal(results)
     if err == nil {
	fmt.Println(string(resultsJson))
     } else {
	panic(err)
     }
}

