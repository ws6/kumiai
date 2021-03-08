package loaf

import (
	"testing"
)

func TestParseLoaf(t *testing.T) {
	// file := `../vcfreader/RTM-A002362.vcf.gz`
	file := `../vcfreader/VG0000349-CPC.hard-filtered.vcf.gz`

	ret := GetLoafFromFile(file, nil)
	t.Logf(`%+v`, ret)

}
