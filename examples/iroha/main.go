package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mpppk/messagen/messagen"
)

func panicIfErrExist(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	opt := &messagen.Option{
		TemplatePickers: []messagen.TemplatePicker{messagen.RandomTemplatePicker, IrohaTemplatePicker},
	}
	generator, err := messagen.New(opt)
	panicIfErrExist(err)

	config, err := messagen.ParseYamlFile("examples/iroha/pokemon.yaml")
	panicIfErrExist(err)

	if err := generator.AddDefinition(config.Definitions...); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		msg, err := generator.Generate("Root", nil, 1)
		panicIfErrExist(err)
		fmt.Println(msg)
		wg.Done()
	}()

	go func() {
		for {
			fmt.Println("goroutine: ", runtime.NumGoroutine())
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}

func IrohaTemplatePicker(def *messagen.DefinitionWithAlias, state *messagen.State) (messagen.Templates, error) {
	var newTemplates messagen.Templates
	for _, template := range def.Templates {
		if !template.IsSatisfiedState(state) {
			newTemplates = append(newTemplates, template)
			continue
		}
		msg, err := template.Execute(state)
		if err != nil {
			return nil, err
		}
		if HasDuplicatedRune(NormalizeKatakanaWord(string(msg))) {
			continue
		}
		newTemplates = append(newTemplates, template)
	}
	return newTemplates, nil
}

func NormalizeKatakanaWord(word string) string {
	m := newNormalizeKatakanaMap()
	var runes []rune
	for _, w := range word {
		if newW, ok := m[w]; ok {
			runes = append(runes, newW)
			continue
		}
		runes = append(runes, w)
	}
	newWord := string(runes)
	return strings.Replace(newWord, "ー", "", -1)
}

func newNormalizeKatakanaMap() map[rune]rune {
	m := map[rune]rune{}
	m['ァ'] = 'ア'
	m['ィ'] = 'イ'
	m['ゥ'] = 'ウ'
	m['ェ'] = 'エ'
	m['ォ'] = 'オ'
	m['ッ'] = 'ツ'
	m['ャ'] = 'ヤ'
	m['ュ'] = 'ユ'
	m['ョ'] = 'ヨ'
	m['ガ'] = 'カ'
	m['ギ'] = 'キ'
	m['グ'] = 'ク'
	m['ゲ'] = 'ケ'
	m['ゴ'] = 'コ'
	m['ザ'] = 'サ'
	m['ジ'] = 'シ'
	m['ズ'] = 'ス'
	m['ゼ'] = 'セ'
	m['ゾ'] = 'ソ'
	m['ダ'] = 'タ'
	m['ヂ'] = 'チ'
	m['ヅ'] = 'ツ'
	m['デ'] = 'テ'
	m['ド'] = 'ト'
	m['バ'] = 'ハ'
	m['ビ'] = 'ヒ'
	m['ブ'] = 'フ'
	m['ベ'] = 'ヘ'
	m['ボ'] = 'ホ'
	m['パ'] = 'ハ'
	m['ピ'] = 'ヒ'
	m['プ'] = 'フ'
	m['ペ'] = 'ヘ'
	m['ポ'] = 'ホ'
	m['ヴ'] = 'ウ'
	return m
}

func HasDuplicatedRune(word string) bool {
	m := map[rune]struct{}{}
	for _, r := range word {
		if _, ok := m[r]; ok {
			return true
		}
		m[r] = struct{}{}
	}
	return false
}
