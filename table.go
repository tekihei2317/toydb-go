package main

import (
	"os"
	"unsafe"
)

const (
	PAGE_SIZE       = 4096
	TABLE_MAX_PAGES = 100
	ROWS_PER_PAGE   = PAGE_SIZE / ROW_SIZE
	TABLE_MAX_ROWS  = ROWS_PER_PAGE * TABLE_MAX_PAGES
)

const (
	ID_OFFSET       = 0
	USERNAME_OFFSET = ID_OFFSET + int(unsafe.Sizeof(int(0)))
	EMAIL_OFFSET    = USERNAME_OFFSET + int(unsafe.Sizeof([32]byte{}))
	ID_SIZE         = int(unsafe.Sizeof(int(0)))
	USERNAME_SIZE   = int(unsafe.Sizeof([32]byte{}))
	EMAIL_SIZE      = int(unsafe.Sizeof([256]byte{}))
	ROW_SIZE        = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE
)

type Page [PAGE_SIZE]byte

type Pager struct {
	file  *os.File
	pages [TABLE_MAX_PAGES](*Page)
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
func (pager *Pager) flushPages(numRows uint32) error {
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

func (pager *Pager) insertRow(rs RowSlot, row []byte) {
	page := pager.getPage(rs.pageNum)
	copy(page[rs.rowStart:rs.rowEnd], row)
}

func (pager *Pager) getRow(rs RowSlot) []byte {
	page := pager.getPage(rs.pageNum)
	return page[rs.rowStart:rs.rowEnd]
}

type Table struct {
	numRows uint32
	pages   [TABLE_MAX_PAGES]Page
	pager   Pager
}

// メモリ上の行の位置
type RowSlot struct {
	pageNum  int
	rowStart int
	rowEnd   int
}

// 行番号から、行のスロットを取得する
func (table *Table) getRowSlot(rowNumUint32 uint32) RowSlot {
	rowNum := int(rowNumUint32)
	pageNum := int(rowNum) / ROWS_PER_PAGE
	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * ROW_SIZE

	return RowSlot{pageNum: pageNum, rowStart: byteOffset, rowEnd: byteOffset + ROW_SIZE}
}

// 行を挿入する
func (table *Table) insertRow(row *Row) {
	rs := table.getRowSlot(table.numRows)
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))

	// ページに書き込む
	table.pager.insertRow(rs, sourceBytes[:])
	table.numRows++
}

// 行を取得する
func (table *Table) getRowByRowNum(rowNum uint32) Row {
	rs := table.getRowSlot(rowNum)

	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	copy(destinationBytes[:], table.pager.getRow(rs))

	return *row
}
