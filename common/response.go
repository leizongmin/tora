package common

import (
	"encoding/json"
	"fmt"
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

func ResponseApiOk(w http.ResponseWriter, data JSON) {
	ResponseJson(w, 200, JSON{"ok": true, "data": data})
}

func ResponseApiError(w http.ResponseWriter, error string, data JSON) {
	ResponseJson(w, 500, JSON{"ok": false, "error": error, "data": data})
}
