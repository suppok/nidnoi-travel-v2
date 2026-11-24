package commons

import (
	"strings"
	"testing"
)

// itemFound: some returned line matches the golden item. A golden item may end with
// " :: kw1|alt, kw2" (items read from the itinerary, worded freely): one line must contain every
// comma-separated group, a group matching any of its "|" alternatives. Without keywords the line
// must contain the item's first 12 letters. Case, spaces and punctuation are ignored.
func itemFound(want string, got []string) bool {
	if _, keys, ok := strings.Cut(want, " :: "); ok {
		for _, g := range got {
			line, all := matchKey(g), true
			for _, group := range strings.Split(keys, ",") {
				any := false
				for _, alt := range strings.Split(group, "|") {
					if k := matchKey(alt); k != "" && strings.Contains(line, k) {
						any = true
						break
					}
				}
				all = all && any
			}
			if all {
				return true
			}
		}
		return false
	}
	key := []rune(matchKey(want))
	if len(key) > 12 {
		key = key[:12]
	}
	for _, g := range got {
		if strings.Contains(matchKey(g), string(key)) {
			return true
		}
	}
	return false
}

func TestItemFound(t *testing.T) {
	got := []string{
		"ค่าบัตรเข้าสวนสนุก ดิสนีย์แลนด์เซี่ยงไฮ้ พร้อมรถรับส่ง",
		"ตัวเลือกเสริม: บัตร Access ซื้อเองในแอป",
		"น้ำหนักกระเป๋าโหลดใต้ท้องเครื่องท่านละ 20 กิโลกรัม",
	}
	for want, ok := range map[string]bool{
		"ค่าบัตรเข้าสวนสนุก Shanghai Disneyland :: disneyland|ดิสนีย์":       true,
		"ตัวเลือกเสริม: บัตร Access :: ตัวเลือกเสริม, access":                true,
		"ตัวเลือกเสริม: ที่นั่งชั้นธุรกิจ :: ตัวเลือกเสริม, business|ธุรกิจ": false, // every group must match one line
		"น้ำหนักกระเป๋า โหลดใต้ท้องเครื่องท่านละ 20 KG.":                     true,  // first 12 letters
		"ค่าเข้าชมโตเกียวสกายทรี :: สกายทรี|skytree":                         false,
	} {
		if itemFound(want, got) != ok {
			t.Errorf("itemFound(%q) = %v, want %v", want, !ok, ok)
		}
	}
}

// noiseKeywords mark lines the customer summary must leave out (prompt: what every tour has or tells
// the customer nothing). Matched with matchKey, so spaces and punctuation don't matter.
var noiseKeywords = []string{
	"ค่าใช้จ่ายส่วนตัว", "ค่าทำหนังสือเดินทาง", "เหนือการควบคุม", "ภาษีมูลค่าเพิ่ม", "หักณที่จ่าย",
	"เกินกว่าสายการบิน", "น้ำดื่ม", "รถโค้ช", "starbucks", "covid", "ตัวเลือกเสริม",
}

// noiseItem returns the noise keyword a returned line contains, or "".
func noiseItem(line string) string {
	key := matchKey(line)
	for _, k := range noiseKeywords {
		if strings.Contains(key, matchKey(k)) {
			return k
		}
	}
	return ""
}

func TestNoiseItem(t *testing.T) {
	for line, noisy := range map[string]bool{
		"ค่าใช้จ่ายส่วนตัวของผู้เดินทาง เช่น ค่าโทรศัพท์":        true,
		"ค่าทำหนังสือเดินทาง ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า": true,
		"ค่าภาษีมูลค่าเพิ่ม 7% และหัก ณ ที่จ่าย 3%":              true,
		"ค่าเครื่องดื่มที่ Starbucks Reserve Roastery":           true,
		"ค่าวีซ่าและค่าดำเนินการยื่นขอวีซ่า":                     false,
		"Option: บัตร Access ดิสนีย์แลนด์":                       false,
		"ค่าทิปไกด์และคนขับรถท้องถิ่น 1,500 บาท/ท่าน/ทริป":       false,
	} {
		if (noiseItem(line) != "") != noisy {
			t.Errorf("noiseItem(%q) = %q, want noisy=%v", line, noiseItem(line), noisy)
		}
	}
}
