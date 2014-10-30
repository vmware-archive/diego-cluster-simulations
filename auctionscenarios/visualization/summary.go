package main

import (
	"encoding/csv"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
)

const LightLoad = "10% start"
const HeavyLoad = "cold start"
const RollingDeploy = "rolling deploy"

const CELLS = "cells"
const CONCURRENCY = "concurrency"
const BIDDING_POOL_FRACTION = "bidding_pool_fraction"
const ALGORITHM = "algorithm"
const SCENARIO = "scenario"
const NUM_AUCTIONS = "num_auctions"
const COMMUNICATIONS = "communications"
const WAIT_TIME = "wait_time"
const BIDDING_TIME = "bidding_time"
const SCORE = "score"
const NUM_MISSING = "num_missing"

type Summary struct {
	Cells               int
	Concurrency         int
	BiddingPoolFraction float64
	Algorithm           string
	Scenario            string
	NumAuctions         int
	Communication       int
	WaitTime            float64
	BiddingTime         float64
	Score               float64
	NumMissing          int
}

func (s Summary) Get(key string) interface{} {
	switch key {
	case CELLS:
		return s.Cells
	case CONCURRENCY:
		return s.Concurrency
	case BIDDING_POOL_FRACTION:
		return s.BiddingPoolFraction
	case ALGORITHM:
		return s.Algorithm
	case SCENARIO:
		return s.Scenario
	case NUM_AUCTIONS:
		return s.NumAuctions
	case COMMUNICATIONS:
		return s.Communication
	case WAIT_TIME:
		return s.WaitTime
	case BIDDING_TIME:
		return s.BiddingTime
	case SCORE:
		return s.Score
	case NUM_MISSING:
		return s.NumMissing
	default:
		log.Fatalf("Unkown key: %s", key)
	}
	return nil
}

func (s Summary) GetFloat(key string) float64 {
	value := reflect.ValueOf(s.Get(key))
	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.Int:
		return float64(value.Int())
	default:
		log.Fatalf("%s is not floatable (%s)", key, value.Type())
	}
	return 0
}

type Summaries []Summary

func ParseInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse int: %s", err.Error())
	}
	return int(i)
}

func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("Failed to parse float: %s", err.Error())
	}
	return f
}

func LoadSummaries(path string) Summaries {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open summary file: %s", err.Error())
	}
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("Failed to read summary file: %s", err.Error())
	}
	summaries := Summaries{}
	for _, record := range records[1:] {
		summary := Summary{
			Cells:               ParseInt(record[0]),
			Concurrency:         ParseInt(record[2]),
			BiddingPoolFraction: ParseFloat(record[3]),
			Algorithm:           record[4],
			Scenario:            record[5],
			NumAuctions:         ParseInt(record[6]),
			Communication:       ParseInt(record[7]),
			WaitTime:            ParseFloat(record[8]),
			BiddingTime:         ParseFloat(record[9]),
			Score:               ParseFloat(record[10]),
			NumMissing:          ParseInt(record[11]),
		}
		summaries = append(summaries, summary)
	}

	return summaries
}

func (s Summaries) Filter(key string, value interface{}) Summaries {
	summaries := Summaries{}
	for _, summary := range s {
		if reflect.DeepEqual(summary.Get(key), value) {
			summaries = append(summaries, summary)
		}
	}
	return summaries
}

func (s Summaries) XY(xKey string, yKey string) XYs {
	xy := make(XYs, len(s))
	for i, summary := range s {
		xy[i].X = summary.GetFloat(xKey)
		xy[i].Y = summary.GetFloat(yKey)
	}

	sort.Sort(xy)
	return xy
}

type XYs []struct{ X, Y float64 }

func (xys XYs) Len() int {
	return len(xys)
}
func (xys XYs) Swap(i, j int)      { xys[i], xys[j] = xys[j], xys[i] }
func (xys XYs) Less(i, j int) bool { return xys[i].X < xys[j].X }

func (xys XYs) XY(i int) (float64, float64) {
	return xys[i].X, xys[i].Y
}
