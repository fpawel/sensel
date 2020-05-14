package pkg

import "testing"

func TestFormatFloatTrimNulls(t *testing.T) {
	type args struct {
		v    float64
		prec int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				v: 500.,
			},
			want: "500",
		},
		{
			args: args{
				v:    500.567,
				prec: 4,
			},
			want: "500.567",
		},
		{
			args: args{
				v:    500.567777,
				prec: 3,
			},
			want: "500.568",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFloatTrimNulls(tt.args.v, tt.args.prec); got != tt.want {
				t.Errorf("FormatFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}
