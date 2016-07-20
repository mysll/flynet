package helper

import (
	"encoding/csv"
	"io"
	"os"
	"server/libs/log"
	"strings"

	"github.com/bitly/go-simplejson"
)

type header struct {
	name  string
	index int
}

const (
	OP_ADD = iota + 1
	OP_SUB
	OP_MUL
	OP_DIV
	OP_SET
)

type PropOp struct {
	Prop   string
	Option int
	Value  string
}

type PropInfo struct {
	Info map[string][]PropOp
}

type csvparser struct {
	entity    int
	heads     []header
	headindex map[string]int
	infos     map[string][]string
	PropOpt   map[string]PropInfo
}

func NewParser() *csvparser {
	p := &csvparser{}
	p.infos = make(map[string][]string)
	p.PropOpt = make(map[string]PropInfo)
	p.headindex = make(map[string]int)
	return p
}

func (p *csvparser) GetKeyIndex(key string) int {
	if index, ok := p.headindex[key]; ok {
		return index
	}

	return -1
}

func (p *csvparser) GetIds() []string {
	ret := make([]string, 0, len(p.infos))
	for k, _ := range p.infos {
		ret = append(ret, k)
	}

	return ret
}

func (p *csvparser) GetPropOpt(id string, prop string) []PropOp {
	if info, ok := p.PropOpt[id]; ok {
		if p, ok := info.Info[prop]; ok {
			return p
		}
	}
	return nil
}

func (p *csvparser) Find(id string) []string {
	if info, ok := p.infos[id]; ok {
		return info
	}
	return nil
}

func (p *csvparser) Parse(f string) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}

	defer file.Close()

	p.infos = make(map[string][]string)
	reader := csv.NewReader(file)
	head := false
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line[0], "//") {
			continue
		}

		if !head {
			if line[0] == "ID" {
				p.heads = make([]header, 0, len(line))
				for k, h := range line {
					if strings.HasPrefix(h, "//") {
						continue
					}
					p.headindex[h] = len(p.heads)
					p.heads = append(p.heads, header{h, k})
					if h == "Entity" {
						p.entity = len(p.heads)
					}
				}
				head = true
			}
		} else {
			id := line[0]
			if id == "" {
				continue
			}
			if _, ok := p.infos[id]; ok {
				log.LogFatalf("config file ", f, " id repeat ", id)
				continue
			}

			lineinfo := make([]string, len(p.heads))
			for k, hi := range p.heads {
				lineinfo[k] = line[hi.index]
				if strings.HasSuffix(hi.name, "_script") {
					if p.entity != 0 && lineinfo[p.entity-1] != "" {
						p.ParseOp(id, hi.name, lineinfo[k])
					}

				}
			}
			p.infos[id] = lineinfo
		}

	}

	return nil
}

func (p *csvparser) ParseOp(id, prop, script string) {
	defer func() {
		if e := recover(); e != nil {
			log.LogError("parse op error:", e)
		}
	}()

	var info PropInfo
	var ok bool
	if info, ok = p.PropOpt[id]; !ok {
		info = PropInfo{Info: make(map[string][]PropOp)}
		p.PropOpt[id] = info
	}

	if script == "" {
		return
	}

	json, err := simplejson.NewJson([]byte(script))
	if err != nil {
		log.LogError("parse script error:", err)
		return
	}

	if ps, ok := json.CheckGet("propertys"); ok {
		index := 0
		oparr := make([]PropOp, 0, 16)
	L:
		for {
			js := ps.GetIndex(index)
			if js.Interface() == nil {
				break
			}

			index++
			ops, err := js.StringArray()
			if err != nil || len(ops) != 3 {
				continue
			}

			op := PropOp{}
			op.Prop = ops[0]
			switch ops[1] {
			case "+":
				op.Option = OP_ADD
				break
			case "-":
				op.Option = OP_SUB
				break
			case "*":
				op.Option = OP_MUL
				break
			case "/":
				op.Option = OP_DIV
				break
			case "=":
				op.Option = OP_SET
				break
			default:
				continue L
			}
			op.Value = ops[2]

			oparr = append(oparr, op)
		}
		info.Info[prop] = oparr
	}

}
