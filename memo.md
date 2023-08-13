# memo

## Chapter 1

SQLが実行されるまで、

トークナイザー→パーサー→コードジェネレーター→

バーチャルマシン→B-tree→pager→OS interface

のようにたどっていく。SQLはSQLite用のバイトコードに変換され、バーチャルマシンで実行される。前半のコンポーネントをフロントエンド、後半をバックエンドという。

入力を受け取るために`fmt.Scanln`を使ったが、この関数は1行を空白区切りで読み取る。1行を読み取るためには、`bufio.Scanne.Scan()`を使えばよい。

構造体のポインタの場合、ドットアクセスする時に*での実体参照を省略できる。

## Chapter 2

文字列が特定の文字列で始まるかどうかは、strings.HasPrefixで判定できる。string.HasSuffixもある。

## Chapter 3

まずは、ハードコードした1つのテーブルだけで、append onlyのインメモリデータベースを作成する。

fmt.Sscanfで文字列をパースする場合、パースに失敗してもエラーにはならない。マッチしなかったプレースホルダ（`%d`、`%s`など）に対応する変数に値が代入されないだけ。

```go
var num1, num2 int
var str1, str2 string
c, _ := fmt.Sscanf("select 1 str", "select %d %s", &num1, &str1)
c2, _ := fmt.Sscanf("select 1", "select %d %s", &num2, &str2)

if c != 2 || num1 != 1 || str1 != "str" {
	t.Error("Parse failed.")
}

// str2は対応する値がなかったので、ゼロ値のまま
if c2 != 1 || num2 != 1 || str2 != "" {
	t.Error("Parse failed.")
}
```

### データの表現方法について

行、ページなどのデータをどのように表現にするかに悩んだ。改善の余地はかなりありそうなので、最初になんとか書いたコードについてメモする。

まず、テーブルの行はプログラム上では次の構造体で表現している。テーブルは1つで定義も決まっているため、とりあえずハードコーディングできる。

```go
type Row struct {
	id       int
	username [32]byte
	email    [256]byte
}
```

stringではなくバイトの配列を使っているのは、メモリに書き込むときに`string`からバイトの配列に変換できなかったからだった気がする。今思えば、`[]byte(str)`のように、バイトのスライスにキャストすることはできる。

```go
str := "abc"

expected := []byte{97, 98, 99}
if !reflect.DeepEqual([]byte(str), expected) {
  t.Errorf("expected %+v, got %+v\n", expected, []byte(str))
}
```

スライス同士の比較には`reflect.DeepEqual`が使える。配列とスライスを比較するとfalseになることに注意（`arr[:]`としてスライスに変換すればOK）。

[2つのsliceがすべて同じ要素をもっているかどうかを比較するショートカット - Qiita](https://qiita.com/taksatou@github/items/48b22d6d37e99a6c21cc)


注意点は、`[32]byte`だと32文字の文字を格納できるわけではない。Goの文字列はUTF-8なので、アルファベットは1バイトだが、ひらがなや絵文字は3~4バイトになる。なので`[]byte`にするべきだったかもしれない。UTF-8とかバイトとかそのものもよくわかっていなかったので、次の記事が参考になった。

[Goのruneを理解するためのUnicode知識 - Qiita](https://qiita.com/seihmd/items/4a878e7fa340d7963fee)

---

テーブルとページは次のように実装した。ページは行を保存する場所で、テーブルはページの集まり。

```go
type Table struct {
	numRows uint32
	pages   [TABLE_MAX_PAGES]unsafe.Pointer
}
```

[Go言語: ポインターとそれに関する型(uintptr, unsafe.Pointer) - Qiita](https://qiita.com/nozmiz/items/291f16f619a939bd7b87)

`unsafe.Pointer`は、任意の型のポインタを表せる型。任意の型のポインタを、`unsafe.Pointer`に変換できる。`unsafe.Pointer`はなんか危なそうなのであまり使わない方がいいのかも。自分でメモリの確保する必要があった（解放も？）。バイト列を保存するのであれば、代わりに`[]byte`を使えるはず。

書き込みの処理は次のような感じになった。rowSlotで次に書き込む位置のポインタを取得して、serializeRowで書き込んでいる。

```go
func executeInsert(statement Statement, table *Table) ExecuteResult {
	if table.numRows >= TABLE_MAX_ROWS {
		return EXECUTE_TABLE_FULL
	}

	rowToInsert := &statement.RowToInsert
	serializeRow(rowToInsert, rowSlot(table, table.numRows))
	table.numRows += 1

	return EXECUTE_SUCCESS
}
```

```go
// テーブルの行番号から、その位置を取得する
func rowSlot(table *Table, rowNum uint32) unsafe.Pointer {
	pageNum := rowNum / ROWS_PER_PAGE
	page := table.pages[pageNum]

	if page == nil {
		pageBytes := make([]byte, PAGE_SIZE)
		page = unsafe.Pointer(&pageBytes)
		table.pages[pageNum] = page
	}

	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * uint32(ROW_SIZE)

	return unsafe.Pointer(uintptr(page) + uintptr(byteOffset))
}
```

```go
// 行をメモリに書き込む
func serializeRow(source *Row, destination unsafe.Pointer) {
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(destination)
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(source))

	copy(destinationBytes[ID_OFFSET:], sourceBytes[ID_OFFSET:])
	copy(destinationBytes[USERNAME_OFFSET:], sourceBytes[USERNAME_OFFSET:])
	copy(destinationBytes[EMAIL_OFFSET:], sourceBytes[EMAIL_OFFSET:])
}
```

rowSlotでは、ページが進んだときにメモリを確保するようにしている。バイトのスライスを作って、そのポインタをページにセットしている。この処理を行わないとページの初期値がnilになるので、次のエラーが発生する。

```text
panic: runtime error: invalid memory address or nil pointer dereference
```

コードを振り返って見て、やっぱりページは`unsafe.Pointer`ではなく`[PAGE_SIZE]byte`でよさそうだ。今のコードでは、メモリを確保するときも書き込む時も、結局バイトのスライス（へのポインタ）に変換している。

## Chapter4

まずはビルドしたスクリプトを実行して、コマンドを入力できるようにする。ビルドは`go build`でできた。

実行するとパニックしてしまった。

```text
$ ./toydb-go < command
db > select
db > panic: runtime error: index out of range [0] with length 0

goroutine 1 [running]:
main.main()
        /Users/tekihei2317/ghq/github.com/tekihei2317/toydb-go/main.go:146 +0x190
```

```text
// command
select
.exit
```

`toydb-go`を直接実行してコンソールから入力すると動くけど、リダイレクトでコマンドを渡すとパニックするみたい。試してみた感じ。`.exit`では大丈夫ですが、`select`だとエラーになるみたいだった。

### 標準入力で無限ループする問題

小さい例で試してみたところ、次は`go run scanner.go < command`で動いた。

```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		input := scanner.Text()
		fmt.Println(input)
		if input == ".exit" {
			break
		}
	}
}
```

でも次は無限ループした。

```go
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		fmt.Println(input)
		if input == ".exit" {
			break
		}
	}
}
```

Scannerが複数あるとダメになるっぽいので、読み込みの関数にScannerを渡すようにしたら動くようになった。元の問題もこれで解決。

---

`[32]byte`を出力しているところで、nilがNULL文字として出力されていたので、nilを除外してから出力するように変更した。これでinsertしてselectするテストがパスした。

NULL文字とは？

あとは他のテストも実装して、ひとまず完成。テストコードのコマンドを実行したり標準入出力を操作する部分はChatGPTに書いてもらったのをそのまま使ったので、まだ理解できていない（deferとかioパッケージとか）。
