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

var mapped []map[string]interface{} = make([]map[string]interface{}, 0)
var reduced map[string]map[string][]interface{} = make(map[string]map[string][]interface{})

func _eval(event map[string]interface{}, condition Condition) bool {
    if condition.left.NodeType == TYPE_PROPERTY {
        condition.left = evalPropertyNode(event, condition.left)
    }
    
    if condition.right.NodeType == TYPE_PROPERTY {
        condition.right = evalPropertyNode(event, condition.right)
    }

    if condition.left.NodeType == TYPE_NIL {
        condition.left.NodeType = condition.right.NodeType
    }

    if condition.right.NodeType == TYPE_NIL {
        condition.right.NodeType = condition.left.NodeType
    }

    op := condition.op
    switch(condition.left.NodeType) {
    case TYPE_INT:
        left := condition.left.IntVal
        switch (condition.right.NodeType) {
        case TYPE_INT:
            right := condition.right.IntVal
            return compareIntToInt(left, right, op)
        case TYPE_FLOAT:
            right := condition.right.FloatVal
            return compareIntToFloat(left, right, op)
        case TYPE_STRING:
            right := condition.right.StringVal
            return compareIntToString(left, right, op)
        }            
    case TYPE_FLOAT:
        left := condition.left.FloatVal
        switch (condition.right.NodeType) {
        case TYPE_INT:
            right := condition.right.IntVal
            return compareFloatToInt(left, right, op)
        case TYPE_FLOAT:
            right := condition.right.FloatVal
            return compareFloatToFloat(left, right, op)
        case TYPE_STRING:
            right := condition.right.StringVal
            return compareFloatToString(left, right, op)
        }            
    case TYPE_STRING:
        left := condition.left.StringVal
        switch (condition.right.NodeType) {
        case TYPE_INT:
            right := condition.right.IntVal
            return compareStringToInt(left, right, op)
        case TYPE_FLOAT:
            right := condition.right.FloatVal
            return compareStringToFloat(left, right, op)
        case TYPE_STRING:
            right := condition.right.StringVal
            return compareStringToString(left, right, op)
        }
    }
    panic("NodeType not supported")
}

func evalPropertyNode(event map[string]interface{}, node EvalNode) EvalNode {
    val := event[node.StringVal]
    if val == nil {
        return EvalNode{NodeType: TYPE_NIL}
    } else if intVal, ok := val.(int); ok {
        return EvalNode{NodeType: TYPE_INT, IntVal: intVal}
    } else if floatVal, ok := val.(float64); ok {
        return EvalNode{NodeType: TYPE_FLOAT, FloatVal: floatVal}
    } else if strVal, ok := val.(string); ok {
        return EvalNode{NodeType: TYPE_STRING, StringVal: strVal}
    } else {
        fmt.Println("Property value didn't match any types")
        return node
    } 
}

func _map(event string, mapper MapStatement) {
    var event_json map[string]interface{}
	if err := json.Unmarshal([]byte(event), &event_json); err == nil {
        var row map[string]interface{} = make(map[string]interface{})
        for _, field := range mapper.Fields {
            row[field] = event_json[field]
        }

        var match bool = true
        for _, condition := range mapper.Conditions {
            if !_eval(event_json, condition) {
                match = false
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
                reduced[key_val] = make(map[string][]interface{})
            }
            for field, val := range row {
                if field != reducer.Key {
                    if _, ok := reduced[key_val][field]; !ok {
                        reduced[key_val][field] = make([]interface{}, 0)
                    }
                    reduced[key_val][field] = append(reduced[key_val][field], val)
                }
            }    
        }
    }
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
    
    if reducer.Key == "" {
        if resultStr, err := json.Marshal(mapped); err != nil {
            panic(err)
        } else {
            fmt.Printf("%s", resultStr)
        }
    } else {
        _reduce(*reducer) 
        if resultStr, err := json.Marshal(reduced); err != nil {
            panic(err)
        } else {
            fmt.Printf("%s", resultStr)
        }
    }
}

