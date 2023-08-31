package consts

const (
	JwtSecret    = "8ab6c8dee22768da1503351069f032cb" // jwt密匙
	JwtTTL       = 3 * 24                             // jwt过期时间
	CtxUserIDKey = "userID"
)

type User struct {
	UserID uint32 `db:"user_id"`
	Token  string
}
