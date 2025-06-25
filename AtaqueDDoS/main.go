package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	targetAddr  = "127.0.0.1:53" // Dirección del servidor DNS
	domainName  = "example.com." // Dominio simulado
	requests    = 1000           // Total de peticiones
	concurrency = 100            // Cantidad de goroutines simultáneas
)

func sendQuery(wg *sync.WaitGroup, id int) {
	defer wg.Done()

	c := new(dns.Client)
	c.Net = "udp"
	m := new(dns.Msg)
	m.SetQuestion(domainName, dns.TypeA)

	_, _, err := c.Exchange(m, targetAddr)
	if err != nil {
		log.Printf("[#%d] Error: %v\n", id, err)
	} else {
		log.Printf("[#%d] Consulta exitosa a %s\n", id, domainName)
	}
}

func main() {
	start := time.Now()
	log.Printf("Iniciando %d consultas DNS con %d hilos...\n", requests, concurrency)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer func() { <-sem }()
			sendQuery(&wg, i)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Completadas %d peticiones en %s\n", requests, elapsed)
}
