package main

import (
	// "bytes"
    "os"
	"encoding/json"
	"log"
	"io"
	"fmt"
	"encoding/csv"
)

const (
	HEADER = -1
)

func main() {
 	// files, _ := ioutil.ReadDir(".")
 	// for _, f := range files {
 	// 	if filename := f.Name(); len(filename) > 4 && filename[len(filename) - 4:] == ".csv" {
  //           convertTsvToJson(f.Name())
 	// 	}
 	// }
    convertTsvToJson("redmonk-language-rankings.csv")

}

func convertTsvToJson(filename string) {
	fi, err := os.Open(filename)
   if err != nil { panic(err) }
   defer func() { if err := fi.Close(); err != nil { panic(err) } }()

	r := csv.NewReader(fi)
	r.Comma = '\t'
	lineCount := HEADER
	fieldMap := make(map[string]int)
	fieldList := make([]string, 0)
	records := make([]map[string]string, 0)
	for {
		record, err := r.Read()
    	if err == io.EOF { break // var EOF = errors.New("EOF")
    	} else if err != nil {
    		switch err := err.(type) {
    		case *csv.ParseError:
    			if err.Err == csv.ErrFieldCount {
    				// fmt.Println(filename)
    				// fmt.Println("ParseError:", err)
    			} else {
    				fmt.Println(filename)
    				fmt.Println("ParseError:", err)
    			}
    		default:
    			fmt.Println("Unknown Error:", err)
    		}
    	}
    	if lineCount == HEADER {
    		fmt.Println("WTF")
    		for i, field := range record {
    			fieldMap[field] = i
    			fieldList = append(fieldList, field)
    		}
    	} else {
    		records = append(records, make(map[string]string))
    		fmt.Println(len(records), lineCount)
    		for field_index, value := range record {
    			field_name := fieldList[field_index]
    			records[lineCount][field_name] = value
    		}
    	}
    	lineCount += 1
    }

	b, err := json.Marshal(records)
	if err != nil { log.Fatal(err) }
	fo, err := os.Create(filename + ".json")
   if err != nil { panic(err) }
   defer func() { if err := fo.Close(); err != nil { panic(err) } }()
   fo.Write(b)
   fmt.Println("done!")
	// fmt.Printf("%q", b);
}
