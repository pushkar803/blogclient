module github.com/username/blogclient

go 1.16

require (
	github.com/ignite-hq/cli v0.20.3
	github.com/username/blog v0.0.0-00010101000000-000000000000
)

replace github.com/username/blog => ../blog

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
