module buster-client/cmd/client

go 1.13

replace buster-client/pkg/input => ../../pkg/input

replace buster-client/utils => ../../lib/utils

require (
	buster-client/pkg/input v0.0.0-00010101000000-000000000000
	buster-client/utils v0.0.0-00010101000000-000000000000
	github.com/dessant/nativemessaging v0.0.0-20161221035708-f4769a80e040
)
