package common

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func ResponseJson(w http.ResponseWriter, statusCode int, data JSON) {
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		msg := strings.Replace(err.Error(), "\"", "\\\"", -1)
		w.Write([]byte(fmt.Sprintf("{\"ok\":false,\"error\":\"JSON stringify failed: %s\"", msg)))
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func ResponseApiOk(log *logrus.Entry, w http.ResponseWriter, data JSON) {
	d := JSON{"ok": true, "data": data}
	log.Debug(d)
	ResponseJson(w, 200, d)
}

func ResponseApiError(log *logrus.Entry, w http.ResponseWriter, error string, data JSON) {
	d := JSON{"ok": false, "error": error, "data": data}
	log.Warn(d)
	ResponseJson(w, 500, d)
}

func ResponseApiErrorWithStatusCode(log *logrus.Entry, w http.ResponseWriter, statusCode int, error string, data JSON) {
	d := JSON{"ok": false, "error": error, "data": data}
	log.Warn(d)
	ResponseJson(w, statusCode, d)
}
