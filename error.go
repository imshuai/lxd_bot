package main

import "errors"

var (
	ErrorKeyNotFound              = errors.New("key not found")
	ErrorCreateUserFailed         = errors.New("发生错误，创建新用户失败")
	ErrorOverQuota                = errors.New("超过可用实例数量")
	ErrorCheckinFailed            = errors.New("签到失败，请稍后重试")
	ErrorUnmarshalInstance        = errors.New("解析实例数据失败")
	ErrorBucketNameBlankOrTooLong = errors.New("bucket name blank or too long")
)
