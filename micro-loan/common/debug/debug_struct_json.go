package main

import (
	"encoding/json"
	"fmt"
)

type Host struct {
	IP           string `json:"ip"`
	Name         string `json:"name"`
	groupCompany struct {
		Name string
		Code string
	}
}

func main() {

	//m := Host{Name: "Sky", IP: "192.168.23.92", GroupCompany: {Name: "chester", Code: "123"}}
	m := &Host{}

	m.IP = "192.168.23.92"
	m.Name = "Sky"
	//m.GroupCompany.Name = "chester"
	//m.GroupCompany.Code = "123"

	b, err := json.Marshal(m)
	if err != nil {

		fmt.Println("Umarshal failed:", err)
		return
	}

	fmt.Println("json:", string(b))
}
