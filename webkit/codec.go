package webkit

import (
	"encoding/json"
	"net/http"

	kerr "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/pkg/errors"
)

const (
	dataReplace = "@DATA"

	// Error code to HTTP status mappings
	codeServerError   = 500
	codeForbidden     = 403
	codeRateLimit     = 429
	codeRateLimitAlt  = 4300
	codeRedirect      = 3002
	codeBadRequest    = 4000
	codeUnauthorized  = 4200
)

type JsonReply struct {
	RetCode int32       `json:"code"`
	Message string      `json:"msg,omitempty"`
	Detail  string      `json:"detail,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func ErrorEncoder(errorReasonValue map[string]int32) khttp.EncodeErrorFunc {
	errorReasonValue["CODEC"] = 400
	errorReasonValue["VALIDATOR"] = 400
	return func(w http.ResponseWriter, r *http.Request, err error) {
		// 尝试从pkg/errors取原始的err，避免向外输出调用栈信息
		err = errors.Cause(err)
		parsedErr := kerr.FromError(err)
		retCode := errorReasonValue[parsedErr.Reason]
		message := parsedErr.Message
		if retCode == 0 || retCode == codeServerError {
			message = "Oops, something went wrong."
		}
		reply := &JsonReply{
			RetCode: retCode,
			Message: message,
		}
		codec, _ := khttp.CodecForRequest(r, "Accept")
		body, err := json.Marshal(reply)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch retCode {
		case codeForbidden:
			http.Error(w, string(body), http.StatusForbidden)
			return
		case codeRateLimit, codeRateLimitAlt:
			http.Error(w, string(body), http.StatusTooManyRequests)
			return
		case codeRedirect:
			http.Redirect(w, r, reply.Message, http.StatusFound)
			return
		case codeBadRequest:
			http.Error(w, string(body), http.StatusBadRequest)
			return
		case codeUnauthorized:
			http.Error(w, string(body), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/"+codec.Name())
		_, err = w.Write(body)
		if err != nil {
			log.Error("error encoding failed", err)
			return
		}
	}
}

func ResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	reply := &JsonReply{
		RetCode: 0,
		Message: "Success",
		Data:    v,
	}

	codec, _ := khttp.CodecForRequest(r, "Accept")
	b, err := json.Marshal(reply)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/"+codec.Name())
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return nil
}
