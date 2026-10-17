package targets

import "testing"

func TestBoards(t *testing.T) {
	names, err := Names()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 4 {
		t.Fatalf("board 数量异常：%v", names)
	}
	b, ok := Find("esp32s3")
	if !ok {
		t.Fatal("esp32s3 未登记")
	}
	if !b.Embedded || b.Arch != "xtensa" {
		t.Fatalf("esp32s3 字段异常：%+v", b)
	}
	if b.Triple == nil || *b.Triple != "xtensa-esp32s3-elf" {
		t.Fatalf("esp32s3 triple 异常：%v", b.Triple)
	}
	if b.Domains["psram"] != "external-ram" {
		t.Fatalf("esp32s3 域异常：%v", b.Domains)
	}
}

func TestHostBoardHasNoTriple(t *testing.T) {
	b, ok := Find("host")
	if !ok {
		t.Fatal("host 未登记")
	}
	if b.Triple != nil {
		t.Fatalf("host 的 triple 应为 null（跟宿主走）：%v", *b.Triple)
	}
}

func TestVersionBandAndKinds(t *testing.T) {
	lo, hi := VersionBand()
	if lo == "" || hi == "" {
		t.Fatal("版本带为空")
	}
	kinds := DomainKinds()
	if len(kinds) == 0 {
		t.Fatal("domain kind 为空")
	}
}
