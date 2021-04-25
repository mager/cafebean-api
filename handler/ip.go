package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// IP represents IP address
type IP struct {
	IPAddress string `json:"ip_address"`
}

type IPResp struct {
	IP IP `json:"ip"`
}

func (h *Handler) getIP(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &IPResp{}
	)

	ipResp, err := http.Get("https://curlmyip.org")
	if err != nil {
		h.logger.Fatalf(err.Error())
	}

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(ipResp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	resp.IP.IPAddress = string(body)

	json.NewEncoder(w).Encode(resp)
}
