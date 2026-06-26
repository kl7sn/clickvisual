package view

import (
	"testing"

	"github.com/clickvisual/clickvisual/api/internal/pkg/utils/mapping"
)

func TestReqStorageCreateJSONEachRowMappingsExtractJSONStringChildren(t *testing.T) {
	req := ReqStorageCreate{
		RawLogField: "content",
		TimeField:   "time",
		SourceMapping: mapping.List{Data: []mapping.Item{
			{Key: "content", Typ: "String"},
			{Key: "count", Typ: "String", Parent: "content", FromJSONString: true},
			{Key: "log_time", Typ: "String", Parent: "content", FromJSONString: true},
			{Key: "time", Typ: "Float64"},
			{Key: "host.name", Typ: "String"},
		}},
	}

	if got, want := req.Mapping2StringJSONEachRowStream(true), "`host.name` String,"; got != want {
		t.Fatalf("stream mapping = %q, want %q", got, want)
	}

	if got, want := req.Mapping2StringJSONEachRowData(true), "`count` String,\n`log_time` String,\n`host.name` String,"; got != want {
		t.Fatalf("data mapping = %q, want %q", got, want)
	}

	if got, want := req.Mapping2StringJSONEachRowView(), "JSONExtractString(`content`, 'count') AS `count`,\nJSONExtractString(`content`, 'log_time') AS `log_time`,\n`host.name`,"; got != want {
		t.Fatalf("view mapping = %q, want %q", got, want)
	}
}
