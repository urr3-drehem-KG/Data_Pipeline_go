package CLDI_Extractor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

//Lines map to a line_no, transliterations, normalizations and annotations
type CLDIData struct {
	CLDI       string
	PUB        string
	TabletList []TabletLine
}

type TabletLine struct {
	TabletLocation  string
	TabletLines     map[string]string //map line_no to transliterations
	NormalizedLines map[string]string
	EntitiyLines    map[string]string
	// Annotation      map[string]string //map line_no to annotation
}

type ATFParser struct {
	path         string
	data         []string
	currCLDIData CLDIData
	out          chan CLDIData
	done         chan struct{}
	t            int
}

func newATFParser(path string) *ATFParser {
	atfParser := &ATFParser{
		path: path,
		out:  make(chan CLDIData, 1000),
		done: make(chan struct{}, 1),
	}
	atfParser.loadCLDIData()
	atfParser.run()
	return atfParser
}

func (p *ATFParser) run() {

	go func() {
		for _, line := range p.data {
			p.parseLines(line)
		}
		p.out <- p.currCLDIData
		close(p.out)
	}()

	go func() {
		p.done <- struct{}{}
	}()
}

func (p *ATFParser) WaitUntilDone() {
	p.done <- struct{}{}
}

func (p *ATFParser) loadCLDIData() {
	f, err := os.Open(p.path)
	if err != nil {
		log.Fatalf("failed reading file: %s", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.data = append(p.data, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("issue reading this file: %s", err)
	}
}

/*
Parse Tablet Line-By-Line.
	- &P indicates a new tablet, we initalize p.currCLDIData for a new tablet
	- Store CLDI, PUB, annotation and transliterations data to this object
*/

func (p *ATFParser) parseLines(line string) {
	line = strings.TrimSpace(line)
	if line != "" && strings.Contains(line, "&P") && strings.Contains(line, " =") {
		if p.currCLDIData.PUB == "" {
			line = strings.ReplaceAll(line, "&", "")
			line = strings.TrimSpace(line)
			data := strings.SplitN(line, " = ", 2)
			p.currCLDIData.CLDI = data[0]
			p.currCLDIData.PUB = data[1]

		} else {
			p.out <- p.currCLDIData
			p.currCLDIData = CLDIData{}
			p.t++

		}
		line = strings.ReplaceAll(line, "&", "")
		line = strings.TrimSpace(line)
		data := strings.SplitN(line, " = ", 2)
		p.currCLDIData.CLDI = data[0]
		p.currCLDIData.PUB = data[1]

	} else if line != "" && strings.Contains(string(line[0:1]), "@") && !strings.Contains(line, "object") && !strings.Contains(line, "tablet") && !strings.Contains(line, "envelope") && !strings.Contains(line, "bulla") {
		newTableLine := TabletLine{}
		newTableLine.TabletLocation = strings.TrimSpace(line)
		newTableLine.TabletLines = make(map[string]string)
		p.currCLDIData.TabletList = append(p.currCLDIData.TabletList, newTableLine)

	} else if strings.Contains(line, "#tr.en") {
		// You can translate tr.en entries
		line = strings.Replace(line, "#tr.en", "", 1)
		line = strings.Replace(line, ":", "", 1)

		// TabletLine.
		// p.currCLDIData.tabletLines["annotations"] = strings.TrimSpace(line)

	} else if !strings.Contains(line, "$") && strings.Contains(line, ". ") && !strings.Contains(string(line[0:1]), "#") {
		_, err := strconv.Atoi(line[0:1])
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		data := strings.SplitN(line, ". ", 2)

		line_no := strings.TrimSpace(data[0])
		translit := strings.TrimSpace(data[1])
		if len(p.currCLDIData.TabletList) > 0 {
			p.currCLDIData.TabletList[len(p.currCLDIData.TabletList)-1].TabletLines[line_no] = translit
		}
	}
}
