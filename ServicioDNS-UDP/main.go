package main

import (
	"log"

	"github.com/miekg/dns"
)

// IP simulada para todas las respuestas
const fakeIP = "1.2.3.4"

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		log.Printf("Pregunta recibida: %s %s\n", question.Name, dns.TypeToString[question.Qtype])

		// Solo responde a tipo A (IPv4)
		if question.Qtype == dns.TypeA {
			rr, err := dns.NewRR(question.Name + " IN A " + fakeIP)
			if err == nil {
				msg.Answer = append(msg.Answer, rr)
			}
		}
	}

	w.WriteMsg(&msg)
}

func main() {
	// Configurar manejador
	dns.HandleFunc(".", handleDNSRequest)

	// Crear servidor UDP
	server := &dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	log.Println("Servidor DNS falso escuchando en UDP :53...")
	// Requiere permisos de root o usar puerto >1024
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error al iniciar servidor: %s\n", err.Error())
	}
}
