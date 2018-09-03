package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/util"
)

func main() {
	var ignore string
	flag.StringVar(&ignore, "ignore", "", "")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("用法：  " + os.Args[0] + " 文件名")
		return
	}

	filename := args[0]

	var menuList []toolbox.Menu
	err := util.FromHjsonFile(filename, &menuList)
	if err != nil {
		fmt.Println(err)
		return
	}

	ignoreList := strings.Split(ignore, ",")
	isIgnore := func(name string) bool {
		for _, nm := range ignoreList {
			if nm == name {
				return true
			}
		}
		return false
	}

	formatMenus(os.Stdout, isIgnore, menuList, 0, true)
}

func formatMenus(out io.Writer, isIgnore func(name string) bool, menuList []toolbox.Menu, layer int, indent bool) {
	if layer > 0 && indent {
		out.Write(bytes.Repeat([]byte("  "), layer))
	}
	out.Write([]byte("[\r\n"))
	layer++
	for idx, menu := range menuList {
		if layer > 0 {
			out.Write(bytes.Repeat([]byte("  "), layer))
		}
		out.Write([]byte("{"))

		needComma := false
		if menu.UID != "" && !isIgnore("uid") {
			io.WriteString(out, `"uid":"`)
			io.WriteString(out, menu.UID)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.Title != "" && !isIgnore("title") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"title":"`)
			io.WriteString(out, menu.Title)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.Permission != "" && !isIgnore("permission") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"permission":"`)
			io.WriteString(out, menu.Permission)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.License != "" && !isIgnore("license") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"license":"`)
			io.WriteString(out, menu.License)
			io.WriteString(out, "\"")
			needComma = true
		}
		if menu.Icon != "" && !isIgnore("icon") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"icon":"`)
			io.WriteString(out, menu.Icon)
			io.WriteString(out, "\"")
		}

		if menu.Classes != "" && !isIgnore("classes") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"classes":"`)
			io.WriteString(out, menu.Classes)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.URL != "" && !isIgnore("url") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"url":"`)
			io.WriteString(out, menu.URL)
			io.WriteString(out, "\"")
			needComma = true
		}

		if len(menu.Children) > 0 && !isIgnore("children") {
			if needComma {
				io.WriteString(out, `,`)
			}

			out.Write([]byte("\r\n"))
			if layer > 0 {
				out.Write(bytes.Repeat([]byte("  "), layer+1))
			}

			io.WriteString(out, `"children":`)
			formatMenus(out, isIgnore, menu.Children, layer+1, false)
		}

		out.Write([]byte("}"))

		if idx != len(menuList)-1 {
			out.Write([]byte(",\r\n"))
		} else {
			out.Write([]byte("\r\n"))
		}
	}

	if (layer - 1) > 0 {
		out.Write(bytes.Repeat([]byte("  "), layer))
	}
	out.Write([]byte("]"))
}
