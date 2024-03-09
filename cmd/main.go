package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"kpless"

	ds "github.com/sealdice/dicescript"
)

type dsvm struct {
	*ds.Context
}

func (v *dsvm) Exec(text string) string {
	err := v.Run(text)
	if err != nil {
		panic(err)
	}
	return v.Ret.ToString()
}

func (v *dsvm) ExecCond(text string) bool {
	v.Exec(text)
	return v.Ret.AsBool()
}

func getCaption(i kpless.Caption) string {
	switch i {
	case kpless.NoFoundBook:
		return "未配置冒险书，无法开始游戏"
	case kpless.NoFoundJumpToScene:
		return "找不到要跳转的目标场景 请联系作者修改"
	case kpless.NoTopNoForkScene:
		return "不是顶级场景又不是分支场景 请联系作者"
	case kpless.OptCaption:
		return " $t选项序号 .  $t选项内容  =>  $t目标场景标题 "
	case kpless.TextCaption:
		return "　　 $t文本行 "
	case kpless.TitleCaption:
		return "#  $t场景标题 "
	}
	return "此处应该有文案 但是没有，请联系开发者"
}

func (v *dsvm) ExecCaption(i kpless.Caption) string {
	re := regexp.MustCompile(`\s\$t.+?\s`)
	t := getCaption(i)
	return re.ReplaceAllStringFunc(t, v.Exec)
}

func (v *dsvm) Store(name, val string) {
	v.StoreName(name, ds.VMValueNewStr(val))
}

func main() {
	kp := kpless.New()
	vm := &dsvm{ds.NewVM()}
	var book string
	var gameName string
	flag.StringVar(&book, "file", "向火独行.md", "指定 md 文件")
	flag.StringVar(&gameName, "game", "向火独行", "指定游戏")
	flag.Parse()
	err := kp.LoadMarkDown(book)
	if err != nil {
		panic(err)
	}
	err = kp.SetGame(vm, "cli", gameName)
	if err != nil {
		panic(err)
	}
	inputs := bufio.NewScanner(os.Stdin)
	for inputs.Scan() {
		if inputs.Text() == "exit" {
			break
		}
		res, err := kp.Input(vm, "cli", inputs.Text())
		if errors.Is(err, io.EOF) {
			fmt.Println(res)
			os.Exit(1)
		}
		fmt.Println("-------------")
		if err != nil {
			fmt.Println(err.Error(), res)
		}
		fmt.Println(res)
	}
	js, _ := json.MarshalIndent(kp, "", "  ")
	os.WriteFile("tmp.json", js, 066)
}
