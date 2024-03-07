package kpless

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Game struct {
	kp    *KPLess
	scene *Scene
	book  *Book
	opts  []*Opt
}

func (g *Game) Next(content string) (string, error) {
	if g.scene == nil {
		g.scene = g.book.SceneList[0]
		return g.scene.Execute(g), nil
	}

	content = strings.TrimSpace(content)
	if content == "" {
		content = "1"
	}

	if i, err := strconv.ParseUint(content, 10, 32); err == nil {
		n := int(i - 1)
		if n < 0 || n >= len(g.opts) {
			return "", fmt.Errorf("输入的数字与选项不匹配 %d", i)
		}
		return g.scene.Jump(g.opts[n], g)
	}

	for _, opt := range g.opts {
		if opt.NextTip == content {
			return g.scene.Jump(opt, g)
		}
	}
	return "不理解选项含义 游戏结束", io.EOF
}

func (g *Game) ResetOpt() {
	g.opts = nil
}

func (g *Game) AddOpt(o *Opt) {
	g.opts = append(g.opts, o)
}
