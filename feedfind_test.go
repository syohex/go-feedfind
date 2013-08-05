package feedfind

import (
	"testing"
)

func find(gots []string, expecteds ...string) bool {
	count := 0
	for _, expected := range expecteds {
		for _, got := range gots {
			if expected == got {
				count++
				break
			}
		}
	}

	if count == len(expecteds) {
		return true
	} else {
		return false
	}
}

func TestHatenablog(t *testing.T) {
	feeds, err := Find("http://shibayu36.hatenablog.com/")
	if err != nil {
		t.Errorf("Can't find rss at hatenablog")
	}

	expecteds := []string{
		"http://shibayu36.hatenablog.com/feed",
		"http://shibayu36.hatenablog.com/rss",
	}

	if !find(feeds, expecteds...) {
		t.Errorf("Can't find '/feed' and '/rss'")
	}
}

func TestLivedoor(t *testing.T) {
	feeds, err := Find("http://blog.livedoor.jp/xaicron")
	if err != nil {
		t.Errorf("Can't find rss at livedoor blog")
	}

	expecteds := []string{
		"http://blog.livedoor.jp/xaicron/index.rdf",
		"http://blog.livedoor.jp/xaicron/atom.xml",
	}
	if !find(feeds, expecteds...) {
		t.Errorf("Can't find '/feed' and '/rss'")
	}
}
