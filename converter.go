package main

import (
"os"
"io"
"fmt"
"encoding/json"
"encoding/csv"
"io/ioutil"
"strconv"
"sort"
"math"
"math/rand"
"strings"
// "reflect"
)

func loadFile(foldername, filename string, fwl FrameworkList, maxT, maxL []float64, github map[string]Github, redmonk map[string]Redmonk) FrameworkList {
    arIndex, arType := indexTypeFromFileName(filename)
	b, err := os.Open(foldername + "/" + filename)
	if err != nil { panic(err) }
	if b == nil { panic(err) }
	r := csv.NewReader(b)
	r.Comma = '\t'
	lineCount := 0
	field := make(map[string]int)

	for {
		record, err := r.Read()
    	if err == io.EOF { break // var EOF = errors.New("EOF")
    	} else if err != nil {
    		switch err := err.(type) {
    		case *csv.ParseError:
    			if err.Err != csv.ErrFieldCount {
                    fmt.Println(filename)
    				fmt.Println("ParseError:", err)
    			}
    		default:
    			fmt.Println("Unknown Error:", err)
    		}
    	} 
    	if lineCount == 0 {
    		for i, f := range record {
    			field[f] = i
    		}
            lineCount += 1
    		continue
    	}
        lineCount += 1

        // EXTRACT DATA IN ALL CASES
        name := record[field["Framework"]]
        errors, _ := strconv.ParseFloat(record[field["Errors"]], 64)
        hps := FNULL
        // throughput := FNULL
        // percentOfHighest :=FNULL
        avgLatency := FNULL
        latencySD := FNULL
        latency := FNULL
        var lng, cls, plt, db, orm string
        if arType == "latency" {
            avgLatency, _ = strconv.ParseFloat(record[field["Avg Latency"]][:len(record[field["Avg Latency"]]) - 3], 64)
            // percentOfHighest, _ = strconv.ParseFloat(record[field["PercentOfHighest"]][:len(record[field["PercentOfHighest"]]) - 1], 64)
            latencySD, _ = strconv.ParseFloat(record[field["Latency SD"]][:len(record[field["Latency SD"]]) - 3], 64)
            latency = avgLatency + latencySD
            // if lineCount > 20 && maxL[arIndex] * 1.75 < latency { // rule about not including latencies that are just too big (works since inupts are in order from least to most)
            //     continue
            // }
        } else if arType[:10] == "throughput" {
            hps, _ = strconv.ParseFloat(strings.Replace(record[field["Hps"]], ",", "", -1), 64)
            if hps == 0 { hps = FNULL }
            // throughput, _ = strconv.ParseFloat(record[field["Percent"]][:len(record[field["Percent"]]) - 1], 64)
            lng = record[field["Lng"]]
            cls = record[field["Cls"]]
            plt = record[field["Plt"]]
            if arType == "throughputdb" {
                db = record[field["DB"]]
                orm = record[field["Orm"]]
            }
        }

        // MATCHER: update existing framework if name, and other strings match       LATER (or are empty????)
        alreadyUpdated := false
        for i, fw := range fwl {
            if fw.N == name {
                if arType == "latency" {
                    alreadyUpdated = true
                    fw.LL[arIndex] = latency
                    maxL[arIndex] = math.Max(maxL[arIndex], latency)
                    // fw.E += errors ... this is duplicate information from the non-latency files
                    fwl[i] = fw
                } else if (fw.LN == "" || fw.LN == redmonk[lng].Name && fw.CS == cls && fw.P == plt && // b/c only Svt--Jty, Mon--Net differences*/
                           (fw.D == "" || db == "" || fw.D == db && fw.O == orm)) {
                    alreadyUpdated = true
                    fw.E += errors
                    fw.TL[arIndex] = hps
                    maxT[arIndex] = math.Max(maxT[arIndex], hps)
                    fw.LN = redmonk[lng].Name
                    fw.LP = redmonk[lng].Popularity
                    fw.CS = cls
                    fw.P  = plt
                    if db != "" {
                        fw.D = db
                        fw.O = orm
                    }
                    fwl[i] = fw
                } else if fw.D == db && fw.O == orm { // if fw.N == "grails" {
                    fmt.Println(filename)
                    fmt.Println(fw.N)
                    fmt.Println(fw.LN == redmonk[lng].Name, fw.CS == cls, fw.P == plt, fw.D == db, fw.O == orm)
                    fmt.Println(fw.LN + "--" +  redmonk[lng].Name, fw.CS + "--" +  cls, fw.P + "--" +  plt, fw.D + "--" +  db, fw.O + "--" +  orm)

                }

            }
        }
        if !alreadyUpdated {
            fw := Framework{
                name,
                rand.Intn(100000000), // ID
                FNULL,                // score
                SNULL,                // class
                SNULL,                // database
                SNULL,                // orm
                SNULL,                // platform
                SNULL,                // language
                github[name].Forks,   // Forks
                github[name].Stars,   // Stars
                FNULL,                // Language Rank
                mkNullFloatSlice(18), // throughput list
                mkNullFloatSlice(18), // latency list
                errors,                    // errors
            }

            if arType == "latency" {
                fw.LL[arIndex] = avgLatency + latencySD
                maxL[arIndex] = math.Max(maxL[arIndex], avgLatency + latencySD)
            } else { // must be throughput
                fw.TL[arIndex] = hps
                maxT[arIndex] = math.Max(maxT[arIndex], hps)
                fw.LN = redmonk[lng].Name
                fw.LP = redmonk[lng].Popularity
                fw.CS = cls
                fw.P  = plt
                if db != "" {
                    fw.D = db
                    fw.O = orm
                }
            }
            fwl = append(fwl, fw)
        }
    }
    return fwl
}

func indexTypeFromFileName(fn string) (int64, string) {
    basefn := strings.Split(fn, ".")
    ar := strings.Split(basefn[0], "-")
    var index int64
    var arType string
    switch ar[0] {
        case "json":             index = 0;  arType = "throughput"
        case "single query":     index = 3;  arType = "throughputdb"
        case "multiple queries": index = 6;  arType = "throughputdb"
        case "fortunes":         index = 9;  arType = "throughputdb"
        case "data updates":     index = 12; arType = "throughputdb"
        case "plaintext":        index = 15; arType = "throughput"
    }
    switch ar[1] {
        case "i7":   index += 0
        case "ec2":  index += 1
        case "peak": index += 2
    }
    if len(ar) == 3 {
        return index, "latency"
    } else {
        return index, arType
    }
}

func csvToRows(filename string, maxByIndex map[int]float64) (records [][]string) {
    fi, err := os.Open(filename)
    if err != nil { panic(err) }
    defer func() { if err := fi.Close(); err != nil { panic(err) } }()

    r := csv.NewReader(fi)
    r.Comma = '\t'
    lineCount := HEADER
    records = make([][]string,0)
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
        if lineCount != HEADER {
            for i, value := range record {
                if _, exists := maxByIndex[i]; exists {
                    floatValue, _ := strconv.ParseFloat(value, 64)
                    if floatValue > maxByIndex[i] {
                        maxByIndex[i] = floatValue
                    }
                }
            }
            records = append(records, record)
        }
        lineCount += 1
    }
    return records
}

func recalc(fwl FrameworkList, maxT, maxL []float64) RowList {
    maxCommunity := 0.
    maxLatency := 0.
    maxThroughput := 0.
    count := 0.
    totalHps := 0.
    fixture := make(RowList, 0)
    for i, fw := range fwl {
        // normalize language percents
        fw.F = float64(int(fw.F * 100))
        fw.SR = float64(int(fw.SR * 100))
        fw.LP = float64(int(fw.LP * 100))
        fixture = append(fixture, Row{fw.N, fw.ID, 0, 0, 0, 0, 0, 0, 0, 0, fw.CS, p2Platform(fw.P, fw.N), d2DatabaseName(fw.D), fw.O, fw.LN})
        fx := fixture[i]
        fx.C = fw.F * 2 + fw.SR / 2 + fw.LP
        count = 0
        totalHps = 0
        for i, v := range fw.TL {
            if v > 0 {
                totalHps += v
                fx.T += v / maxT[i]
                count += 1
            }
        }
        fx.NU = int(count)
        if count == 0 { fx.T = -1 } else { fx.T /= count }
        count = 0
        for i, v := range fw.LL {
            if v > 0 {
                fx.L += v / maxL[i]
                count += 1
            }
        }
        if count == 0 { fx.L = -1 } else { fx.L /= count }
        fx.NU += int(count)
        fx.E = totalHps / (totalHps + fw.E)
        fixture[i] = fx
        maxCommunity = math.Max(fx.C, maxCommunity)
        maxThroughput = math.Max(fx.T, maxThroughput)
        maxLatency = math.Max(fx.L, maxLatency)
        fwl[i] = fw
    }
    maxScore := 0.
    for i, fx := range fixture {
        fx.T = fx.T / maxThroughput; fx.T = float64(int(fx.T * 100))
        if fx.L > 0 {
            fx.L = (maxLatency - fx.L) / maxLatency; fx.L = float64(int(fx.L * 100))
        } else {
            fx.L = 0
        }
        fx.E = float64(int(fx.E * 100000)) / 1000
        fx.C = fx.C / maxCommunity; fx.C = float64(int(fx.C * 100))
        fx.S = fx.T * 2    +    fx.L    +    fx.C / 2    +     fx.E
        maxScore = math.Max(fx.S, maxScore)
        fixture[i] = fx
    }
    for i, fx := range fixture {
        fx.S /= maxScore
        fx.S = float64(int(fx.S * 100))
        fixture[i] = fx
        fw := fwl[i]
        fw.S = fx.S
        fwl[i] = fw
    }
    return fixture
}

func main() {
    github := make(map[string]Github)
    maxByIndex := map[int]float64{ 1: 0, 2: 0}
    for _, row := range csvToRows("github.csv", maxByIndex)  {
        stars, _ := strconv.ParseFloat(row[1], 64)
        forks, _ := strconv.ParseFloat(row[2], 64)
        github[row[0]] = Github{stars / maxByIndex[1], forks / maxByIndex[2]}
    }
    redmonk := make(map[string]Redmonk)
    maxByIndex = map[int]float64{ 2: 0}
    for _, row := range csvToRows("redmonk-language-rankings.csv", maxByIndex) {
        rank, _ := strconv.ParseFloat(row[2], 64)
        redmonk[row[1]] = Redmonk{row[0], (maxByIndex[2] - rank) / maxByIndex[2]}
    }

    maxT := make([]float64,18)
    maxL := make([]float64,18)
    fwl := make(FrameworkList, 0)
    foldername := "r09"
 	files, _ := ioutil.ReadDir("./"+ foldername +"/")
 	for _, file := range files {
 		if filename := file.Name(); len(filename) > 4 && filename[len(filename) - 4:] == ".csv" {
            fwl = loadFile(foldername, file.Name(), fwl, maxT, maxL, github, redmonk)
 		}
 	}
    fixture := recalc(fwl, maxT, maxL)
    sort.Sort(sort.Reverse(fixture))
    sort.Sort(sort.Reverse(fwl))
    // r09 := Round{fwl, maxT, maxL, fixture}

    // add in historical data
    pastMaxT := make([]float64,18)
    pastMaxL := make([]float64,18)
    pastFwl := make(FrameworkList, 0)
    foldername = "r08"
    files, _ = ioutil.ReadDir("./"+ foldername +"/")
    for _, file := range files {
        if filename := file.Name(); len(filename) > 4 && filename[len(filename) - 4:] == ".csv" {
            pastFwl = loadFile(foldername, file.Name(), pastFwl, pastMaxT, pastMaxL, github, redmonk)
        }
    }
    frameworkDeltas(fixture, fwl, pastFwl)


    // COPY PASTE
    filename := "cp.json"
    if _, err := os.Stat(filename); os.IsExist(err) {
        os.Remove(filename)
    }
    b, err := json.Marshal(fwl)
    if err != nil { panic(err) }
    str := "        raw : " + strings.Replace(string(b),"\"", "'", -1) + ",\n"
    b, err = json.Marshal(fixture)
    if err != nil { panic(err) }
    str += "        fixture : " + strings.Replace(string(b),"\"", "'", -1) + ",\n"
    b, err = json.Marshal(maxT)
    if err != nil { panic(err) }
    str += "        maxT : " + strings.Replace(string(b),"\"", "'", -1) + ",\n"
    b, err = json.Marshal(maxL)
    if err != nil { panic(err) }
    str += "        maxL : " + strings.Replace(string(b),"\"", "'", -1) + ",\n"

    // b, err := json.Marshal(r09)
    // if err != nil { panic(err) }
    // str := "        r09 : " + strings.Replace(string(b),"\"", "'", -1) + ",\n"

    fo, err := os.Create(filename)
    if err != nil { panic(err) }
    defer func() { if err := fo.Close(); err != nil { panic(err) } }()
    fo.WriteString(str)



    // .Reverse inside .Sort
    // composite := func(a, b *Framework) bool { return a.S < b.S }
    // By(composite).Sort(fwl)

    sort.Sort(fixture)

    fmt.Println("------------------------------------------------------------------------------------------")
 	// for _, fx := range fixture {
  //       fmt.Printf("|% 20s \t| %2.1f \t| %2.1f \t| %2.1f \t| %2.1f \t| %2.1f \t|\n", fx.N, fx.S, fx.L, fx.C, fx.T, fx.E)
 	// }

 }

func frameworkDeltas(fixture RowList, fwl, pastFwl FrameworkList) {
    for i, fw := range fwl {
        for _, pfw := range pastFwl {
            if fw.N == pfw.N {
                fixture[i].TD = 0
                fixture[i].LD = 0
                count := 0.
                Lcount := 0.
                for j := 0; j < 18; j++ {
                    if fw.TL[j] > 0 && pfw.TL[j] > 0 {
                        fixture[i].TD += (fw.TL[j] - pfw.TL[j]) / pfw.TL[j]
                        count++
                    }
                    if fw.LL[j] > 0 && pfw.LL[j] > 0 {
                        fixture[i].LD += (fw.LL[j] - pfw.LL[j]) / pfw.LL[j]
                        Lcount++
                    }
                }
                if count != 0 {
                    fixture[i].TD = float64(int(100 * fixture[i].TD / count))
                }
                if Lcount != 0 {
                    fixture[i].LD = float64(int(100 * fixture[i].LD / Lcount))
                }
                break
            }
        }
    }
}


func objectToJsonFile(object interface{}, filename string) {
    b, err := json.Marshal(object)
    if err != nil { panic(err) }
    fo, err := os.Create(filename)
    if err != nil { panic(err) }
    defer func() { if err := fo.Close(); err != nil { panic(err) } }()
    fo.Write(b)

}

func d2DatabaseName(s string) string {
    switch s {
    case "Pg": return "Postgres"
    case "Mo": return "MongoDB"
    case "My": return "MySQL"
    }
    return s
}

func p2Platform(s string, name string) string {
    switch s {
    case "Cpl": return "C++"
    case "Tre": return "Treefrog"
    case "Go": return "Go"

 // Netty || Jetty || Servlet || JRuby || Untertow || Plain || http-kit || Activeweb
    case "Svt": return "Servlet"
    case "Nty": return "Netty"
    case "Jty": return "Jetty"
    case "JRb": return "JRuby"
    case "Utw": return "Undertow"
    case "Und": return "Undertow Edge"
    case "Pla": return "Plain Scala" // ??????
    case "htk": return "http-kit"
    case "Act": return "Activeweb"

    case "Mon": return "Mono"
    case "Net": return "Dot Net"

    case "hhv": return "HHVM"
    case "FPM": return "PHP-FPM"
    
    case "Rac": return "Rack"
    case "njs": return "Node JS"
    case "Rin": return "RingoJS"
    case "Dar": return "Dart"

    case "Nim": return "Nimrod"
    case "Oni": return "Onion"

    case "Tor": return "Tornado"
    case "Wsg": return "Wsgi"
    case "uWS": return "uWsgi"
    case "Gun": return "GUnicorn"

    case "Cow": return "Cowboy"
    case "eli": return "Elli"

    case "Ur/": return "Ur"
    case "OpR": return "Openresty"
    case "Plk": return "Perl"
     // Haskell
    case "Snp": return "Snap"
    case "Wai": return "Wai"
    }
    fmt.Println(s + " - " + name)
    return s
}
