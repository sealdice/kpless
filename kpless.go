package kpless

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/Masterminds/semver/v3"
)

type Logger interface {
	Infof(format string, args ...any)
}

var log Logger

func New(logger Logger) *KPLess {
	log = logger
	return &KPLess{
		mu:    sync.RWMutex{},
		Games: make(map[string]*Game),
		Books: make(map[string]*Book),
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
	ModBreakUpdate
	LoadGameNoFoundBookT
)

var (
	ErrModBreakUpdate      = errors.New("游玩的模组发生了破坏性更新")
	ErrLoadGameNoFoundBook = errors.New("无法找到曾经游玩的模组")
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
	Books map[string]*Book `json:"books"`
}

func (l *KPLess) Input(vm RollVM, id, content string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if g, ok := l.Games[id]; ok {
		return g.Next(vm, content)
	}
	return vm.ExecCaption(NoFoundBook), nil
}

const (
	MetaName    = "name"
	metaVersion = "version"
)

var mustMeta = []string{
	MetaName,
	metaVersion,
}

func (l *KPLess) LoadMarkDownBook(name string) error {
	p := newMDParser()
	err := p.loadFile(name)
	if err != nil {
		return err
	}
	p.scenes = append(p.scenes, p.cur)
	book := p.getBook()
	for _, s := range mustMeta {
		if _, ok := book.Meta[s]; !ok {
			return fmt.Errorf("必须提供元属性：%s", s)
		}
	}
	book.Name = book.Meta[MetaName]
	l.Books[book.Name] = book
	return nil
}

func (l *KPLess) SetGame(vm RollVM, id, name string) error {
	if book, ok := l.Books[name]; ok {
		l.mu.Lock()
		l.Games[id] = &Game{
			book:     book,
			BookName: book.Name,
		}
		l.mu.Unlock()
		return nil
	}
	return errors.New(vm.ExecCaption(NoFoundBook))
}

func (l *KPLess) LoadGames(b []byte) error {
	err := json.Unmarshal(b, &l.Games)
	if err != nil {
		return err
	}
	for _, game := range l.Games {
		book, ok := l.Books[game.BookName]
		if !ok {
			game.err = ErrLoadGameNoFoundBook
			continue
		}
		constraint, err := semver.NewConstraint("^" + game.BookVersion)
		if err != nil {
			return err
		}
		bv, err := semver.NewVersion(book.Meta[metaVersion])
		if err != nil {
			return err
		}
		if !constraint.Check(bv) {
			game.err = ErrModBreakUpdate
		}
	}
	return nil
}

func (l *KPLess) SaveGames() ([]byte, error) {
	return json.Marshal(&l.Games)
}
