package kpless

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
)

func newMDParser() *mdParser {
	return &mdParser{meta: map[string]string{}}
}

type mdParser struct {
	isMeta    bool
	isCodes   bool
	meta      map[string]string
	lineCount int
	pageCount int
	cur       *Scene
	book      Book
	scenes    []*Scene
	codes     []byte
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

// * cond [tip](#title) after
var optRegexp = regexp.MustCompile(`^\*\s+([\s\S]+)?\s*\[(.+)]\(#(.+)\)\s*([\s\S]+)?`)

func (p *mdParser) addNode(line string) error {
	ls := optRegexp.FindStringSubmatch(line)
	if len(ls) != 0 {
		if ls[3] == p.cur.Title && ls[1] == "" {
			return fmt.Errorf("跳转到自己，且没有跳出条件，形成死循环 %d", p.lineCount)
		}
		p.cur.AddBlock(&Opt{
			BeforeCondition: ls[1],
			Content:         ls[2],
			NextTitle:       ls[3],
			AfterEval:       ls[4],
		})
		return nil
	}
	// A { B1 { B2 } B3 { B4 } B5 } C1 { D1 } E1
	count := 0
	i := 0
	for {
		if i >= len(line) {
			line += "\n"
			p.cur.AddBlock(&line)
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

var (
	sceneRegexp  = regexp.MustCompile(`^#+\s+`)
	metaFlag     = []byte("---")
	metaKVRegexp = regexp.MustCompile(`^(.+):\s*(.+)`)
	codesFlag    = []byte("```")
)

func (p *mdParser) parseLine(line []byte) error {
	p.lineCount++
	if len(line) == 0 {
		return nil
	}

	if p.lineCount == 1 {
		if bytes.Equal(line, metaFlag) {
			p.isMeta = true
			return nil
		}
	}
	if p.isMeta {
		if bytes.Equal(line, metaFlag) {
			p.isMeta = false
			return nil
		}
		res := metaKVRegexp.FindSubmatch(line)
		if len(res[1]) == 0 {
			return errors.New("无法识别的元属性")
		}
		p.meta[string(res[1])] = string(res[2])
		return nil
	}

	if p.isCodes {
		if bytes.Equal(line, codesFlag) {
			p.isCodes = false
			p.cur.AddBlock(&Codes{val: string(p.codes)})
			return nil
		}
		line2 := bytes.Clone(line)
		p.codes = append(p.codes, line2...)
		p.codes = append(p.codes, []byte("\n")...)
		return nil
	}
	if bytes.Equal(line, codesFlag) {
		p.isCodes = true
		return nil
	}

	i := sceneRegexp.FindIndex(line)
	if len(i) == 0 {
		err := p.addNode(string(line))
		return err
	}
	level := bytes.Count(line[:i[1]], []byte("#"))
	title := string(line[i[1]:]) // 将缓冲区的 byte 转为永久的字符串
	err := p.createScene(title, level)
	if err != nil {
		return err
	}
	return nil
}
