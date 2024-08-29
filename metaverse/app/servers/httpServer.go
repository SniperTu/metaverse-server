/**
    * @Description:
    * @Author: daniel(2023/2/21)
**/
package server

import (
	"fmt"
	"log"
	"metaverse/conf"
	"metaverse/logger"
	"metaverse/models"
	"metaverse/pbs"
	"metaverse/utils"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func RunHTTPServer() {
	mux := http.NewServeMux()
	uurl, _ := url.Parse(fmt.Sprintf("http://localhost:%d/", utils.PProf()))
	mux.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
		httputil.NewSingleHostReverseProxy(uurl).ServeHTTP(w, r)
	})
	mux.HandleFunc("/UsageDataSummaryExport", UsageDataSummaryExport)
	srv := &http.Server{
		Addr:    conf.Conf.HTTPPort,
		Handler: mux,
	}
	ln, err := net.Listen("tcp", conf.Conf.HTTPPort)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Printf("http server listening on port%s\n", conf.Conf.HTTPPort)
		if err = srv.Serve(ln); err != nil {
			log.Fatal(err)
		}
	}()
}
func UsageDataSummaryExport(w http.ResponseWriter, r *http.Request) {
	var err error
	var loginUser models.User
	loginUser, err = models.GetTokenCache(r.URL.Query().Get("token"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var schoolName string
	if loginUser.SchoolInfo != nil && loginUser.UserType != pbs.UserType_superAdmin {
		schoolName = loginUser.SchoolInfo.Name
	}
	fileName := schoolName + "使用数据统计.xlsx"
	fi, er := os.Lstat(fileName)
	if er == nil && fi.ModTime().Add(5*time.Minute).After(time.Now()) {
		http.ServeFile(w, r, fileName)
		return
	}
	var rows []*pbs.User
	rows, _, _, err = new(models.User).List("", "", schoolName, 0, -1, loginUser.UserType, true, true, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("获取用户信息列表失败"))
		return
	}
	f := excelize.NewFile()
	var idx int
	idx, err = f.GetSheetIndex("Sheet1")
	if err != nil {
		logger.Errorf("%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("生成excel文件失败"))
		return
	}
	f.SetSheetName("Sheet1", "使用数据统计")
	f.SetActiveSheet(idx)
	f.SetSheetRow("使用数据统计", "A1", &[]interface{}{"角色名称", "手机号", "最近登录时间", "累积登录次数", "累计登录时长(分钟)"})
	var lastLoginTime string
	var roleName string

	for i, r := range rows {
		roleName = ""
		if r.LastLogin == 0 {
			lastLoginTime = ""
		} else {
			lastLoginTime = time.Unix(r.LastLogin, 0).Format("2006.01.02 15:04:05")
		}

		if len(r.Roles) != 0 {
			rns := []string{}
			for _, role := range r.Roles {
				rns = append(rns, role.Name)
			}
			roleName = strings.Join(rns, ",")
		}
		f.SetSheetRow("使用数据统计", "A"+strconv.Itoa(i+2), &[]interface{}{roleName, r.Mobile,
			lastLoginTime, r.LoginTimes, r.LoginDurationMin})
	}
	f.SaveAs(fileName)
	if err = f.Close(); err != nil {
		logger.Errorf("%v", err)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	http.ServeFile(w, r, fileName)
	return
}
