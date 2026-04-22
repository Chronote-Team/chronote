package errs

import "testing"

func TestMapHTTPStatusAndMessage(t *testing.T) {
	cases := []struct {
		name           string
		err            error
		wantStatusCode int
		wantMessage    string
	}{
		{"validation", Validation("请求参数无效"), 400, "请求参数无效"},
		{"unauthorized", Unauthorized("Without Token, and you're unauthorized!"), 401, "Without Token, and you're unauthorized!"},
		{"forbidden", Forbidden("无权限访问该明信片"), 403, "无权限访问该明信片"},
		{"not found", NotFound("明信片不存在"), 404, "明信片不存在"},
		{"conflict", Conflict("username 已存在"), 409, "username 已存在"},
		{"internal", Internal("用户注册失败"), 500, "用户注册失败"},
		{"degraded", Degraded("Some services degraded"), 207, "Some services degraded"},
		{"unavailable", Unavailable("Service Unavailable"), 503, "Service Unavailable"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotMessage := MapHTTP(tc.err)
			if gotStatus != tc.wantStatusCode || gotMessage != tc.wantMessage {
				t.Fatalf("MapHTTP(%q) = (%d, %q), want (%d, %q)", tc.name, gotStatus, gotMessage, tc.wantStatusCode, tc.wantMessage)
			}
		})
	}
}
