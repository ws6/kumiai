package loaf

import (
	"fmt"
	"io"
	"strconv"
	"strings"

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
}

type Loaf struct {
	Average float64
	Count   int64
}

func NewDefaultLoafParams() *LoafParams {

	return &LoafParams{
		ReadDepthCutOff:       DEFAULT_READ_DEPTH_CUTOFF,
		Filters:               []string{DEFAULT_FILTER},
		AlleleFrequencyCutoff: DEFAULT_ALLELE_FREQUENCY_CUTOFF,
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

func LoafOfVcf(file io.Reader, params *LoafParams) *Loaf {
	ret := new(Loaf)
	if params == nil {
		params = NewDefaultLoafParams()
	}
	rows := vcfreader.VCFParser(file)
	for row := range rows {
		if len(row) < 10 {
			continue
		}
		infoField := row[7]
		if f, err := getFloatFromInfoFiled(`DP`, infoField); err == nil {
			if f < params.ReadDepthCutOff {
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
		if len(sp2) < 9 {

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
