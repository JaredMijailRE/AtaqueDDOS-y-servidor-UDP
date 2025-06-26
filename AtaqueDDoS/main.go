package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	targetAddr = "127.0.0.1:53" // Dirección del servidor DNS
	domainName = "example.com." // Dominio simulado
)

var (
	requests    int
	concurrency int
)

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Advertencia: valor inválido para %s: %s. Usando %d", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func init() {
	// Si REQUESTS no está definida, se comporta como cliente normal (10 requests)
	requests = getEnvInt("REQUESTS", 10)
	// La concurrencia por defecto es igual a requests, pero puedes cambiarlo si quieres
	concurrency = getEnvInt("CONCURRENCY", requests)
}

func getTargetAddr() string {
	addr := os.Getenv("TARGET_ADDR")
	if addr == "" {
		addr = "dns_udp:53"
	}
	return addr
}

func getClientType() string {
	if requests > 1000 {
		return "malicioso"
	}
	return "normal"
}

var csvFilePath = "/app/resultados/resultados.csv"
var csvHeader = []string{"timestamp", "id", "success", "client_type", "target_addr", "domain", "latency_ms"}

func writeCSVHeaderIfNeeded() {
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		file, err := os.OpenFile(csvFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("No se pudo crear %s: %v", csvFilePath, err)
			return
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		if err := writer.Write(csvHeader); err != nil {
			log.Printf("No se pudo escribir encabezado en %s: %v", csvFilePath, err)
		}
	}
}

func logToCSV(id int, success bool, latencyMs int64) {
	writeCSVHeaderIfNeeded()
	file, err := os.OpenFile(csvFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("No se pudo abrir %s: %v", csvFilePath, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	timestamp := time.Now().Format(time.RFC3339Nano)
	targetAddr := getTargetAddr()
	record := []string{
		timestamp,
		strconv.Itoa(id),
		strconv.FormatBool(success),
		getClientType(),
		targetAddr,
		domainName,
		strconv.FormatInt(latencyMs, 10),
	}
	if err := writer.Write(record); err != nil {
		log.Printf("No se pudo escribir en %s: %v", csvFilePath, err)
	}
}

func sendQuery(wg *sync.WaitGroup, id int, targetAddr string) {
	defer wg.Done()

	c := new(dns.Client)
	c.Net = "udp"
	m := new(dns.Msg)
	m.SetQuestion(domainName, dns.TypeA)

	start := time.Now()
	_, _, err := c.Exchange(m, targetAddr)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		log.Printf("[#%d] Error: %v\n", id, err)
		logToCSV(id, false, latency)
	} else {
		log.Printf("[#%d] Consulta exitosa a %s\n", id, domainName)
		logToCSV(id, true, latency)
	}
}

func main() {
	targetAddr := getTargetAddr()
	start := time.Now()
	log.Printf("Iniciando %d consultas DNS con %d hilos a %s...\n", requests, concurrency, targetAddr)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer func() { <-sem }()
			sendQuery(&wg, i, targetAddr)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Completadas %d peticiones en %s\n", requests, elapsed)
}
