package dbdata

import "testing"

func TestParseUserAgent(t *testing.T) {
	type args struct {
		userAgent string
	}
	type res struct {
		os_idx     uint8
		client_idx uint8
		ver        string
	}
	tests := []struct {
		name string
		args args
		want res
	}{
		{
			name: "mac os 1",
			args: args{userAgent: "cisco anyconnect vpn agent for mac os x 4.10.05085"},
			want: res{os_idx: 2, client_idx: 1, ver: "4.10.05085"},
		},
		{
			name: "mac os 2",
			args: args{userAgent: "anyconnect darwin_i386 4.10.05085"},
			want: res{os_idx: 2, client_idx: 1, ver: "4.10.05085"},
		},
		{
			name: "windows",
			args: args{userAgent: "cisco anyconnect vpn agent for windows 4.8.02042"},
			want: res{os_idx: 1, client_idx: 1, ver: "4.8.02042"},
		},
		{
			name: "iPad",
			args: args{userAgent: "anyconnect applesslvpn_darwin_arm (ipad) 4.10.04060"},
			want: res{os_idx: 5, client_idx: 1, ver: "4.10.04060"},
		},
		{
			name: "iPhone",
			args: args{userAgent: "cisco anyconnect vpn agent for apple iphone 4.10.04060"},
			want: res{os_idx: 5, client_idx: 1, ver: "4.10.04060"},
		},
		{
			name: "android",
			args: args{userAgent: "anyconnect android 4.10.05096"},
			want: res{os_idx: 4, client_idx: 1, ver: "4.10.05096"},
		},
		{
			name: "linux",
			args: args{userAgent: "cisco anyconnect vpn agent for linux v7.08"},
			want: res{os_idx: 3, client_idx: 1, ver: "7.08"},
		},
		{
			name: "openconnect",
			args: args{userAgent: "openconnect-gui 1.5.3 v7.08"},
			want: res{os_idx: 0, client_idx: 2, ver: "7.08"},
		},
		{
			name: "unknown",
			args: args{userAgent: "unknown 1.4.3 aabcd"},
			want: res{os_idx: 0, client_idx: 0, ver: ""},
		},
		{
			name: "unknown 2",
			args: args{userAgent: ""},
			want: res{os_idx: 0, client_idx: 0, ver: ""},
		},
		{
			name: "anylink",
			args: args{userAgent: "anylink vpn agent for linux v1.0"},
			want: res{os_idx: 3, client_idx: 3, ver: "1.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os_idx, client_idx, ver := UserActLogIns.ParseUserAgent(tt.args.userAgent); os_idx != tt.want.os_idx || client_idx != tt.want.client_idx || ver != tt.want.ver {
				t.Errorf("ParseUserAgent() = %v, %v, %v, want %v, %v, %v", os_idx, client_idx, ver, tt.want.os_idx, tt.want.client_idx, tt.want.ver)
			}
		})
	}
}
