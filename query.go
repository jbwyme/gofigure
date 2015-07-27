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
import "reflect"
import "time"

var mapped []map[string]interface{} = make([]map[string]interface{}, 0)
var reduced map[string]map[string]interface{} = make(map[string]map[string]interface{})

func pluck(prop string, collection []interface{}) []interface{} {
	var res []interface{} = make([]interface{}, 0)
	for _, entry := range collection {
		if mapVal, ok := entry.(map[string]interface{}); ok {
			if val, ok := mapVal[prop]; ok {
				res = append(res, val)
			}
		}
	}
	return res
}

func _eval(event map[string]interface{}, condition Condition) bool {
	if condition.left.GetType() == TYPE_PROPERTY {
		condition.left = evalPropertyNode(event, condition.left)
	}

	if condition.right.GetType() == TYPE_PROPERTY {
		condition.right = evalPropertyNode(event, condition.right)
	}

	if condition.left.GetType() == TYPE_NIL {
		condition.left.SetType(condition.right.GetType())
	}

	if condition.right.GetType() == TYPE_NIL {
		condition.right.SetType(condition.left.GetType())
	}

	var l *Field
	if val, ok := condition.left.(*Field); ok {
		l = val
	} else {
		panic("left field is not a field")
	}

	var r *Field
	if val, ok := condition.right.(*Field); ok {
		r = val
	} else {
		panic("left field is not a field")
	}

	op := condition.op
	switch l.GetType() {
	case TYPE_INT:
		left := l.IntVal
		switch r.Type {
		case TYPE_INT:
			right := r.IntVal
			return compareIntToInt(left, right, op)
		case TYPE_FLOAT:
			right := r.FloatVal
			return compareIntToFloat(left, right, op)
		case TYPE_STRING:
			right := r.StringVal
			return compareIntToString(left, right, op)
		}
	case TYPE_FLOAT:
		left := l.FloatVal
		switch r.Type {
		case TYPE_INT:
			right := r.IntVal
			return compareFloatToInt(left, right, op)
		case TYPE_FLOAT:
			right := r.FloatVal
			return compareFloatToFloat(left, right, op)
		case TYPE_STRING:
			right := r.StringVal
			return compareFloatToString(left, right, op)
		}
	case TYPE_STRING:
		left := l.StringVal
		switch r.Type {
		case TYPE_INT:
			right := r.IntVal
			return compareStringToInt(left, right, op)
		case TYPE_FLOAT:
			right := r.FloatVal
			return compareStringToFloat(left, right, op)
		case TYPE_STRING:
			right := r.StringVal
			return compareStringToString(left, right, op)
		}
	}
	panic(fmt.Sprintf("Type %d not supported\n", l.GetType()))
}

// could probably remove this in favor of evalField
func evalPropertyNode(event map[string]interface{}, node IField) *Field {
	var f *Field
	if field, ok := node.(*Field); ok {
		f = field
	} else {
		panic("node is not Field")
	}

	val := event[f.StringVal]
	if val == nil {
		return &Field{Type: TYPE_NIL}
	} else if intVal, ok := val.(int); ok {
		return &Field{Type: TYPE_INT, IntVal: intVal}
	} else if floatVal, ok := val.(float64); ok {
		return &Field{Type: TYPE_FLOAT, FloatVal: floatVal}
	} else if strVal, ok := val.(string); ok {
		return &Field{Type: TYPE_STRING, StringVal: strVal}
	} else if listVal, ok := val.([]interface{}); ok {
		return &Field{Type: TYPE_LIST, ListVal: listVal}
	} else {
		fmt.Println("Property value didn't match any types")
		return &Field{}
	}
}

func evalField(event_json map[string]interface{}, field IField) interface{} {
	if f, ok := field.(*Field); ok {
		if _, ok := event_json[field.GetName()]; ok {
			return event_json[f.GetName()]
		} else {
			// fmt.Printf("No field %s in row %s\n", field.GetName(), event_json)
			return nil
		}
	} else if itr, ok := field.(*FieldItr); ok {
		collection := evalField(event_json, itr.Collection)
		if collection != nil {
			if collection, ok := collection.([]interface{}); ok {
				return pluck(itr.GetName(), collection)
			} else {
				panic(fmt.Sprintf("Expected evalField to return a list, instead got %s (type %s)", collection, reflect.TypeOf(collection)))
			}
		}
	} else if exp, ok := field.(*BinaryExpr); ok {
		left := evalField(event_json, exp.Left)
		right := evalField(event_json, exp.Right)
		switch exp.Operator {
		case MULTIPLY:
			var l float64 = 0
			var r float64 = 0
			if lInt, ok := left.(int); ok {
				l = float64(lInt)
			} else if lFloat, ok := left.(float64); ok {
				l = lFloat
			} else {
				// fmt.Printf("%s is not an int or float", left)
			}

			if rInt, ok := right.(int); ok {
				r = float64(rInt)
			} else if rFloat, ok := right.(float64); ok {
				r = rFloat
			} else {
				// fmt.Printf("%s is not an int or float", right)
			}
			return l * r
		}
	} else if agg, ok := field.(*Aggregator); ok {
		if collection, ok := evalField(event_json, agg.Target).([]interface{}); ok {
			switch agg.Method {
			case AGG_SUM:
				var sum float64 = 0
				for _, entry := range collection {
					if entryInt, ok := entry.(int); ok {
						sum += float64(entryInt)
					} else if entryFloat, ok := entry.(float64); ok {
						sum += entryFloat
					} else {
						// fmt.Printf("%s is not an int or float", entry)
					}
				}
				return sum
			case AGG_COUNT:
				return len(collection)
			default:
				panic(fmt.Sprintf("No aggregate method found for %d", agg.Method))
			}
		} else {
			// fmt.Printf("Can't aggregate %s because it's not a list. val: %s\n", agg.Target.GetName(), event_json)
		}
	}
	return nil
}

func _map(event string, mapper Statement) {
	var event_json map[string]interface{}
	if err := json.Unmarshal([]byte(event), &event_json); err == nil {
		var row map[string]interface{} = make(map[string]interface{})
		for _, field := range mapper.Fields {
			row[field.GetName()] = evalField(event_json, field)
		}

		var match bool = true
		for _, condition := range mapper.Conditions {
			if !_eval(row, condition) {
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
				reduced[key_val] = make(map[string]interface{})
			}
			for field, val := range row {
				if field != reducer.Key {
					if _, ok := reduced[key_val][field]; !ok {
						reduced[key_val][field] = make([]interface{}, 0)
					}
					if f, ok := reduced[key_val][field].([]interface{}); ok {
						reduced[key_val][field] = append(f, val)
					} else {
						panic(fmt.Sprintf("field %s on key %s in reducer is not a list", field, key_val))
					}
				}
			}
		}
	}

	for key, data := range reduced {
		for _, field := range reducer.GetFields() {
			reduced[key][field.GetName()] = evalField(data, field)
		}

		var match bool = true
		for _, condition := range reducer.Conditions {
			if !_eval(data, condition) {
				match = false
			}
		}

		if !match {
			delete(reduced, key)
		}
	}
}

func scanFile(file *os.File, mapper *Statement) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		_map(scanner.Text(), *mapper)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file.Close()
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
				scanFile(file, mapper)
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
