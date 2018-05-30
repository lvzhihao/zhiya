package main

import (
	"log"

	"github.com/shopex/prism-go"
)

func main() {
	c, err := prism.NewClient("http://omsbbc2.shopexprism.onex.software:8080/api", "5k7y7hnq", "5ulx44honatwk7rnhnxd")
	if err != nil {
		log.Println("create client: ", err)
	}

	n, err := c.Notify()
	if err != nil {
		log.Println("create notify: ", err)
	}

	ch, err := n.Consume("messages")
	if err != nil {
		log.Println("consume queue: ", err)
	}

	for {
		msg := <-ch
		log.Printf("%#v\n", msg)
		msg.Ack()
	}
}
