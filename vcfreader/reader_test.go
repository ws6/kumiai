package vcfreader

import (
	"testing"
)

func TestVcfReader(t *testing.T) {
	filename := `RTM-A002362.vcf.gz`

	ch := ParseFromFile(filename)
	for v := range ch {
		t.Logf(`%+v`, v)
	}
}
