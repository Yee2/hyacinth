package webui

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	port, err := strconv.Atoi(ps.ByName("port"))
	if err != nil {
		logger.Info("%s", err)
		http.NotFound(w, r)
		return
	}

	ss, ok := tables.Rows[port]
	if !ok {
		logger.Info("端口%d未找到", port)
		http.NotFound(w, r)
		return
	}

	encryption := strings.ToUpper(r.PostFormValue("encryption"))
	encryptionCheck := false
	for _, m := range shadowsocks_methods {
		if m == encryption {
			encryptionCheck = true
			break
		}
	}

	if !encryptionCheck {
		res_error(w, http.StatusBadRequest, "错误请求!")
		return
	}

	password := r.PostFormValue("password")
	if len(password) < 6 {
		res_error(w, http.StatusBadRequest, "密码太短，最少6位")
		return
	}

	if len(password) > 32 {
		res_error(w, http.StatusBadRequest, "密码太长，最多32位")
		return
	}
	ss.Cipher = encryption
	ss.Password = password

	go func() {
		if err := ss.Stop(); err != nil {
			logger.Warning("无法正常关闭服务:%s", err)
			return
		}
		time.Sleep(time.Second * 5) //
		if err := ss.Start(); err != nil {
			logger.Warning("无法正常启动服务:%s", err)
			return
		}
	}()

	data, _ := json.Marshal(struct {
		Port     int    `json:"port"`
		Password string `json:"password"`
		Method   string `json:"method"`
	}{ss.Port, ss.Password, ss.Cipher})
	w.Write(data)

}
