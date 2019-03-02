module github.com/dessant/buster-client/cmd/client

require (
	bou.ke/monkey v1.0.1 // indirect
	buster-client/utils v0.0.0
	github.com/dessant/nativemessaging v0.0.0-20161221035708-f4769a80e040
	github.com/go-vgo/robotgo v0.0.0-20190208124536-364f19217850
	github.com/otiai10/mint v1.2.1 // indirect
	github.com/stretchr/testify v1.3.0 // indirect
	golang.org/x/net v0.0.0-20190206173232-65e2d4e15006 // indirect
)

replace buster-client/utils => ../../lib/utils
