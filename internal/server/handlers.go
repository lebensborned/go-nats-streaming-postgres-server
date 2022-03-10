package server

import (
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

type Response struct {
	Msg interface{} `json:"msg"`
}

func writeStringResponse(v interface{}) []byte {
	resp := Response{Msg: v}
	bytes, _ := json.Marshal(resp)
	return bytes
}

// Handler for /order/{id}
func (s *Server) GetOrderById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := vars["id"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write(writeStringResponse("Bad request"))
		return
	}
	order, err := s.repo.FindById(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		w.Write(writeStringResponse("Order not found"))
		return
	}
	tmp, err := template.ParseFiles("ui/templates/order.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(writeStringResponse(err)))
		return
	}
	err = tmp.Execute(w, order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(err.Error()))
		return
	}

}
