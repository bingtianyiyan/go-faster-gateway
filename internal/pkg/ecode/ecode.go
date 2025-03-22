package ecode

var (
	UpstreamNotInit        = New(1000, 0, "balancer not init", "")
	InternalServerErrorErr = New(1001, 500, "Internal Server Error", "InternalServerError")
	BackendTimeoutErr      = New(1002, 504, "Backend timeout", "iot.apigw.BackendTimeout")
)
