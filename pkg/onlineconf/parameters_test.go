package onlineconf

import "testing"

func TestMakePath(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty path",
			args: args{s: []string{}},
			want: "",
		},
		{
			name: "Single path element",
			args: args{s: []string{"foo"}},
			want: "/foo",
		},
		{
			name: "Multiple path elements",
			args: args{s: []string{"foo", "bar", "baz"}},
			want: "/foo/bar/baz",
		},
		// Add more test cases as needed
		{
			name: "Path with empty elements",
			args: args{s: []string{"", "foo", "", "bar"}},
			want: "//foo//bar",
		},
		{
			name: "Path with special characters",
			args: args{s: []string{"foo.bar", "baz/qux"}},
			want: "/foo.bar/baz/qux",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakePath(tt.args.s...); got != tt.want {
				t.Errorf("MakePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
