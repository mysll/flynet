package main

import (
	"fmt"
	"server"
)

type Obj struct {
	I int
	J int
}

func main() {

	msg, err := server.CreateMessage(int8(1), int16(2), "test123", []byte{4, 5, 6}, &Obj{6, 7})
	fmt.Println(err)
	var i int8
	var j int16
	var k string
	var l []byte
	var o Obj
	err = server.ParseArgs(msg, &i, &j, &k, &l, &o)
	fmt.Println(i, j, k, l, o, err)
}
