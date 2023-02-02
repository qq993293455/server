package utils

var (
	InvalidArguments = &CodeInfo{
		Code: 1001,
		Msg:  "invalid arguments",
	}
	IllegalRequest = &CodeInfo{
		Code: 1002,
		Msg:  "illegal request",
	}
	InvalidSign = &CodeInfo{
		Code: 1002,
		Msg:  "invalid sign",
	}
	RequiredSign = &CodeInfo{
		Code: 1003,
		Msg:  `arguments "sign" is required`,
	}
	RequiredTime = &CodeInfo{
		Code: 1004,
		Msg:  `arguments "time" is required`,
	}
	DuplicateSN = &CodeInfo{
		Code: 1100,
		Msg:  "duplicate sn",
	}
	PlayerNotExist = &CodeInfo{
		Code: 1100,
		Msg:  "player not exist",
	}
)
