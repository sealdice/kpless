package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
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
	case kpless.LoadGameNoFoundBookT:
		return "你玩的模组已经绝版了！换个游戏玩吧！"
	case kpless.ModBreakUpdate:
		return "你玩的模组发生了破坏性更新，游戏可能无法进入正确流程，还有继续吗"
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

type logger struct{}

func (l logger) Infof(format string, args ...any) {
	fmt.Println(fmt.Sprintf(format, args...))
}

type Render struct {
	Name string
	Data string
}

func main() {
	log := logger{}
	kp := kpless.New(log)
	vm := &dsvm{ds.NewVM()}
	var book string
	var gameName string
	var mode bool
	flag.StringVar(&book, "file", "向火独行.md", "指定 md 文件")
	flag.StringVar(&gameName, "game", "向火独行", "指定游戏")
	flag.BoolVar(&mode, "mode", true, "")
	flag.Parse()
	err := kp.LoadMarkDownBook(book)
	if err != nil {
		panic(err)
	}
	err = kp.SetGame(vm, "cli", gameName)
	if err != nil {
		panic(err)
	}
	if mode {
		text, _ := os.ReadFile("tmpl.html")
		tmpl, _ := template.New("index").Parse(string(text))
		_ = os.Remove("index.html")
		jsonRaw, _ := json.Marshal(kp.Books[gameName])
		r := Render{
			Name: kp.Games["cli"].BookName,
			Data: string(jsonRaw),
		}
		file, _ := os.Create("index.html")
		_ = tmpl.Execute(file, r)
		return
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
			fmt.Println(err.Error())
		}
		fmt.Println(res)
	}
}
