package main

import (
	"bufio"
	"data/datatype"
	"flag"
	"fmt"
	"libs/log"
	"libs/parser"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

var (
	path   = flag.String("p", "", "entity define file's path")
	tpl    = flag.String("t", "", "template file's path")
	output = flag.String("o", "", "parser file's output path")
	clear  = flag.Bool("c", false, "clear output path")
)

func GetDbUpdateName(typ string, name string) string {
	name = strings.ToLower(name)
	if t := datatype.GetUserType(typ, name); t != nil {
		ret := ""
		sep := ""
		for k, v := range t {
			sep = ""
			if k != 0 {
				sep = ","
			}
			ret += fmt.Sprintf("%s`%s`=?", sep, v[0])
		}
		return ret
	}

	return fmt.Sprintf("`p_%s`=?", strings.ToLower(name))
}

func GetDbInsertName(typ string, name string) string {
	name = strings.ToLower(name)
	if t := datatype.GetUserType(typ, name); t != nil {
		ret := ""
		sep := ""
		for k, v := range t {
			sep = ""
			if k != 0 {
				sep = ","
			}
			ret += fmt.Sprintf("%s`%s`", sep, v[0])
		}
		return ret
	}

	return fmt.Sprintf("`p_%s`", strings.ToLower(name))
}

func GetPlaceHold(typ string, name string) string {
	name = strings.ToLower(name)
	if t := datatype.GetUserType(typ, name); t != nil {
		ret := ""
		sep := ""
		for k := range t {
			sep = ""
			if k != 0 {
				sep = ","
			}
			ret += sep + "?"
		}
		return ret
	}

	return "?"
}

func GetDbValueStr(typ string, pre string, name string) string {
	switch typ {
	case "Vector2":
		name = pre + "." + name
		return fmt.Sprintf("%s.X, %s.Y", name, name)
	case "Vector3":
		name = pre + "." + name
		return fmt.Sprintf("%s.X, %s.Y, %s.Z", name, name, name)
	default:
		return pre + "." + name
	}
}

func createEntity(obj *parser.Object, outfile string) {

	t, err := template.New("object.tpl").Funcs(template.FuncMap{"tolower": strings.ToLower,
		"toupper":    strings.ToUpper,
		"ispublic":   IsPublic,
		"isprivate":  IsPrivate,
		"updatename": GetDbUpdateName,
		"insertname": GetDbInsertName,
		"placehold":  GetPlaceHold,
		"valuestr":   GetDbValueStr}).ParseFiles(*tpl + "/object.tpl")
	if err != nil {
		fmt.Println("writer", err)
		return
	}

	//save file
	file, err := os.Create(outfile)
	if err != nil {
		fmt.Println("writer", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	err = t.Execute(writer, obj)
	if err != nil {
		fmt.Println("writer", err)
	}

	writer.Flush()

	cmd := exec.Command("gofmt", "--w", outfile)
	cmd.Run()

	log.TraceInfo("entity", "create file:", outfile)

}

func IsPublic(val string) string {
	if strings.Index(val, "other") > -1 {
		return "true"
	}

	return "false"
}

func IsPrivate(val string) string {
	if strings.Index(val, "self") > -1 {
		return "true"
	}

	return "false"
}

func main() {

	flag.Parse()
	if *path == "" || *output == "" || *tpl == "" {
		fmt.Println("usage:data -t ./tpl -p ./entity -o ../src/data/entity [-c]")
		flag.PrintDefaults()
		return
	}

	if *clear {
		outdir, _ := os.Open(*output)
		outfiles, _ := outdir.Readdir(0)
		for _, f := range outfiles {
			if !f.IsDir() {
				err := os.Remove(*output + "/" + f.Name())
				if err != nil {
					fmt.Println(err)
					return
				}
				log.TraceInfo("entity", "delete file:", *output+"/"+f.Name())
			}
		}
	}

	parser.LoadAllDef(*path)
	reg := make([]map[string]string, len(parser.Defs))
	idx := 0
	for k, obj := range parser.Defs {
		createEntity(obj, *output+"/"+strings.ToLower(k)+".go")
		reg[idx] = map[string]string{"Name": obj.Name, "Type": strings.ToUpper(obj.Type)}
		idx++
	}

	/*dir, _ := os.Open(*path)
	files, _ := dir.Readdir(0)

	reg := make([]map[string]string, len(files))

	for idx, f := range files {
		if !f.IsDir() {
			obj := parser.ParseEntity(*path + "/" + f.Name())
			createEntity(obj, *output+"/"+strings.Replace(f.Name(), ".xml", ".go", -1))
			reg[idx] = map[string]string{"Name": obj.Name, "Type": strings.ToUpper(obj.Type)}
		}
	}*/

	file, err := os.Create(*output + "/" + "entity.go")
	if err != nil {
		fmt.Println("writer", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	t1, err1 := template.ParseFiles(*tpl + "/reg.tpl")
	if err1 != nil {
		fmt.Println("writer", err1)
		return
	}
	t1.Execute(writer, reg)
	writer.Flush()
	log.TraceInfo("entity", "create file:", *output+"/"+"entity.go")
	cmd := exec.Command("gofmt", "--w", *output+"/"+"entity.go")
	cmd.Run()
	reader := bufio.NewReader(os.Stdin)
	log.TraceInfo("entity", "press any key to quit")
	reader.ReadString('\n')
}
