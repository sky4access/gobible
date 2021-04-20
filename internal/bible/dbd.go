package bible

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/doug-martin/goqu.v4"
	"gopkg.in/yaml.v2"
)

type BibleDB struct {
	inputFile    string
	lang         string
	db           *goqu.Database
	Input        Input
	dbFile       string
	MemoryVerses []string
	Verses       []string
}

func NewBileDB(lang string, inputFile string) BibleDB {
	return BibleDB{
		lang:      lang,
		inputFile: inputFile,
	}
}

func (m *BibleDB) Init() {

	source, err := ioutil.ReadFile(m.inputFile)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(source, &m.Input)
	if err != nil {
		log.Fatal(err)
	}
	m.dbFile = ESV_FILE
	if m.lang == "kor" {
		m.dbFile = KRV_FILE
	}
	d, err := sql.Open("sqlite3", m.dbFile)
	if err != nil {
		log.Fatal(err)
	}

	m.db = goqu.New("sqllite3", d)
}

func toValidParam(v string) string {
	space := regexp.MustCompile(`\s+`)
	return strings.Replace(space.ReplaceAllString(v, " "), " ", "+", 1)
}

func (m *BibleDB) Fetch() {
	verses := make([]string, 0)
	for _, v := range m.Input.Memories {
		p := m.ParseVerses(toValidParam(v))
		verses = append(verses, p)
	}
	m.MemoryVerses = verses

	verses = make([]string, 0)

	for _, v := range m.Input.Verses {
		p := m.ParseVerses(toValidParam(v))
		verses = append(verses, p)
	}
	m.Verses = verses
}

func (m *BibleDB) Generate() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("#  %s\n", m.Input.Title))

	if len(m.MemoryVerses) > 0 {
		b.WriteString(fmt.Sprintln("\n## Memory Verses"))
		for _, v := range m.MemoryVerses {
			v = strings.Replace(v, "\n\n", " ", -1)
			b.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}
	b.WriteString(fmt.Sprintln("\n## Verses"))
	for _, v := range m.Verses {
		v = strings.Replace(v, "\n\n", " ", -1)
		b.WriteString(fmt.Sprintf("- %s\n", v))
	}

	return b.String()
}

func (m BibleDB) QueryBible(book string, chapter int, verses ...int) (string, []Entry, error) {
	book = strings.ToLower(book)
	ds := m.db.From("bible")
	var q *goqu.Dataset

	bookNumber := BooksByName[book]
	bookNameStr := BooksByNumber[bookNumber]
	bookNameArr := strings.Split(bookNameStr, ",")
	if len(bookNameArr) != 2 {
		log.Fatalf("there as problem gettin book %v from %v\n", book, m.inputFile)
	}
	bookName := ""
	if m.lang == "kor" {
		bookName = bookNameArr[0]
	} else {
		bookName = bookNameArr[1]
	}
	var entryName string
	if len(verses) == 1 {
		q = ds.Select(
			goqu.I("verse").As("verse"),
			goqu.I("content").As("content")).
			Where(goqu.I("book").Eq(BooksByName[book]),
				goqu.I("chapter").Eq(chapter),
				goqu.I("verse").Eq(verses[0]))
		entryName = fmt.Sprintf("%v %v:%v", bookName, chapter, verses[0])

	} else if len(verses) == 2 {
		q = ds.Select(
			goqu.I("verse").As("verse"),
			goqu.I("content").As("content")).
			Where(goqu.I("book").Eq(BooksByName[book]),
				goqu.I("chapter").Eq(chapter),
				goqu.I("verse").Gte(verses[0]),
				goqu.I("verse").Lte(verses[1]))
		entryName = fmt.Sprintf("%v %v:%v-%v\n", bookNameStr, chapter, verses[0], verses[1])
	} else {
		q = nil
	}

	if q != nil {
		var entries []Entry
		if err := q.ScanStructs(&entries); err != nil {
			return entryName, nil, err
		}
		return strings.TrimSpace(entryName), entries, nil
	} else {
		return entryName, nil, errors.New("verses must be 1 or 2 arguments")
	}
}

func (m *BibleDB) ParseVerses(s string) string {
	arr := strings.Split(s, "+")
	b, cv := arr[0], arr[1]
	arr = strings.Split(cv, ":")
	c, v := arr[0], arr[1]
	vs := strings.Split(v, "-")

	chapter, _ := strconv.Atoi(c)
	if len(vs) == 1 {
		v1, _ := strconv.Atoi(vs[0])
		entryName, entries, err := m.QueryBible(b, chapter, v1)
		if err != nil {
			log.Fatal(err)
		}
		return printVerse(entryName, entries)
	} else if len(vs) == 2 {
		v1, _ := strconv.Atoi(vs[0])
		v2, _ := strconv.Atoi(vs[1])
		entryName, entries, err := m.QueryBible(b, chapter, v1, v2)
		if err != nil {
			log.Fatal(err)
		}
		return printVerse(entryName, entries)
	}
	return ""
}

func printVerse(entryName string, entries []Entry) string {

	buffer := bytes.NewBufferString("")
	buffer.WriteString(entryName + " ")
	for _, entry := range entries {
		buffer.WriteString(fmt.Sprintf("(%v)%v", strconv.Itoa(int(entry.Verse)), entry.Content))
	}
	return buffer.String()
}
