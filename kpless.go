package kpless

import (
	"errors"
	"sync"
)

func New() *KPLess {
	return &KPLess{
		mu:    sync.RWMutex{},
		Games: make(map[string]*Game),
		Books: nil,
	}
}

type Caption int

const (
	NoFoundBook        Caption = iota // 未配置冒险书
	NoFoundJumpToScene                // 找不到要跳转的目标场景
	NoTopNoForkScene                  // 不是顶级场景又不是分支场景
	TitleCaption
	OptCaption
	TextCaption
)

type RollVM interface {
	Exec(text string) string
	ExecCond(text string) bool
	ExecCaption(i Caption) string
	Store(name, val string)
}

type KPLess struct {
	mu    sync.RWMutex
	Games map[string]*Game `json:"games"`
	Books []*Book          `json:"books"`
}

func (l *KPLess) SetGame(vm RollVM, id, name string) error {
	for _, book := range l.Books {
		if book.Name == name {
			l.mu.Lock()
			l.Games[id] = &Game{
				kp:   l,
				book: book,
			}
			l.mu.Unlock()
			return nil
		}
	}
	return errors.New(vm.ExecCaption(NoFoundBook))
}

func (l *KPLess) Input(vm RollVM, id, content string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if g, ok := l.Games[id]; ok {
		return g.Next(vm, content)
	}
	return "", errors.New("")
}

func (l *KPLess) LoadMarkDown(name string) error {
	p := newMDParser()
	err := p.loadFile(name)
	if err != nil {
		return err
	}
	p.scenes = append(p.scenes, p.cur)
	l.Books = append(l.Books, p.getBook())
	return nil
}
