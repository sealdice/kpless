package kpless

import (
	"bufio"
	"errors"
	"os"
	"sync"
)

func New() *KPLess {
	return &KPLess{
		mu:    sync.RWMutex{},
		games: make(map[string]*Game),
		books: nil,
	}
}

type KPLess struct {
	mu    sync.RWMutex
	games map[string]*Game
	books []*Book
}

func (l *KPLess) SetGame(id, name string) error {
	for _, book := range l.books {
		if book.Name == name {
			l.mu.Lock()
			l.games[id] = &Game{
				book: book,
			}
			l.mu.Unlock()
			return nil
		}
	}
	return errors.New("找不到冒险书")
}

func (l *KPLess) Input(id, content string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if g, ok := l.games[id]; ok {
		return g.Next(content)
	}
	return "", errors.New("未配置冒险书，无法开始游戏书")
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
	b := &Book{Scenes: p.scenes, Name: name}
	l.books = append(l.books, b)
	return nil
}
