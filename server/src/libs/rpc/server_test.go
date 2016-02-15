// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	"testing"
)

type Serv1 struct {
}

func (s *Serv1) Add(src MailBox, i int, j int) error {
	fmt.Println(i, j)
	return nil
}

func (s *Serv1) TestArgs(src MailBox, i int, j float32, msg string) error {
	fmt.Println(i, j, msg)
}
func TestRpc(t *testing.T) {
	s1 := make(map[string]interface{})
	s1["Serv1"] = &Serv1
	s, err := CreateRpcService(s1, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	CreateService(s)
}
