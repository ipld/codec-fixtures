package codec_fixtures

type FixtureName = string
type FixtureBlacklistReason = string

var FixtureBlacklist = map[FixtureName]FixtureBlacklistReason{
	// negative integer outside of the int64 range, no support for that in Go
	"int--11959030306112471732": "integer out of int64 range",
	// dag-json strconv parsing error, out of int64 range
	"int-11959030306112471731": "integer out of int64 range",
	// dag-json strconv parsing error, out of int64 range
	"int-18446744073709551615": "integer out of int64 range",
	// dag-pb codec not strict on encoding unordered named link lists, should
	// error but does not for these two:
	"dag-pb/encode/bad sort":               "pre-sort not required",
	"dag-pb/encode/bad sort (incl length)": "pre-sort not required",
}
