package vcfreader

import (
	"compress/gzip"
	"os"
	"testing"
)

func _TestVcfParseFromFile(t *testing.T) {
	filename := `STM-0000208-G09.hard-filtered.vcf.gz`

	ch := ParseFromFile(filename)
	for v := range ch {
		t.Logf(`%+v`, v)
	}
}
func TestVcfReader(t *testing.T) {
	filename := `STM-0000208-G09.hard-filtered.vcf.gz`
	file, err := os.Open(filename)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)

	if err != nil {
		t.Fatal(err.Error())
	}
	defer gz.Close()
	ch := VCFParser(gz)
	for v := range ch {
		t.Logf(`%+v`, v)
	}
}
