package kpless

import (
	"errors"
	"strconv"
	"strings"
)

type Book struct {
	Name      string         `json:"name"`
	SceneList []*Scene       `json:"scene_list"`
	Scenes    map[int]string `json:"scenes"`
}

type Code struct {
	val string
}

type Opt struct {
	Content         string `json:"next_tip"`
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

func (s *Scene) jump(vm RollVM, opt *Opt, g *Game) (string, error) {
	if s.Level == 1 {
		for _, top := range g.book.SceneList {
			if top.Title == opt.NextTitle {
				return top.Execute(vm, g), nil
			}
		}
		return "", errors.New(vm.ExecCaption(NoFoundJumpToScene))
	}
	for _, child := range s.children {
		if child.Title == opt.NextTitle {
			return child.Execute(vm, g), nil
		}
	}
	if s.parent != nil {
		return s.parent.jump(vm, opt, g)
	}
	return "", errors.New(vm.ExecCaption(NoTopNoForkScene))
}

func (s *Scene) Jump(vm RollVM, opt *Opt, g *Game) (string, error) {
	res, err := s.jump(vm, opt, g)
	if err == nil && opt.AfterEval != "" {
		_ = vm.Exec(opt.AfterEval)
	}
	return res, err
}

func (s *Scene) Execute(vm RollVM, g *Game) string {
	g.scene = s
	g.ResetOpt()
	var sb strings.Builder
	var i uint64
	vm.Store("$t场景标题", s.Title)
	sb.WriteString(vm.ExecCaption(TitleCaption))
	sb.WriteString("\n")
	for _, b := range s.Block {
		switch v := b.(type) {
		case *string:
			vm.Store("$t文本行", *v)
			sb.WriteString(vm.ExecCaption(TextCaption))
			sb.WriteString("\n")
		case *Opt:
			if v.BeforeCondition != "" {
				if vm.ExecCond(v.BeforeCondition) {
					continue
				}
			}
			g.AddOpt(v)
			i++
			vm.Store("$t选项序号", strconv.FormatUint(i, 10))
			vm.Store("$t选项内容", v.Content)
			vm.Store("$t目标场景标题", v.NextTitle)
			sb.WriteString(vm.ExecCaption(OptCaption))
			sb.WriteString("\n")
		case *Code:
			sb.WriteString(vm.Exec(v.val))
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
