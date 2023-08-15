package persistence

import (
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
	file  *os.File
	pages [TABLE_MAX_PAGES](*Page)
}

func InitPager(name string) (*Pager, uint32, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, 0, err
	}

	// ページャーを初期化する
	pager := Pager{
		file:  f,
		pages: [TABLE_MAX_PAGES]*Page{},
	}

	// 行数を計算する。ファイルサイズと1行のサイズから計算できる。
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	numRows := uint32(fi.Size() / int64(ROW_SIZE))

	return &pager, numRows, err
}

func (pager *Pager) getPage(pageNum int) *Page {
	if pager.pages[pageNum] != nil {
		return pager.pages[pageNum]
	}

	// ディスクから読み取って、ページャに設定する
	pager.file.Seek(int64(pageNum*PAGE_SIZE), 0)
	page := Page{}
	pager.file.Read(page[:])
	pager.pages[pageNum] = &page

	return pager.pages[pageNum]
}

// ページの内容をディスクに書き込む
func (pager *Pager) FlushPages(numRows uint32) error {
	defer pager.file.Close()

	file := pager.file

	// 全ての行が埋まっているページの書き込み
	numFullPages := numRows / uint32(ROWS_PER_PAGE)
	for i := uint32(0); i < numFullPages; i++ {
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

	// 埋まっていないページの書き込み
	numAdditionalRows := numRows % uint32(ROWS_PER_PAGE)
	if numAdditionalRows > 0 {
		if page := pager.pages[numFullPages]; page != nil {
			file.Seek(int64(PAGE_SIZE*numFullPages), 0)

			_, err := file.Write(page[0 : int(numAdditionalRows)*ROW_SIZE])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// メモリ上の行の位置
type RowSlot struct {
	pageNum  int
	rowStart int
	rowEnd   int
}

func (pager *Pager) InsertRow(rs RowSlot, row []byte) {
	page := pager.getPage(rs.pageNum)
	copy(page[rs.rowStart:rs.rowEnd], row)
}

func (pager *Pager) GetRow(rs RowSlot) []byte {
	page := pager.getPage(rs.pageNum)
	return page[rs.rowStart:rs.rowEnd]
}

func GetRowSlot(rowNumUint32 uint32) RowSlot {
	rowNum := int(rowNumUint32)
	pageNum := int(rowNum) / ROWS_PER_PAGE
	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * ROW_SIZE

	return RowSlot{pageNum: pageNum, rowStart: byteOffset, rowEnd: byteOffset + ROW_SIZE}
}
