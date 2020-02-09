module buster-client/cmd/setup

go 1.13

replace buster-client/utils => ../../lib/utils

require (
	buster-client/utils v0.0.0-00010101000000-000000000000
	github.com/dessant/open-golang v0.0.0-20190104022628-a2dfa6d0dab6
	github.com/gofrs/uuid v3.2.0+incompatible
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5
)
