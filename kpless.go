package kpless

import (
	"bufio"
	"errors"
	"os"
	"sync"
)

func New() *KPLess {
	captionFun := func(i Caption) string {
		switch i {
		case NoBookFound:
			return "未配置冒险书，无法开始游戏"
		}
		return "此处应该有文案 但是没有，请联系开发者"
	}
	return &KPLess{
		mu:         sync.RWMutex{},
		games:      make(map[string]*Game),
		books:      nil,
		GetCaption: captionFun,
	}
}

type Caption int

const (
	NoBookFound Caption = iota // 未配置冒险书，无法开始游戏
)

type KPLess struct {
	mu         sync.RWMutex
	games      map[string]*Game
	books      []*Book
	GetCaption func(i Caption) string
}

func (l *KPLess) SetGame(id, name string) error {
	for _, book := range l.books {
		if book.Name == name {
			l.mu.Lock()
			l.games[id] = &Game{
				kp:   l,
				book: book,
			}
			l.mu.Unlock()
			return nil
		}
	}
	return errors.New(l.GetCaption(NoBookFound))
}

func (l *KPLess) Input(id, content string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if g, ok := l.games[id]; ok {
		return g.Next(content)
	}
	return "", errors.New("")
}

func (l *KPLess) LoadMarkDown(name string) error {
	p := mdParser{}
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	s := bufio.NewScanner(file)
	for s.Scan() {
		err := p.parseLine(s.Bytes())
		if err != nil {
			return err
		}
	}
	p.scenes = append(p.scenes, p.cur)
	b := &Book{SceneList: p.scenes, Name: name}
	l.books = append(l.books, b)
	return nil
}
