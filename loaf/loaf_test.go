package loaf

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var vcfFiles = func() func() []string {

	return func() []string {
		root := `../vcfreader`
		files, _ := ioutil.ReadDir(root)
		ret := []string{}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if strings.HasSuffix(f.Name(), ".vcf.gz") {
				ret = append(ret,

					filepath.Join(root, f.Name()),
				)
			}
		}

		return ret
	}
}()

func TestParseLoaf(t *testing.T) {
	// file := `../vcfreader/RTM-A002362.vcf.gz`

	for _, file := range vcfFiles() {
		t.Log(file)
		ret := GetLoafFromFileV2(file, nil)
		t.Logf(`%+v`, ret)
	}

}
