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
		pager.getPage(0)
	}

	return &pager, err
}

func (pager *Pager) getPage(pageNum uint32) *Page {
	if pager.pages[pageNum] != nil {
		return pager.pages[pageNum]
	}

	// ディスクから読み取って、ページャに設定する
	pager.file.Seek(int64(pageNum*PAGE_SIZE), 0)
	page := Page{}
	pager.file.Read(page[:])
	pager.pages[pageNum] = &page

	// ここでページ数を増やすのちょっと変
	if uint32(pageNum) >= pager.numPages {
		pager.numPages = uint32(pageNum)
	}

	return pager.pages[pageNum]
}

// ページのセルの数を取得する
func GetNumCells(pager *Pager, pageNum uint32) uint32 {
	page := pager.pages[pageNum]

	if page == nil {
		return 0
	}
	return getLeafNodeNumCells(page)
}

// ページのセルの数を増やす
func IncrementNumCells(pager *Pager, pageNum uint32) {
	newNumCells := GetNumCells(pager, pageNum) + 1
	page := pager.pages[pageNum]
	writeLeafNodeNumCells(page, newNumCells)
}

func WriteLeafNodeKey(pager *Pager, pageNum uint32, key uint32) {
	page := pager.pages[pageNum]
	writeLeafNodeKey(page, pageNum, key)
}

func WriteLeafNodeValue(pager *Pager, pageNum uint32, value []byte) {
	page := pager.pages[pageNum]
	writeLeafNodeValue(page, pageNum, value)
}

// ページの内容をディスクに書き込む
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

// メモリ上の行の位置
type RowSlot struct {
	pageNum  uint32
	rowStart uint32
	rowEnd   uint32
}

func (pager *Pager) InsertRow(rs RowSlot, row []byte) {
	page := pager.getPage(rs.pageNum)
	copy(page[rs.rowStart:rs.rowEnd], row)
}

func (pager *Pager) GetRow(rs RowSlot) []byte {
	page := pager.getPage(rs.pageNum)
	return page[rs.rowStart:rs.rowEnd]
}

// ページ上での位置を返す
func GetRowSlot(pageNum uint32, cellNum uint32) RowSlot {
	start, end := leafNodeCell(cellNum)

	return RowSlot{pageNum: pageNum, rowStart: start, rowEnd: end}
}
