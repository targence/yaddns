package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type config struct {
	Token     string
	Domain    string
	Subdomain string
	TTL       int
}

type request struct {
	URL      string
	PddToken string
	Content  string
	Domain   string
	RecordID int
	TTL      int
	Method   string
}

type domainInfo struct {
	Domain  string `json:"domain"`
	Records []struct {
		Content   string `json:"content"`
		Domain    string `json:"domain"`
		Fqdn      string `json:"fqdn"`
		RecordID  int    `json:"record_id"`
		Subdomain string `json:"Subdomain"`
		TTL       int    `json:"ttl"`
		Type      string `json:"type"`
	} `json:"records"`
	Error   string `json:"error"`
	Success string `json:"success"`
}

type updateRecordResponse struct {
	Domain string `json:"domain"`
	Record struct {
		Content   string `json:"content"`
		Domain    string `json:"domain"`
		Fqdn      string `json:"fqdn"`
		Operation string `json:"operation"`
		Priority  string `json:"priority"`
		RecordID  int    `json:"record_id"`
		Subdomain string `json:"subdomain"`
		TTL       int    `json:"ttl"`
		Type      string `json:"type"`
	} `json:"record"`
	RecordID int    `json:"record_id"`
	Success  string `json:"success"`
	Error    string `json:"error"`
}

func getIP() string {
	var addr string
	r := request{}
	r.Method = "GET"
	r.URL = "http://ipv4.myexternalip.com/raw"
	body := getURL(r)
	addr = strings.Trim(string(body), " \r\n")

	log.Printf("IP address received %s\n", addr)
	return addr
}

var client = http.Client{Timeout: time.Duration(20 * time.Second)}

func getURL(r request) []byte {
	params := ""
	if r.Method == "POST" {
		params = fmt.Sprintf("domain=%s&record_id=%d&ttl=%d&content=%s", r.Domain, r.RecordID, r.TTL, r.Content)
	}

	req, err := http.NewRequest(r.Method, r.URL, strings.NewReader(params))
	req.Header.Set("User-Agent", "yaddns")
	if err != nil {
		log.Fatalf("error: '%v'\n", err)
	}

	if r.PddToken != "" {
		req.Header.Set("PddToken", r.PddToken)
	}

	if r.Method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		req.Header.Set("User-Agent", "yaddns")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error: '%v'\n", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("error: unexpecetd HTTP status code: %d\n", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error: '%v'\n", err)
	}

	return body
}

func getDomainInfo(conf config) domainInfo {
	getDomainInfoURLTemplate := "https://pddimp.yandex.ru/api2/admin/dns/list?token=%s&domain=%s"

	r := request{}
	r.Method = "GET"
	r.URL = fmt.Sprintf(getDomainInfoURLTemplate, conf.Token, conf.Domain)
	r.PddToken = conf.Token

	body := getURL(r)

	info := parseDomainInfoData(body)
	verifyDomainInfoData(info, conf)
	log.Println("list of dns record received")

	return info
}

func parseDomainInfoData(data []byte) domainInfo {
	info := domainInfo{}

	err := json.Unmarshal(data, &info)
	if err != nil {
		log.Fatalf("failed to parse response from Yandex DNS API service %v\n", err)
	}

	return info
}

func verifyDomainInfoData(info domainInfo, conf config) {
	if info.Error != "" {
		log.Fatalf("invalid status while calling 'dns/list' Yandex DNS API command: %v\n", info.Error)
	}

	if info.Domain != conf.Domain {
		log.Fatalf("invariand failed: %s != %s\n", info.Domain, conf.Domain)
	}

	if len(info.Records) == 0 {
		log.Fatalf("empty response\n")
	}
}

func verifyUpdateRecordResponse(data []byte) {
	resp := updateRecordResponse{}

	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Fatalf("failed to parse response from Yandex DNS API service %v\n", err)
	}

	if resp.Error != "" {
		log.Fatalf("update failed, error message: %v\n", resp.Error)
	}
}

func updateDomainAddress(info domainInfo, extIPAddr string, conf config) {

	var r request
	update := func() {
		body := getURL(r)

		verifyUpdateRecordResponse(body)
		log.Printf("IP address for '%s' set to %s\n", conf.Subdomain, extIPAddr)
	}

	updated := false
	for _, record := range info.Records {
		if record.Fqdn != conf.Subdomain {
			continue
		}

		if record.Type == "A" && extIPAddr != "" {
			r.Method = "POST"
			r.URL = "https://pddimp.yandex.ru/api2/admin/dns/edit"
			r.PddToken = conf.Token
			r.Domain = conf.Domain
			r.RecordID = record.RecordID
			r.TTL = conf.TTL
			r.Content = extIPAddr
		} else {
			continue
		}

		update()
		updated = true
	}

	if !updated {
		log.Fatalf("domain '%s' not known for Yandex.DNS\n", conf.Subdomain)
	}
}

func main() {
	conf := config{}
	conf.Token = "FUTFBGIUYUFLSVGVD5WAJH4TX343BBJFHGDSFGSF"
	conf.Domain = "yourdomain.com"
	conf.Subdomain = "home.yourdomain.com" // repeat Domain if there is no Subdomain
	conf.TTL = 900

	extIPAddr := getIP()
	domainInfo := getDomainInfo(conf)
	updateDomainAddress(domainInfo, extIPAddr, conf)
}
