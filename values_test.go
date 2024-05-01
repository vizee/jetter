package main

import (
	"reflect"
	"testing"
)

func Test_allKeyFields(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want []keyField
	}{
		{args: args{""}, want: nil},
		{args: args{"."}, want: []keyField{{"", false}}},
		{args: args{"[]"}, want: []keyField{{"", true}}},
		{args: args{"a"}, want: []keyField{{"a", false}}},
		{args: args{".a"}, want: []keyField{{"a", false}}},
		{args: args{"[a]"}, want: []keyField{{"a", true}}},
		{args: args{"a.b[c][d].e"}, want: []keyField{{"a", false}, {"b", false}, {"c", true}, {"d", true}, {"e", false}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allKeyFields(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("allKeyFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applyAssign(t *testing.T) {
	type args struct {
		values any
		fields []keyField
		value  string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			args: args{
				values: nil,
				fields: nil,
				value:  "1",
			},
			want:    "1",
			wantErr: false,
		},
		{
			args: args{
				values: nil,
				fields: []keyField{
					{name: "a"},
					{name: "1", index: true},
				},
				value: "1",
			},
			want: map[string]any{
				"a": []any{
					1: "1",
				},
			},
			wantErr: false,
		},
		{
			args: args{
				values: []any{"0"},
				fields: []keyField{
					{name: "1", index: true},
					{name: "a"},
				},
				value: "1",
			},
			want: []any{
				0: "0",
				1: map[string]any{
					"a": "1",
				},
			},
			wantErr: false,
		},
		{
			args: args{
				values: []any{"0"},
				fields: []keyField{
					{name: "a"},
				},
				value: "1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			args: args{
				values: map[string]any{"a": "0"},
				fields: []keyField{
					{name: "1", index: true},
				},
				value: "1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			args: args{
				values: []any{[]any{"a"}},
				fields: []keyField{
					{name: "0", index: true},
					{name: "a", index: true},
				},
				value: "1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			args: args{
				values: map[string]any{"a": "0"},
				fields: []keyField{
					{name: "a"},
					{name: "0", index: true},
				},
				value: "1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyAssign(tt.args.values, tt.args.fields, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyAssign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyAssign() = %v, want %v", got, tt.want)
			}
		})
	}
}
