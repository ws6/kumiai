package loaf

import (
	"fmt"
	"os"

	"strconv"
	"strings"

	"github.com/montanaflynn/stats"
	"github.com/ws6/kumiai/vcfreader"
)

//loaf.go  -- parse the .vcf file with specific formats then fitler the AF values
//Doc page -https://confluence.illumina.com/display/FBS/loaf+-+LOwer+half+Allele++Frequency

const (
	DEFAULT_ALLELE_FREQUENCY_CUTOFF = 0.5
	DEFAULT_READ_DEPTH_CUTOFF       = 20
	DEFAULT_FILTER                  = `PASS`
)

type LoafParams struct {
	ReadDepthCutOff       float64
	Filters               []string
	AlleleFrequencyCutoff float64

	LowAfFilter float64 //default 0.15

	HighAfPercent float64
	LowAfPercent  float64

	NumDataPointsDiscard int
}

type Loaf struct {
	Average float64
	Count   int64

	REFPercentile    float64
	ALTPercentile    float64
	ALTREFDiff       float64
	FilteredAfMedian float64
	OutLayerCount    int
}

func NewDefaultLoafParams() *LoafParams {

	return &LoafParams{
		ReadDepthCutOff:       DEFAULT_READ_DEPTH_CUTOFF,
		Filters:               []string{DEFAULT_FILTER},
		AlleleFrequencyCutoff: DEFAULT_ALLELE_FREQUENCY_CUTOFF,

		HighAfPercent:        25,
		LowAfPercent:         75,
		LowAfFilter:          float64(0.15),
		NumDataPointsDiscard: 3,
	}
}

func getFloatFromInfoFiled(fieldName, infoSection string) (float64, error) {
	sp := strings.Split(infoSection, ";")
	for _, sec := range sp {
		sp2 := strings.Split(sec, "=")
		if len(sp2) <= 1 {
			continue
		}
		k := sp2[0]
		if k != fieldName {
			continue
		}
		return strconv.ParseFloat(sp2[1], 64)
	}

	return 0, fmt.Errorf(`not found`)
}

func isInfilter(s string, filters []string) bool {
	for _, f := range filters {
		if s == f {
			return true
		}
	}
	return false
}

func GetLoafFromFile(filename string, params *LoafParams) *Loaf {

	ch := vcfreader.ParseFromFile(filename)
	return LoafOfVcf(ch, params)
}

func GetLoafFromFileV2(filename string, params *LoafParams) *Loaf {

	ch := vcfreader.ParseFromFile(filename)
	return LoafOfVcfV2(ch, params)
}

func LoafOfVcf(ch chan []string, params *LoafParams) *Loaf {
	ret := new(Loaf)
	if params == nil {
		params = NewDefaultLoafParams()
	}

	for row := range ch {

		if len(row) < 10 {
			continue
		}

		infoField := row[7]

		if f, err := getFloatFromInfoFiled(`DP`, infoField); err == nil {
			if f < params.ReadDepthCutOff {
				fmt.Println(`DP`, f)
				continue
			}
		}
		filtered := row[6]
		if !isInfilter(filtered, params.Filters) {
			continue
		}

		formatValueField := row[9]
		// GT:SQ:AD:AF:F1R2:F2R1:DP:SB:MB	0/1:12.58:2,1:0.333:1,0:1,1:3:1,1,1,0:2,0,1,0
		sp2 := strings.Split(formatValueField, ":")
		if len(sp2) < 4 {

			continue
		}
		af, err := strconv.ParseFloat(sp2[3], 64)
		if err == nil {
			if af > params.AlleleFrequencyCutoff {
				continue
			}
		}
		ret.Count++
		ret.Average += af

	}
	if ret.Count != 0 {
		ret.Average /= float64(ret.Count)
	}
	return ret
}

//v2 added the new values
func LoafOfVcfV2(ch chan []string, params *LoafParams) *Loaf {
	ret := new(Loaf)

	// ret.REFPercentile = 1
	// ret.ALTPercentile = 1
	// ret.FilteredAfMedian = 0
	if params == nil {
		params = NewDefaultLoafParams()
	}

	highaf := []float64{}
	lowaf := []float64{}

	for row := range ch {

		if len(row) < 10 {
			continue
		}

		infoField := row[7]
		if f, err := getFloatFromInfoFiled(`DP`, infoField); err == nil {
			if f < params.ReadDepthCutOff {
				fmt.Println(`DP`, f)
				continue
			}
		}
		filtered := row[6]
		if !isInfilter(filtered, params.Filters) {
			continue
		}

		formatValueField := row[9]
		// GT:SQ:AD:AF:F1R2:F2R1:DP:SB:MB	0/1:12.58:2,1:0.333:1,0:1,1:3:1,1,1,0:2,0,1,0
		sp2 := strings.Split(formatValueField, ":")
		if len(sp2) < 4 {

			continue
		}
		af, err := strconv.ParseFloat(sp2[3], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "strconv.ParseFloat:%s", err.Error())
			continue
		}

		if af <= params.LowAfFilter {
			ret.OutLayerCount++
			continue
		}

		if af >= params.AlleleFrequencyCutoff {
			highaf = append(highaf, af)
		}

		if af < params.AlleleFrequencyCutoff {
			lowaf = append(lowaf, af)
		}

		ret.Count++
		ret.Average += af

	}

	if ret.Count != 0 {
		ret.Average /= float64(ret.Count)
	}

	if len(lowaf) < params.NumDataPointsDiscard {
		lowaf = []float64{}
	}
	if len(highaf) < params.NumDataPointsDiscard {
		highaf = []float64{}
	}

	allaf := []float64{}
	allaf = append(allaf, highaf...)
	allaf = append(allaf, lowaf...)

	if f, err := stats.Median(allaf); err == nil {
		ret.FilteredAfMedian = f
	}

	if f, err := stats.Percentile(
		highaf,
		params.HighAfPercent,
	); err == nil {
		ret.ALTPercentile = f
	}

	if f, err := stats.Percentile(
		lowaf,
		params.LowAfPercent,
	); err == nil {
		ret.REFPercentile = f
	}

	ret.ALTREFDiff = ret.ALTPercentile - ret.REFPercentile

	if len(highaf) == 0 {
		ret.ALTREFDiff = ret.REFPercentile
	}
	if len(lowaf) == 0 {
		ret.ALTREFDiff = ret.ALTPercentile
	}

	if len(highaf) == 0 && len(lowaf) == 0 {
		ret.ALTREFDiff = 1 //hard coded value
	}

	return ret
}
