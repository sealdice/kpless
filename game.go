package kpless

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Game struct {
	kp       *KPLess
	scene    *Scene
	book     *Book
	Opts     []*Block `json:"opts"`
	BookName string   `json:"book_name"`
	SceneId  int      `json:"scene_id"`
}

func (g *Game) Next(ctx RollVM, content string) (string, error) {
	if g.scene == nil {
		g.scene = g.book.scenes[0]
		return g.scene.Execute(ctx, g), nil
	}

	content = strings.TrimSpace(content)
	if content == "" {
		content = "1"
	}

	if i, err := strconv.ParseUint(content, 10, 32); err == nil {
		n := int(i - 1)
		if n < 0 || n >= len(g.Opts) {
			return "", fmt.Errorf("输入的数字与选项不匹配 %d", i)
		}
		return g.scene.Jump(ctx, g.Opts[n], g)
	}

	for _, opt := range g.Opts {
		if opt.NextTip == content {
			return g.scene.Jump(ctx, opt, g)
		}
	}
	return "不理解选项含义 游戏结束", io.EOF
}

func (g *Game) ResetOpt() {
	g.Opts = nil
}

func (g *Game) AddOpt(o *Block) {
	g.Opts = append(g.Opts, o)
}
