package main

import (
	"HW/controllers/stdhttp"
	"HW/gate/psg"
	"fmt"
	"log"
)
func main() {
	psg, err := psg.NewPsg("postgres://127.0.0.1:5432/svepodd", "postgres", "qwerty123")

	if err != nil {
		log.Println(err)
	}

	c := stdhttp.NewController(":8080", psg)
	err = c.Start()
	if err != nil{
		fmt.Println("Error occuped:", err)
		return
	}
}