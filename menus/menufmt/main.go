package main

import (
	"flag"
	"fmt"
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

	toolbox.FormatMenus(os.Stdout, isIgnore, menuList, 0, true)
}
