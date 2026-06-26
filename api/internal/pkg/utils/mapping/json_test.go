package mapping

import (
	"reflect"
	"testing"
)

func Test_mapping(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want List
	}{
		// TODO: Add test cases.
		{
			name: "test-1",
			args: args{
				input: `{"Followers":8900}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "Followers",
						Typ: "Float64",
					},
				},
			},
		},
		{
			name: "test-2",
			args: args{
				input: `{"Name":"gopher"}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "Name",
						Typ: "String",
					},
				},
			},
		},
		{
			name: "test-3",
			args: args{
				input: `{"IsAdmin":false}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "IsAdmin",
						Typ: "Bool",
					},
				},
			},
		},
		{
			name: "test-4",
			args: args{
				input: `{"tags": [
        {
            "key": "otel.library.name",
            "vStr": "enter_file"
        },
        {
            "key": "http.client_ip",
            "vStr": "219.233.199.199"
        }
]}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "tags",
						Typ: "Array(JSON)",
					},
				},
			},
		},
		{
			name: "test-5",
			args: args{
				input: `{    "process": {
        "serviceName": "frontend",
        "tags": [
            {
                "key": "telemetry.sdk.language",
                "vStr": "webjs"
            },
            {
                "key": "telemetry.sdk.name",
                "vStr": "opentelemetry"
            },
            {
                "key": "telemetry.sdk.version",
                "vStr": "1.8.0"
            }
        ]
    }}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "process",
						Typ: "JSON",
					},
					{
						Key:    "serviceName",
						Typ:    "String",
						Parent: "process",
					},
					{
						Key:    "tags",
						Typ:    "Array(JSON)",
						Parent: "process",
					},
				},
			},
		},
		{
			name: "json string children",
			args: args{
				input: `{"content":"{\"count\":\"19\",\"log_time\":\"2025-05-27T11:28:15.568171944+08:00\"}","time":1748316495}`,
			},
			want: List{
				Data: []Item{
					{
						Key: "content",
						Typ: "String",
					},
					{
						Key:            "count",
						Typ:            "String",
						Parent:         "content",
						FromJSONString: true,
					},
					{
						Key:            "log_time",
						Typ:            "String",
						Parent:         "content",
						FromJSONString: true,
					},
					{
						Key: "time",
						Typ: "Float64",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := Handle(tt.args.input, true); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapping() = %v, want %v", got, tt.want)
			}
		})
	}
}
