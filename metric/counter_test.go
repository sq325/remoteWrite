package metric

import (
	"testing"
)

func TestPBCounter_GetValue(t *testing.T) {
	type fields struct {
		pbcounter *PBCounter
	}
	type args struct {
		lvs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				pbcounter: NewPBCounter("test", "test", []string{"label1", "label2"}),
			},
			args: args{
				lvs: []string{"1", "2"},
			},
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.fields.pbcounter
			c.Add(tt.args.lvs, tt.want)
			got, err := c.GetValue(tt.args.lvs)
			if (err != nil) != tt.wantErr {
				t.Errorf("PBCounter.GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PBCounter.GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
