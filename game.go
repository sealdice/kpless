package kpless

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type Game struct {
	scene       *Scene
	book        *Book
	err         error
	Opts        []*Block `json:"opts"`
	BookName    string   `json:"book_name"`
	BookVersion string   `json:"book_version"`
	SceneId     int      `json:"scene_id"`
}

var EqualOne = []string{"", ".", "。"}

func WithEqualOne(sl []string) {
	EqualOne = sl
}

func (g *Game) Next(vm RollVM, content string) (string, error) {
	if g.err != nil {
		if errors.Is(g.err, ErrModBreakUpdate) {
			return vm.ExecCaption(ModBreakUpdate), nil
		}
		if errors.Is(g.err, ErrLoadGameNoFoundBook) {
			return vm.ExecCaption(LoadGameNoFoundBookT), nil
		}
	}

	if g.scene == nil {
		g.scene = g.book.scenes[0]
		return g.scene.Execute(vm, g), nil
	}

	content = strings.TrimSpace(content)
	if lo.Contains(EqualOne, content) {
		content = "1"
	}

	if i, err := strconv.ParseInt(content, 10, 32); err == nil {
		n := int(i - 1)
		if n < 0 || n >= len(g.Opts) {
			return "", fmt.Errorf("输入的数字超出选项区间 %d", i)
		}
		return g.scene.Jump(vm, g.Opts[n], g)
	}

	for _, opt := range g.Opts {
		if opt.NextTip == content {
			return g.scene.Jump(vm, opt, g)
		}
	}
	return "", errors.New("无法解析的输入")
}

func (g *Game) ResetOpt() {
	g.Opts = nil
}

func (g *Game) AddOpt(o *Block) {
	g.Opts = append(g.Opts, o)
}
