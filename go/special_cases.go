package codec_fixtures

type FixtureName = string
type FixtureBlacklistReason = string

var FixtureBlacklist = map[FixtureName]FixtureBlacklistReason{
	"int--11959030306112471732": "integer out of int64 range",
	"int-11959030306112471731":  "integer out of int64 range",
	"int-18446744073709551615":  "integer out of int64 range",
}
