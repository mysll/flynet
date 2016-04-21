package main

import (
	"fmt"
	"util"
)

func main() {
	ar := util.NewStoreArchiver(nil)
	data := make([]string, 0, 2)
	data = append(data, "123")
	data = append(data, "456")
	ar.Write(data)
	lr := util.NewLoadArchiver(ar.Data())
	var data1 []string
	lr.ReadObject(&data1)
	fmt.Println(data1)
}
