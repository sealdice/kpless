package kpless

import (
	"bytes"
	"fmt"
	"regexp"
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

// * {cond} [tip](#title)
var optRegexp = regexp.MustCompile(`^\*\s+(\{[\s\S]+})?\s*\[(.+)]\(#(.+)\)\s*(\{[\s\S]+})?`)

func (p *mdParser) addNode(line string) error {
	ls := optRegexp.FindStringSubmatch(line)
	if len(ls) != 0 {
		if ls[3] == p.cur.Title {
			return fmt.Errorf("跳转到自己死循环 %d", p.lineCount)
		}
		p.cur.AddBlock(&Opt{
			BeforeCondition: ls[1],
			NextTip:         ls[2],
			NextTitle:       ls[3],
		})
		return nil
	}
	p.cur.AddBlock(&Text{line})
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
		p.addNode(string(line)) // TODO: 完成条件判断后记得处理错误
		return nil
		// return err
	}
	level := bytes.Count(line[:i[1]], []byte("#"))
	title := string(line[i[1]:])
	err := p.createScene(title, level)
	if err != nil {
		return err
	}
	return nil
}
