package consts

const (
	Success = 0
	//
	DatabaseOperationError = 52
	RequestErr             = 400

	// 系统：1001 ~ 1029
	InitRedisErr          = 1001
	JwtGenerateFailed     = 1002
	TokenCannotBeNull     = 1003
	TokenIsInvalid        = 1004
	TimeStampCannotBeNull = 1006
	SignCannotBeNull      = 1007
	SignValidError        = 1008
	TimeStampExpired      = 1009
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
	// 服务器：3500 ~
	ServiceIsUnavailable = 3500
	ClientIpcSendErr     = 3501
	ReqParamErr          = 521 + iota

	SystemErr = 99999
)

var ErrorMessageList = func() map[int]string {
	return map[int]string{
		Success:                "Success",
		DatabaseOperationError: "Database Operation Error!!!",

		InitRedisErr:          "Redis initialization error",
		JwtGenerateFailed:     "Token generated failure",
		TokenCannotBeNull:     "Token cannot be empty",
		TokenIsInvalid:        "The token is invalid",
		TimeStampCannotBeNull: "Time stamp cannot be empty",
		SignCannotBeNull:      "The signature cannot be empty",
		SignValidError:        "The signature verification is incorrect",
		TimeStampExpired:      "The link has expired",
		ParamsError:           "Missing parameters",
		ParamsLenLimitError:   "The parameter length exceeds the limit",
		ParamsStatusError:     "State abnormal",
		DataNotExists:         "Data is not exist",
		DataExists:            "Data is exist",
		NameError:             "Name format error",
		TokenHasExpired:       "Token has expired",
		SqlError:              "Database exception",
		SystemBusy:            "The system is busy",
		RedisWriteError:       "Redis writes abnormal",
		ReqParamErr:           "The parameter error",
		ServiceIsUnavailable:  "The service is unavailable!",
		ClientIpcSendErr:      "IPC sending exception",
		SystemErr:             "System exception",
		RequestErr:            "Request errors",
	}
}
