package server

import (
	"fmt"
	"net/http"
	"vosskamp-reisen-3/internal/structs"
)

func (s *Server) RegisterEmailRoutes(mux *http.ServeMux) {
	mux.Handle("POST /email", http.HandlerFunc(s.sendEmail))
}

func (s *Server) sendEmail(w http.ResponseWriter, r *http.Request) {
	emailForm := structs.EmailData{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}
	if emailForm.Name == "" || emailForm.Email == "" || emailForm.Subject == "" || emailForm.Message == "" {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<p style='color: red;'> Fehler </p>"))
		return
	}
	formattedMessage := fmt.Sprintf("Kundenname: %s<br><br>Kunden E-mail: %s<br><br>Nachricht: %s",
		emailForm.Name,
		emailForm.Email,
		emailForm.Message)
	err := s.emailService.SendEmail([]string{"hendrikvosskamp@gmail.com"}, emailForm.Subject, formattedMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<p style='color: green;'> Danke für Ihre Anfrage. Wir melden uns bei Ihnen. </p>"))
}
