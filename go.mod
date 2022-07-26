module github.com/username/blogclient

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.45.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/decred/dcrd/bech32 v1.1.2
	github.com/gin-gonic/gin v1.8.1
	github.com/ignite-hq/cli v0.20.3
	github.com/tendermint/tendermint v0.34.19
	github.com/username/blog v0.0.0-00010101000000-000000000000
)

replace github.com/username/blog => ../blog

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
