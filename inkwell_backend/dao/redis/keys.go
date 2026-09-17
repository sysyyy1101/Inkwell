package redis

/*
	Redis Key
*/

const (
	KeyPostInfoHashPrefix = "inkwell:post:"
	KeyPostTimeZSet       = "inkwell:post:time"
	KeyPostScoreZSet      = "inkwell:post:score"
	//KeyPostVotedUpSetPrefix   = "inkwell:post:voted:down:"
	//KeyPostVotedDownSetPrefix = "inkwell:post:voted:up:"
	KeyPostVotedZSetPrefix = "inkwell:post:voted:"

	KeyCommunityPostSetPrefix = "inkwell:community:"
)
