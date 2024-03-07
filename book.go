package kpless

import (
	"errors"
	"fmt"
	"strings"
)

type Book struct {
	Name      string         `json:"name"`
	SceneList []*Scene       `json:"scene_list"`
	Scenes    map[int]string `json:"scenes"`
}

type Text struct {
	val string
}

type Opt struct {
	NextId          int    `json:"next_id"`
	NextTip         string `json:"next_tip"`
	NextTitle       string `json:"next_title"`
	BeforeCondition string `json:"before_condition"`
	AfterEval       string `json:"after_eval"`
}

type Scene struct {
	Id         int
	Title      string `json:"title"`
	Level      int    `json:"level"`
	Block      []any  `json:"block"`
	parent     *Scene
	children   []*Scene
	ParentId   int   `json:"parent_id"`
	ChildrenId []int `json:"children_id"`
}

func (s *Scene) AutoId() {
	s.ParentId = s.parent.Id
	for _, child := range s.children {
		s.ChildrenId = append(s.ChildrenId, child.Id)
	}
}

func (s *Scene) Jump(opt *Opt, g *Game) (string, error) {
	if s.Level == 1 {
		for _, s2 := range g.book.SceneList {
			if s2.Title == opt.NextTitle {
				return s2.Execute(g), nil
			}
		}
		return "", errors.New("找不到要跳转的目标场景 请联系作者修改")
	}
	for _, child := range s.children {
		if child.Title == opt.NextTitle {
			return child.Execute(g), nil
		}
	}
	if s.parent != nil {
		return s.parent.Jump(opt, g)
	}
	return "", errors.New("不是顶级场景又不是分支场景 请联系作者")
}

func (s *Scene) Execute(g *Game) string {
	g.scene = s
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
			sb.WriteString(fmt.Sprintf("%d. %s => %s", i, a.NextTip, a.NextTitle))
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
	s.children = append(s.children, c)
}

func (s *Scene) AddBlock(b any) {
	s.Block = append(s.Block, b)
}
