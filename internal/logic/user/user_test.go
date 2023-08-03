/*
*

	@author: bnqkl
	@since: 2023/6/8/008 16:40
	@desc: //TODO

*
*/
package user

import "testing"

func Test_sUser_GenerateProxyServerCVisitorIni(t *testing.T) {
	type args struct {
		visitorConfig ProxyServerCVisitor
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"had_set_uid",
			args{
				ProxyServerCVisitor{
					"md5Identification_visitor",
					"visitor",
					"md5Identification",
					"127.0.0.1",
					1999,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sUser{}
			if err := s.GenerateProxyServerCVisitorIni(tt.args.visitorConfig); (err != nil) != tt.wantErr {
				t.Errorf("GenerateProxyServerCVisitorIni() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
