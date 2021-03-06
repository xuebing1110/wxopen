package web

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bingbaba/hhsecret"
	"github.com/kataras/iris/context"
)

var (
	ERROR_USER_NOTALLOW  = errors.New("user not allow")
	ERROR_USER_NOTLOGIN  = errors.New("user not login")
	ERROR_REQUIRE_PASSWD = errors.New("require password to login")
)

func UserWhiteList(ctx context.Context) {
	username := ctx.Params().Get("username")
	for _, un_tmp := range DefaultCfg.UserWhiteList {
		if username == un_tmp {
			ctx.Next()
			return
		}
	}

	ctx.StatusCode(401)
	ctx.JSON(NewResponseWithErr(ERROR_USER_NOTALLOW, nil))
}

func UserLoginCheckHander(ctx context.Context) {
	var result = make(map[string]string)
	username := ctx.Params().Get("username")
	client, found := GetClientByUser(username)
	if !found {
		ctx.JSON(NewResponseWithErr(ERROR_USER_NOTLOGIN, result))
	} else {
		ctx.JSON(NewResponse(client.LoginData))
	}
}

func UserLoginHander(ctx context.Context) {
	var err error
	var result interface{}
	username := ctx.Params().Get("username")
	defer func() {
		if err != nil {
			logger.Errorf("%s login failed:", username, err)
		}
		ctx.JSON(NewResponseWithErr(err, result))
	}()

	userInfo := make(map[string]string)
	err = ctx.ReadJSON(&userInfo)
	if err != nil {
		body, err2 := ioutil.ReadAll(ctx.Request().Body)
		if err2 != nil {
			err = err2
			return
		}
		if len(body) != 0 {
			return
		}
	}

	var found bool
	var client *hhsecret.Client
	password, found := userInfo["password"]
	if !found {
		client, found = GetClientByUser(username)
		if !found {
			err = ERROR_REQUIRE_PASSWD
			return
		}
	} else {
		client = hhsecret.NewClient(username, password, DefaultCfg.ConsumerKey, DefaultCfg.ConsumerSecret)
		err = client.Login()
		if err != nil {
			return
		}
	}
	result = client.LoginData
	SaveClient(username, client)

	return
}

func UserSignHander(ctx context.Context) {
	var err error
	var result interface{}
	username := ctx.Params().Get("username")
	defer func() {
		if err != nil {
			logger.Errorf("%s sign failed:", username, err)
		}
		ctx.JSON(NewResponseWithErr(err, result))
	}()

	client, found := GetClientByUser(username)
	if !found {
		err = ERROR_USER_NOTLOGIN
		return
	}

	result, err = client.Sign()
	return
}
func UserListSignHander(ctx context.Context) {
	var err error
	var result interface{}
	username := ctx.Params().Get("username")
	defer func() {
		ctx.JSON(NewResponseWithErr(err, result))
	}()

	client, found := GetClientByUser(username)
	if !found {
		err = ERROR_USER_NOTLOGIN
		return
	}

	result, err = client.ListSignPost()
	return
}

func UserMonthSignHandler(ctx context.Context) {
	username := ctx.Params().Get("username")
	year := ctx.Params().Get("year")
	month := ctx.Params().Get("month")

	var ms *hhsecret.MonthSign
	var err error
	defer func() {
		ctx.JSON(NewResponseWithErr(err, ms))
	}()
	ms, err = hhsecret.GetMonthSign(username, year, month)
	return
}

func NoticeHander(ctx context.Context) {
	username := ctx.Params().Get("username")

	var notice = false
	var afternoon = false
	if time.Now().Hour() >= 12 {
		afternoon = true

		// not send notice before 17:00
		if time.Now().Hour() < 17 {
			notice = false
			ctx.JSON(NewResponse(notice))
			return
		}
	}

	client, found := GetClientByUser(username)
	if !found {
		ctx.JSON(NewResponseWithErr(errors.New("user not login"), false))
		return
	}

	lsd, err := client.ListSignPost()
	if err != nil {
		ctx.JSON(NewResponseWithErr(err, nil))
	} else {
		if len(lsd.Signs) == 0 {
			notice = true
		} else {
			if afternoon {
				mtime := lsd.Signs[0].GetMinuteSecode()
				if strings.Compare(mtime, "17:30") < 0 {
					notice = true
				}
			}
		}

		ctx.JSON(NewResponse(notice))
	}
}
