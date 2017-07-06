package errcode

import (
	"fmt"
	"strconv"
	"strings"
)

type ErrCode int

const (
	//需要返回http500的错误码
	OKCode            ErrCode = 0
	InternalErrorCode ErrCode = 1
	DbErrCode         ErrCode = 2
	CacheErrCode      ErrCode = 3
	DecodeErrCode     ErrCode = 4
	EncodeErrCode     ErrCode = 5
	RPCErrCode        ErrCode = 6
	HttpErrCode       ErrCode = 7
	//直接抛给用户的错误码
	//userprofile
	UserNotFoundCode  ErrCode = 100
	EmailNotFoundCode ErrCode = 101
	EmailRepeatCode   ErrCode = 102
	EmailWrongCode    ErrCode = 103
	UserDisableCode   ErrCode = 104
	PhoneRepeatCode   ErrCode = 105
	PhoneWrongCode    ErrCode = 106
	PhoneNotExist     ErrCode = 107
	PasswordWrong     ErrCode = 108
	VerifyCodeWrong   ErrCode = 109
	MaxTimeToday      ErrCode = 110
	PasswordSame      ErrCode = 111
	//gps
	RouteIdNotFoundCode ErrCode = 201
	RouteDuplicateCode  ErrCode = 202
	//feed
	FeedNotFoundCode    ErrCode = 301
	CommentNotFoundCode ErrCode = 302
	PraiseNotFoundCode  ErrCode = 303
	PraiseReplicateCode ErrCode = 304
	BeBlock             ErrCode = 305
	//geo
	PointWrongCode ErrCode = 401
	//global
	ParameterErrCode ErrCode = 501 //userid 为0 或userid为自己
	AuthorityErrCode ErrCode = 502 //无权限
	HighFrequency    ErrCode = 503

	MaxUserError ErrCode = 9999
)

type InternalError struct {
	Code       ErrCode
	Error_info error
}

//组装带errorcode的err
func NewInternalError(code ErrCode, err error) error {
	if code < UserNotFoundCode && code > ErrCode(0) {
		//CheckError(&InternalError{Code: ErrCode(code), Error_info: err})
	}
	//panic(err)
	return &InternalError{Code: ErrCode(code), Error_info: err}
}

//组装带errorcode的err
func NewInternalErrorByStr(code ErrCode, err string) error {
	if code < UserNotFoundCode && code > ErrCode(0) {
		//CheckError(&InternalError{Code: ErrCode(code), Error_info: fmt.Errorf(err)})
	}
	//panic(err)
	return &InternalError{Code: ErrCode(code), Error_info: fmt.Errorf(err)}
}

func (err *InternalError) Error() string {
	return fmt.Sprintf("%d:%s", err.Code, err.Error_info)
}

//判断是否系统错误 并返回错误码以及错误描述
func IsUserErr(original_err error) (bool, int, string) {
	infos := strings.Split(original_err.Error(), ":")
	if len(infos) < 2 {
		return false, 0, infos[0]
	}
	code, err := strconv.Atoi(infos[0])
	if nil != err {
		return false, 0, infos[1]
	}
	if code >= int(UserNotFoundCode) && code <= int(MaxUserError) {
		return true, code, infos[1]
	}
	return false, 0, infos[1]
}

func GetErrInfo(original_err error) string {
	var code int
	var info string
	fmt.Sscanf(original_err.Error(), "%d:%s", &code, &info)
	return info
}

/*
type UserError struct {
	Code       ErrCode
	Error_info string
}

type DbError struct {
	Code       ErrCode
	Error_info error
}

type CacheError struct {
	Code       ErrCode
	Error_info error
}

type DecodeErr struct {
	Code       ErrCode
	Error_info error
}

type EncodeErr struct {
	Code       ErrCode
	Error_info error
}

type UserNotFoundError struct {
	Code    ErrCode
	User_id int64
}

type EmailNotFoundError struct {
	Code  ErrCode
	Email string
}

type EmailRepeatError struct {
	Code  ErrCode
	Email string
}

type EmailWrongError struct {
	Code  ErrCode
	Email string
}

type PhoneRepeatError struct {
	Code  ErrCode
	Phone string
}

type PhoneWrongError struct {
	Code  ErrCode
	Phone string
}

type PhoneNotExistError struct {
	Code  ErrCode
	Phone string
}

type PasswordWrongError struct {
	Code  ErrCode
	Phone string
}

type UserDisableError struct {
	Code       ErrCode
	Error_info string
}

type RouteIdNotFoundError struct {
	Code    ErrCode
	RouteId int64
}

type RouteDuplicateError struct {
	Code       ErrCode
	Error_info string
}

type FeedNotFoundError struct {
	Code   ErrCode
	FeedId int64
}

type CommentNotFoundError struct {
	Code      ErrCode
	CommentId int64
}

type PraiseNotFoundError struct {
	Code     ErrCode
	PraiseId int64
}

type PraiseReplicateError struct {
	Code ErrCode
	Info string
}

type PointError struct {
	Code       ErrCode
	Error_info error
}

type ParameterError struct {
	Code       ErrCode
	Error_info string
}

type AuthorityError struct {
	Code       ErrCode
	Error_info string
}

func (err *UserError) Error() string {
	return fmt.Sprintf("%d:%s", err.Code, err.Error_info)
}

func (err *DbError) Error() string {
	return fmt.Sprintf("%d:database error:%s", err.Code, err.Error_info)
}

func (err *CacheError) Error() string {
	return fmt.Sprintf("%d:cache error:%s", err.Code, err.Error_info)
}

func (err *DecodeErr) Error() string {
	return fmt.Sprintf("%d:decode error:%s", err.Code, err.Error_info)
}

func (err *EncodeErr) Error() string {
	return fmt.Sprintf("%d:encodeErr error:%s", err.Code, err.Error_info)
}

func (err *UserNotFoundError) Error() string {
	return fmt.Sprintf("%d:user:%d not found", err.Code, err.User_id)
}

func (err *EmailNotFoundError) Error() string {
	return fmt.Sprintf("%d:email:%s not found", err.Code, err.Email)
}

func (err *EmailRepeatError) Error() string {
	return fmt.Sprintf("%d:email:%s repeated in database", err.Code, err.Email)
}

func (err *EmailWrongError) Error() string {
	return fmt.Sprintf("%d:email:%s type wrong", err.Code, err.Email)
}

func (err *PhoneRepeatError) Error() string {
	return fmt.Sprintf("%d:phone:%s repeated in database", err.Code, err.Phone)
}

func (err *PhoneWrongError) Error() string {
	return fmt.Sprintf("%d:phone:%s type wrong", err.Code, err.Phone)
}

func (err *PhoneNotExistError) Error() string {
	return fmt.Sprintf("%d:phone:%s not exist", err.Code, err.Phone)
}

func (err *PasswordWrongError) Error() string {
	return fmt.Sprintf("%d:Wrong password", err.Code)
}

func (err *RouteIdNotFoundError) Error() string {
	return fmt.Sprintf("%d:rout_id:%d not found", err.Code, err.RouteId)
}

func (err *RouteDuplicateError) Error() string {
	return fmt.Sprintf("%d:inter gps route failed:%s", err.Code, err.Error_info)
}

func (err *FeedNotFoundError) Error() string {
	return fmt.Sprintf("%d:feed_id:%d not found", err.Code, err.FeedId)
}

func (err *CommentNotFoundError) Error() string {
	return fmt.Sprintf("%d:feed_id:%d not found", err.Code, err.CommentId)
}

func (err *PraiseNotFoundError) Error() string {
	return fmt.Sprintf("%d:feed_id:%d not found", err.Code, err.PraiseId)
}

func (err *PraiseReplicateError) Error() string {
	return fmt.Sprintf("%d::%s", err.Code, err.Info)
}

func (err *PointError) Error() string {
	return fmt.Sprintf("%d:points:%s wrong", err.Code, err.Error_info)
}

func (err *ParameterError) Error() string {
	return fmt.Sprintf("%d:parameter err:%s ", err.Code, err.Error_info)
}

func (err *AuthorityError) Error() string {
	return fmt.Sprintf("%d:authority err:%s ", err.Code, err.Error_info)
}

func (err *UserDisableError) Error() string {
	return fmt.Sprintf("%d:LoginError:%s", err.Code, err.Error_info)
}
*/
