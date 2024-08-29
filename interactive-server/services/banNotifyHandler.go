package services

import (
	"interactive-server/logger"
	"net/http"
	"strings"
	"time"
)

func BanNotifyHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("banNotifyHandler error:%v", err)
			return
		}
	}()
	ut := strings.Split(r.URL.Path, "/")
	if len(ut) < 5 || len(ut[2]) == 0 || len(ut[3]) == 0 || len(ut[4]) == 0 {
		logger.Infof("req path:%s", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("req path eg: /banNotify/{userid}/{closetype}/{0/1}"))
		return
	}
	closeType := ut[3]
	var ci chBanInfo
	ci.UserId = ut[2]
	var ban bool
	if ut[4] == "1" {
		ban = true
	}
	if closeType == "0" {
		ci.Reason = "用户被禁言"
		if !ban {
			ci.Reason = "用户解除禁言"
		}
		vch, exist := GameSvcBanCtrlr.Load(ci.UserId)
		if !exist {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no chat ws conn found with user:" + ci.UserId))
			return
		}
		select {
		case vch.(chan chBanInfo) <- chBanInfo{
			Reason:  ci.Reason,
			UserId:  ci.UserId,
			Ban:     ban,
			BanType: closeType,
		}:
		case <-time.After(time.Second):
		}
	} else if closeType == "1" {
		ci.Reason = "用户被禁用"
		if !ban {
			ci.Reason = "用户解除禁用"
		}
		vch, exist := GameSvcBanCtrlr.Load(ci.UserId)
		if !exist {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no chat ws conn found with user:" + ci.UserId))
			return
		}
		select {
		case vch.(chan chBanInfo) <- chBanInfo{
			Reason:  ci.Reason,
			UserId:  ci.UserId,
			Ban:     ban,
			BanType: closeType,
		}:
		case <-time.After(time.Second):
		}
	}
}
