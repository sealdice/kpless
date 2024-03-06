package kpless

import (
	"fmt"
	"strings"
)

type Book struct {
	Name   string   `json:"name"`
	Scenes []*Scene `json:"scenes"`
}

type Text struct {
	val string
}

type Opt struct {
	cond   string
	tip    string
	target string
}

type Scene struct {
	Title    string   `json:"title"`
	Level    int      `json:"level"`
	Block    []any    `json:"block"`
	Children []*Scene `json:"children"`
	parent   *Scene
}

func (s *Scene) Execute(g *Game) string {
	g.ResetOpt()
	var sb strings.Builder
	i := 0
	sb.WriteString("#")
	sb.WriteString(s.Title)
	sb.WriteString("\n")
	for _, b := range s.Block {
		switch a := b.(type) {
		case *Text:
			sb.WriteString(a.val)
			sb.WriteString("\n")
		case *Opt:
			g.AddOpt(a)
			i++
			sb.WriteString(fmt.Sprintf("%d. %s => %s", i, a.tip, a.target))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Find 依次查找 level 小于等于 n 的父节点
func (s *Scene) Find(n int) *Scene {
	if s.parent == nil || n <= 1 {
		return s
	}
	if s.parent.Level <= n {
		return s.parent
	} else {
		return s.parent.Find(n)
	}
}

func (s *Scene) AddChild(c *Scene) {
	c.parent = s
	s.Children = append(s.Children, c)
}

func (s *Scene) AddBlock(b any) {
	s.Block = append(s.Block, b)
}

func (s *Scene) Jump(opt *Opt, g *Game) (string, error) {
	if s.Level == 1 {
		for _, s2 := range g.book.Scenes {
			if s2.Title == opt.target {
				return s2.Execute(g), nil
			}
		}
	}
	for _, child := range s.Children {
		if child.Title == opt.target {
			return child.Execute(g), nil
		}
	}
	if s.parent != nil {
		return s.parent.Jump(opt, g)
	}
	return "", nil
}
