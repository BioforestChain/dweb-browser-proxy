package consts

const (
	// 系统：1001 ~ 1029
	InitRedisErr          = 1001
	JwtGenerateFailed     = 1002
	TokenCannotBeNull     = 1003
	TokenIsInvalid        = 1004
	TimeStampCannotBeNull = 1006
	SignCannotBeNull      = 1007
	SignValidError        = 1008
	TimeStampExpired      = 1009

	请求过快 = 1009

	// 参数：1030 ~ 1049
	ParamsError         = 1030
	ParamsLenLimitError = 1031
	ParamsStatusError   = 1032

	// 数据：1050 ~ 1069
	DataNotExists         = 1050
	ActivityDataNotExists = 1060
	DataExists            = 1051

	// 名称：1070 ~ 1089
	NameError = 1070

	// 其他
	TokenHasExpired = 3000
	SqlError        = 3001
	SystemBusy      = 3002
	RedisWriteError = 3003

	ReqParamErr = 521 + iota
)

var ErrorMessageList = func() map[int32]string {
	return map[int32]string{
		InitRedisErr:          "redis初始化错误",
		JwtGenerateFailed:     "token生成失败",
		TokenCannotBeNull:     "token不能为空",
		TokenIsInvalid:        "token是无效的",
		TimeStampCannotBeNull: "时间戳不能为空",
		SignCannotBeNull:      "签名不能为空",
		SignValidError:        "签名校验不正确",
		TimeStampExpired:      "链接已经失效",
		ParamsError:           "缺少参数",
		ParamsLenLimitError:   "参数长度超限",
		ParamsStatusError:     "状态异常",
		DataNotExists:         "数据不存在",
		DataExists:            "数据已存在",
		NameError:             "名称格式格式",
		TokenHasExpired:       "token已过期",
		SqlError:              "数据库异常",
		SystemBusy:            "系统繁忙",
		RedisWriteError:       "redis写入异常",
		ReqParamErr:           "参数错误",
	}
}
