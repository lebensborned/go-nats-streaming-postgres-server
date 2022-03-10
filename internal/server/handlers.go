package server

import (
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

// Handler for /order/{id}
func (s *Server) GetOrderById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := vars["id"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
		return
	}
	order, err := s.repo.FindById(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Order not found"))
		return
	}
	tmp, err := template.ParseFiles("ui/templates/order.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = tmp.Execute(w, order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

}
