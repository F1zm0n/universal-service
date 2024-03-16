package transport

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"

	models "github.com/F1zm0n/universal-mailer/internal"
	"github.com/F1zm0n/universal-mailer/internal/service"
)

type HttpTransport struct {
	svc service.Service
}

func NewHttpTransport() HttpTransport {
	return HttpTransport{
		svc: service.NewBaseService(),
	}
}

type errorWrapper struct {
	Err error `json:"error"`
}

func (t HttpTransport) HandleMail(w http.ResponseWriter, r *http.Request) {
	var val models.VerDto
	err := json.NewDecoder(r.Body).Decode(&val)
	if err != nil {
		fmt.Println("val: ", val)
		writeJson(w, errorWrapper{Err: err}, 500)
		return
	}
	defer r.Body.Close()
	err = t.svc.SendEmail(val)
	if err != nil {
		writeJson(w, errorWrapper{Err: err}, 500)
		return
	}
	writeJson(w, errorWrapper{Err: nil}, 200)
}

func (t HttpTransport) HandleVerify(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	verId, err := uuid.Parse(id)
	if err != nil {
		writeJson(w, errorWrapper{Err: err}, 500)
		return
	}
	err = t.svc.VerifyMail(verId)
	if err != nil {
		log.Println("error verifying email ", err)
		writeJson(w, errorWrapper{Err: err}, 500)
		return
	}
	writeJson(w, errorWrapper{Err: nil}, 200)
}

func writeJson(w http.ResponseWriter, v any, code int) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		return err
	}
	return nil
}
