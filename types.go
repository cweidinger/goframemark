package main

import (
"sort"
)

const (
    HEADER = -1
    SNULL = ""
    FNULL float64 = -1
)

type Framework struct {
    N  string    // Name
    ID int     // Unique ID
    S  float64   // Composite
    CS string    // Class
    D  string    // database
    O  string    // orm
    P  string    // platform
    LN string    // language
    F  float64   // Forks
    SR float64   // Stars
    LP float64   // Language Popularity
    TL []float64 // throughput list
    LL []float64 // latency list
    E  float64     // errors
}

type FrameworkList []Framework
func (p FrameworkList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p FrameworkList) Len() int { return len(p) }
func (p FrameworkList) Less(i, j int) bool { return p[i].S < p[j].S }

type Row struct {
    N  string
    ID int
    S  float64   // score
    C  float64   // community result
    T  float64   // throughput result
    TD float64
    L  float64   // latency result
    LD float64
    E  float64   // errors
    NU int     // number of records
    CS string
    P  string
    D  string
    O  string
    LN string
}

type RowList []Row
func (p RowList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p RowList) Len() int { return len(p) }
func (p RowList) Less(i, j int) bool { return p[i].S < p[j].S }


type By func(a, b *Framework) bool
func (by By) Sort(fwList FrameworkList) {
    fs := &fwSorter{
        fwList: fwList,
        by:      by, // The Sort method's receiver is the function (closure) that defines the sort order.
    }
    sort.Sort(fs)
}
type fwSorter struct {
    fwList FrameworkList
    by      func(a, b *Framework) bool
}
func (s *fwSorter) Len() int { return len(s.fwList) }
func (s *fwSorter) Swap(i, j int) { s.fwList[i], s.fwList[j] = s.fwList[j], s.fwList[i] }
func (s *fwSorter) Less(i, j int) bool { return s.by(&s.fwList[i], &s.fwList[j]) }

type Round struct {
    Raw     FrameworkList
    MaxT    []float64
    MaxL    []float64
    Fixture RowList
}

type Github struct {
    Stars float64
    Forks float64 
}

type Redmonk struct {
    Name string
    Popularity float64
}



func mkNullFloatSlice(size int64) []float64 {
    res := make([]float64, size)
    for i := range res {
        res[i] = FNULL
    }
    return res
}

