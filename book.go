package kpless

import (
	"errors"
	"strconv"
	"strings"
)

func newBook() Book {
	return Book{Meta: map[string]string{}, Scenes: map[int]*Scene{}}
}

type Book struct {
	Name   string            `json:"name"`
	Meta   map[string]string `json:"meta"`
	Scenes map[int]*Scene    `json:"scenes"`
	scenes []*Scene
}

type blockType int

const (
	textType blockType = iota
	innerCodeType
	multiLineCodeType
	optType
)

type Block struct {
	Type          blockType `json:"type"`
	Text          string    `json:"text,omitempty"`
	InnerCode     string    `json:"inner_code,omitempty"`
	MultiLineCode string    `json:"code,omitempty"`
	// Opt
	NextTip         string `json:"next_tip,omitempty"`
	NextTitle       string `json:"next_title,omitempty"`
	BeforeCondition string `json:"before_condition,omitempty"`
	AfterEval       string `json:"after_eval,omitempty"`
}

func newText(text string) *Block {
	return &Block{Type: textType, Text: text}
}

func newInnerCode(code string) *Block {
	return &Block{Type: innerCodeType, InnerCode: code}
}

func newMultiLineCode(code string) *Block {
	return &Block{Type: multiLineCodeType, MultiLineCode: code}
}

func newOpt(tip, title, cond, after string) *Block {
	return &Block{
		Type:            optType,
		NextTip:         tip,
		NextTitle:       title,
		BeforeCondition: cond,
		AfterEval:       after,
	}
}

type Scene struct {
	Id         int      `json:"id"`
	Title      string   `json:"title"`
	Level      int      `json:"level"`
	Block      []*Block `json:"block"`
	ParentId   int      `json:"parent_id"`
	ChildrenId []int    `json:"children_id"`
	parent     *Scene
	children   []*Scene
}

func (s *Scene) jump(vm RollVM, opt *Block, g *Game) (string, error) {
	if s.Level == 1 {
		for _, top := range g.book.scenes {
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

func (s *Scene) Jump(vm RollVM, opt *Block, g *Game) (string, error) {
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
		switch b.Type {
		case textType:
			vm.Store("$t文本行", b.Text)
			sb.WriteString(vm.ExecCaption(TextCaption))
		case optType:
			if b.BeforeCondition != "" {
				if vm.ExecCond(b.BeforeCondition) == false {
					continue
				}
			}
			g.AddOpt(b)
			i++
			vm.Store("$t选项序号", strconv.FormatUint(i, 10))
			vm.Store("$t选项内容", b.NextTip)
			vm.Store("$t目标场景标题", b.NextTitle)
			sb.WriteString(vm.ExecCaption(OptCaption))
			sb.WriteString("\n")
		case innerCodeType:
			sb.WriteString(vm.Exec(b.InnerCode))
		case multiLineCodeType:
			vm.Exec(b.MultiLineCode)
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
	c.ParentId = s.ParentId
	s.children = append(s.children, c)
	s.ChildrenId = append(s.ChildrenId, c.Id)
}

func (s *Scene) AddBlock(b *Block) {
	s.Block = append(s.Block, b)
}
