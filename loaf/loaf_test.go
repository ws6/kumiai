package loaf

import (
	"os"
	"testing"
)

func TestParseLoaf(t *testing.T) {
	file := `../vcfreader/RTM-A002362.vcf.gz`
	fh, err := os.Open(file)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer fh.Close()
	ret := LoafOfVcf(fh, nil)
	t.Logf(`%+v`, ret)

}
