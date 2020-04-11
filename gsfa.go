package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/api"
)

// TaskURL はAtCoderの問題一覧が記載されているURLです。
const TaskURL string = "https://atcoder.jp/contests/%s/tasks"

func main() {
	Run(os.Args[1:])
}

// Run はgsfaを実行します。
// 引数が0要素の場合は使用例を表示して終了します。
// args[0]:contestName
func Run(args []string) {

	usage := `
gsfa --  Get a sample from AtCoder
  Get input/output example from AtCoder.
  The retrieved information is stored in "./atqdl/[contestName]".

Args
  1, contestName
     Please pass a string corresponding to contestName in "atcoder.jp/contacts/contestName".
`

	if len(args) < 1 {
		fmt.Println(usage)
		return
	}

	GetSamplesUseChromeDriver(args[0])
}

// GetSamplesUseChromeDriver は引数のコンテストの入出力例をchromedriverを利用して取得します。
func GetSamplesUseChromeDriver(contestName string) {

	d := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
		}),
	)

	GetSamples(d, contestName)
}

// GetSamples はdのWebDriverを利用してコンテストの入出力例を取得します。
func GetSamples(d *agouti.WebDriver, contestName string) {

	p, _ := os.Getwd()
	contestHome := filepath.Join(p, "gsfa", contestName)
	// ディレクトリが既に存在する場合、サンプルDL済みとみなす。
	if f, err := os.Stat(contestHome); !os.IsNotExist(err) && f.IsDir() {
		fmt.Println("Contest folder already exists below \"./gsfa\".")
		fmt.Println("The sample case has already been downloaded.")
		return
	}

	// 対象コンテストの問題URLを取得
	q := GetQuestionURLs(d, contestName)
	if len(q) < 1 {
		fmt.Println("Failed to get problem URL.")
		fmt.Printf("URL:%v", fmt.Sprintf(TaskURL, contestName))
		return
	}

	// サンプル入力/出力を取得
	samples := GetSampleCaces(d, q)
	if len(samples) < 1 {
		fmt.Println("Failed to get I/O example.")
		return
	}

	// サンプル保存用のディレクトリ生成
	os.MkdirAll(contestHome, 0777)
	// 所得した入力/出力をファイル化して保存
	if CreateSampleFiles(contestHome, samples) {
		fmt.Println("Getting I/O examples succeeded!!")
		fmt.Printf("The output file is saved in \"%v\"", contestHome)
	} else {
		fmt.Println("There was a problem saving the sample case.")
	}
}

// GetQuestionURLs は対象のコンテストの各問題のURLを取得し、
// map[問題名]URLの形で返却します。
func GetQuestionURLs(d *agouti.WebDriver, contestName string) map[string]string {

	if err := d.Start(); err != nil {
		log.Fatalf("Failed to start driver : %v", err)
	}
	defer d.Stop()
	page, _ := d.NewPage()
	defer page.CloseWindow()

	tasks := fmt.Sprintf(TaskURL, contestName)
	if err := page.Navigate(tasks); err != nil {
		fmt.Println("Page transition failed.\nThe URL may be incorrect.")
		fmt.Println("url :", TaskURL)
		log.Fatalf("error : %v", err)
	}

	sel := "#main-container > div.row > div:nth-child(2) > div > table > tbody"
	tbody := page.Find(sel)

	// key:Question(A-Z), value:Question URL
	questions := make(map[string]string)
	// A to Z
	for i := 0x41; i <= 0x5a; i++ {
		a := tbody.FirstByLink(string(i))
		q, e1 := a.Text()
		href, e2 := a.Attribute("href")
		if e1 != nil || e2 != nil {
			break
		}

		questions[q] = href
	}

	return questions
}

// GetSampleCaces は引数で渡されたURLから入出力例を取得します。
func GetSampleCaces(d *agouti.WebDriver, q map[string]string) map[string][]*SampleCase {

	if err := d.Start(); err != nil {
		log.Fatalf("Failed to start driver : %v", err)
	}
	defer d.Stop()

	page, _ := d.NewPage()
	defer page.CloseWindow()

	sc := make(map[string][]*SampleCase, len(q))
	format := "pre-sample%v"
	for k, v := range q {
		page.Navigate(v)
		// ページ描画のため1秒待つ
		time.Sleep(1000)

		sc[k] = make([]*SampleCase, 0)
		idx := 1
		for i, j := 0, 1; ; i, j = (i + 2), (j + 2) {
			ie, _ := page.FindByID(fmt.Sprintf(format, i)).Elements()
			oe, _ := page.FindByID(fmt.Sprintf(format, j)).Elements()

			in := exText(ie)
			out := exText(oe)
			if in == "" && out == "" {
				break
			}

			s := &SampleCase{
				Question: k,
				No:       idx,
				In:       in,
				Out:      out,
			}

			sc[k] = append(sc[k], s)
			idx++
		}
	}

	return sc
}

func exText(e []*api.Element) string {

	if e == nil {
		return ""
	}

	li := len(e) - 1
	r := bytes.NewBuffer(make([]byte, 0, 100))
	for i := 0; i <= li; i++ {
		t, err := e[i].GetText()
		if err != nil {
			break
		}
		r.Write([]byte(t))

		if i != li {
			r.Write([]byte(`\n`))
		}
	}

	return r.String()
}

// CreateSampleFiles は取得した入出力例をファイル化し保存します。
func CreateSampleFiles(dir string, sc map[string][]*SampleCase) (ok bool) {

	create := func(dir, fname, write string) error {

		f, err := os.Create(filepath.Join(dir, fname))
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(write); err != nil {
			return err
		}

		return nil
	}

	ok = true
	for k, v := range sc {
		qdir := filepath.Join(dir, k)
		os.MkdirAll(qdir, 0777)

		for _, s := range v {
			inFile := fmt.Sprintf("%v_%v_in.txt", k, s.No)
			if err := create(qdir, inFile, s.In); err != nil {
				ok = false
				fmt.Printf("File create error. [file : %v]\n%v", inFile, err)
			}
			outFile := fmt.Sprintf("%v_%v_out.txt", k, s.No)
			create(qdir, outFile, s.Out)
			if err := create(qdir, inFile, s.In); err != nil {
				ok = false
				fmt.Printf("File create error. [file : %v]\n%v", outFile, err)
			}
		}
	}

	return ok
}

// SampleCase は入出力サンプルケースを表します。
type SampleCase struct {
	Question string // 問題番号(A,B,C...)
	No       int    // 入出力例番号
	In       string // 入力内容
	Out      string // 出力内容
}
