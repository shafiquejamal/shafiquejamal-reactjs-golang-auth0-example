package errorhandler

import (
	"log"
	"net/http"
)

func ReturnError(w *http.ResponseWriter, sC int, m string, err error) {
	(*w).WriteHeader(sC)
	(*w).Write([]byte(m))
	log.Println(m, err)
}
