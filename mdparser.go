package kpless

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type mdParser struct {
	lineCount int
	pageCount int
	cur       *Scene
	book      Book
	scenes    []*Scene
}

func (p *mdParser) createScene(title string, nextLevel int) error {
	p.pageCount++
	next := &Scene{Id: p.pageCount, Title: title, Level: nextLevel}

	if p.cur == nil {
		p.cur = next
		return nil
	}
	if nextLevel == 1 {
		p.scenes = append(p.scenes, p.cur)
		p.cur = next
		return nil
	}

	switch r := p.cur.Level - nextLevel; {
	// (h2 >> h3 >>) h4 >> h3
	case r < 0:
		p.cur.Find(nextLevel - 1).AddChild(next)
	// h2 >> h2
	case r == 0:
		p.cur.parent.AddChild(next)
	// h2 >> h3
	case r == 1:
		p.cur.AddChild(next)
	// h2 >> h5
	// 嵌套本来就很反人类，还是加点限制吧
	default:
		return fmt.Errorf("层级不连续递增：%s (line:%d)", title, p.lineCount)
	}
	p.cur = next
	return nil
}

// * {cond} [tip](#title) {after}
var optRegexp = regexp.MustCompile(`^\*\s+({[\s\S]+})?\s*\[(.+)]\(#(.+)\)\s*({[\s\S]+})?`)

func (p *mdParser) addNode(line string) error {
	ls := optRegexp.FindStringSubmatch(line)
	if len(ls) != 0 {
		if ls[3] == p.cur.Title && ls[1] == "" {
			return fmt.Errorf("跳转到自己，且没有跳出条件，形成死循环 %d", p.lineCount)
		}
		r := strings.NewReplacer("{", "", "}", "")
		p.cur.AddBlock(&Opt{
			BeforeCondition: r.Replace(ls[1]),
			Content:         ls[2],
			NextTitle:       ls[3],
			AfterEval:       r.Replace(ls[4]),
		})
		return nil
	}
	// A { B1 { B2 } B3 { B4 } B5 } C1 { D1 } E1
	count := 0
	i := 0
	for {
		if i >= len(line) {
			if line != "" {
				p.cur.AddBlock(&line)
			}
			break
		}
		if line[i:i+1] == "{" {
			if count == 0 {
				l := line[:i]
				p.cur.AddBlock(&l)
				line = line[i+1:]
				i = 0
			}
			count++
		}
		if line[i:i+1] == "}" {
			count--
			if count == 0 {
				p.cur.AddBlock(&Code{line[:i]})
				line = line[i+1:]
				i = 0
			}
		}
		i++
	}
	return nil
}

var sceneRegexp = regexp.MustCompile(`^#+\s+`)

func (p *mdParser) parseLine(line []byte) error {
	p.lineCount++
	if len(line) == 0 {
		return nil
	}
	i := sceneRegexp.FindIndex(line)
	if len(i) == 0 {
		err := p.addNode(string(line))
		return err
	}
	level := bytes.Count(line[:i[1]], []byte("#"))
	title := string(line[i[1]:])
	err := p.createScene(title, level)
	if err != nil {
		return err
	}
	return nil
}
