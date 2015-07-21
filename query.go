package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import "bytes"
import "bufio"
import "encoding/json"
import "flag"
import "fmt"
import "log"
import "os"
import "path/filepath"
import "time"
import "github.com/bitly/go-simplejson"

var mapped []map[string]interface{} = make([]map[string]interface{}, 0)
var reduced map[string]map[string]interface{} = make(map[string]map[string]interface{})

func _map(event string, mapper MapStatement) {
	if event_json, err := simplejson.NewJson([]byte(event)); err == nil {
        var row map[string]interface{} = make(map[string]interface{})
        for _, field := range mapper.Fields {
            val := event_json.Get(field)
            row[field] = val.Interface()
        }

        var match bool = true
        for _, condition := range mapper.Conditions {
            if left, err := event_json.Get(condition.left).Float64(); err != nil {
                //panic(err)
            } else {
                op := condition.op
                right := condition.right
                switch op {
                case GT:
                    if !(left > right) {
                        match = false
                    }
                    break
                case EQ:
                    if !(left == right) {
                        match = false
                    }
                    break
                }
            }
        }
        if match {
            mapped = append(mapped, row)
        }
	}
}

func _reduce(reducer ReduceStatement) {
    for _, row := range mapped {
        key_val, ok := row[reducer.Key].(string)
        if ok {
            if _, ok := reduced[key_val]; !ok {
                reduced[key_val] = make(map[string]interface{})
            }
            for field, val := range row {
                if _, ok := reduced[key_val][field]; !ok {
                    reduced[key_val][field] = float64(0)
                }
                if iVal, ok := val.(json.Number); ok {
                    if fVal, err := iVal.Float64(); err == nil {
                        reduced[key_val][field] = reduced[key_val][field].(float64) + fVal
                    }
                }
            }    
        }
    }
    fmt.Printf("%s", reduced)
}

func main() {
    queryPtr := flag.String("query", "", "Query to run. E.g. \"MAP field_1, field_2 REDUCE ON field_1\"")
    startPtr := flag.Int64("start", 0, "Start date (in seconds)")
    endPtr := flag.Int64("end", time.Now().Unix(), "End date (in seconds)")
    flag.Parse()
    startTm := time.Unix(*startPtr, 0)
    endTm := time.Unix(*endPtr, 0)
    startFile := GenerateFileName(startTm)
    endFile := GenerateFileName(endTm)
    dirname := "data" + string(filepath.Separator)

    query := bytes.NewBufferString(*queryPtr)
    p := NewParser(query)
    mapper, reducer, err := p.Parse()
    if err != nil {
        panic(err)
    }

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
                    _map(scanner.Text(), *mapper)
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

    _reduce(*reducer) 
}

