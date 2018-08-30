package web

import (
	"github.com/json-iterator/go"
	"io/ioutil"
)

type ContextUtil struct {
	ctx *Context
}

func NewContextUtil(ctx *Context) *ContextUtil {
	return &ContextUtil{ctx}
}

func (u *ContextUtil) ParseBodyJson(v interface{}) error {
	b, err := ioutil.ReadAll(u.ctx.Req.Body)
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal(b, v)
}

func (u *ContextUtil) ParseBodyJsonIterator() (jsoniter.Any, error) {
	b, err := ioutil.ReadAll(u.ctx.Req.Body)
	if err != nil {
		return nil, err
	}
	return jsoniter.Get(b), err
}

func (u *ContextUtil) ResponseJson(v interface{}) error {
	b, err := jsoniter.Marshal(v)
	if err != nil {
		return err
	}
	u.ctx.Res.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = u.ctx.Res.Write(b)
	return err
}
