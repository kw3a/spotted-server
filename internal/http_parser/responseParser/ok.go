package responseparser

import "net/http"

func Ok(w http.ResponseWriter) {
	w.WriteHeader(200)
}
