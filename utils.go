package main
import "fmt"
import "time"
func GenerateFileName(t time.Time) string {
    return fmt.Sprintf("%d-%02d-%02dT%02d", t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), t.UTC().Hour())
}
