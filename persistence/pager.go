package persistence

import (
	"fmt"
	"os"
)

// TODO:
const ROW_SIZE = 296

const (
	PAGE_SIZE       = 4096
	TABLE_MAX_PAGES = 100
	ROWS_PER_PAGE   = PAGE_SIZE / ROW_SIZE
	TABLE_MAX_ROWS  = ROWS_PER_PAGE * TABLE_MAX_PAGES
)

type Page [PAGE_SIZE]byte

type Pager struct {
	file     *os.File
	pages    [TABLE_MAX_PAGES](*Page)
	numPages uint32
}

// ページャを初期化する。DBファイルのサイズからページ数を計算して、設定する。
func InitPager(name string) (*Pager, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// ページ数を計算する。ファイルサイズとページサイズから計算できる。
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	numPages := fi.Size() / PAGE_SIZE

	if fi.Size()%PAGE_SIZE != 0 {
		fmt.Printf("Db file is not a whole number of pages. Corrupt file.\n")
		os.Exit(1)
	}

	// ページャーを初期化する
	pager := Pager{
		file:     f,
		pages:    [TABLE_MAX_PAGES]*Page{},
		numPages: uint32(numPages),
	}

	// ページ0を初期化する
	if pager.numPages == 0 {
		node := pager.GetPage(0)
		initLeafNode(node)
	}

	return &pager, err
}

// ページを取得する。ページがキャッシュされていない場合は、ファイルから読み取ってキャッシュする。
func (pager *Pager) GetPage(pageNum uint32) *Page {
	if pager.pages[pageNum] != nil {
		return pager.pages[pageNum]
	}

	// ファイルから読み取って、ページャに設定する
	pager.file.Seek(int64(pageNum*PAGE_SIZE), 0)
	page := Page{}
	pager.file.Read(page[:])
	pager.pages[pageNum] = &page

	// ここでページ数を増やすのちょっと変
	if uint32(pageNum) >= pager.numPages {
		pager.numPages = uint32(pageNum) + 1
	}

	return pager.pages[pageNum]
}

// 使っていない新しいページを取得する
func (pager *Pager) GetNewPage() (*Page, uint32) {
	newPageNum := pager.numPages
	pager.numPages++

	return pager.GetPage(newPageNum), newPageNum
}

// ページャの内容をディスクに書き込む
func (pager *Pager) FlushPages() error {
	defer pager.file.Close()

	file := pager.file

	// 全ての行が埋まっているページの書き込み
	for i := uint32(0); i < pager.numPages; i++ {
		// ページがキャッシュにない場合は何もしない（読み取りも書き込みもされていない場合）
		page := pager.pages[i]
		if page == nil {
			continue
		}

		file.Seek(int64(PAGE_SIZE*i), 0)
		_, err := file.Write(page[:])
		if err != nil {
			return err
		}
	}

	return nil
}
