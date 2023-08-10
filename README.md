# toydb-go

https://cstack.github.io/db_tutorial/

## 学びたいこと

- データがどのように保存されているか。B-treeを実際に実装してみて、インデックスがどのように使われるのかをイメージできるようになりたい。
- SQLの構文解析の実装を知りたい。ORMを作るときに、クエリとスキーマから入出力の型を生成する処理を書きたいため（sqlcみたいなイメージ）。構文解析はほとんど実装したことがない。

## データベースについて知っていること

- SQLはデータをどのように取得するかではなく、何を取得したいかを記述する。データの具体的な取得方法はDBMSのオプティマイザが決める。
- その性質上、SQLを高速する方法で最も一般的なのはインデックスを付与すること。
- インデックスはソート済みの列（インデックスを付与した列）を別で持っているイメージをしている。そのため、検索（WHERE）は二分探索によって、ソート（ORDER BY）は順番に取り出すことによって高速化できる。
- SQLは集合論や述語論理などの数学の理論を基礎にしている。
- リレーショナルモデルのリレーションとは、タプルの集合のこと。リレーションとは、実装の言葉で言うとテーブルのこと。
- データの更新を正しく行うために、ロックやトランザクションなどの機構がある。アルゴリズムや実装は調べたことがない。
- 従来のRDBMSは、書き込みを直列に行うようになっている（気がする）。読み込みの負荷はデータをコピーしたレプリカを作ることで分散することができるが、書き込みはできない。そのため、書き込みが多いサービスでは（ちょっと前は）NoSQLを使っていた。

## Chapter 1

SQLが実行されるまで、

トークナイザー→パーサー→コードジェネレーター→

バーチャルマシン→B-tree→pager→OS interface

のようにたどっていく。SQLはSQLite用のバイトコードに変換され、バーチャルマシンで実行される。前半のコンポーネントをフロントエンド、後半をバックエンドという。

とりあえず入力を受け取れるようにする。fmt.Scanlnに文字列のポインタではなく文字列を渡すと、入力を読み取る前に次に進んでしまった。

構造体のポインタを持っている時、ドットアクセスする時は*を省略できるんでしたっけ。

## Chapter 2

fmt.Scanlnだと'a b'と入力した時にaしか取得できなかった。空白区切りで変数に割り当てられるようだった。なので、bufio.Scanner.scan()で一行読み取るように変更した。

文字列が特定の文字列で始まるかどうかは、strings.HasPrefixで判定できる。string.HasSuffixもある。

## Chapter 3

まずは、ハードコードした1つのテーブルだけで、append onlyのインメモリデータベースを作成する。

fmt.Sscanfで文字列をパースしているけれど、適当に入力してもパースは失敗しなかった。その場合は、構造体にはゼロ値が入ったままだった。

データは配列に保存している。データのサイズを考慮しながら、memcopyでID、ユーザー名、メールアドレス、ID、ユーザー名、メールアドレスのように保存している。

---

Goで書き換える場合は、unsafe.Pointerというのを使えばいいらしい。

[Go言語: ポインターとそれに関する型(uintptr, unsafe.Pointer) - Qiita](https://qiita.com/nozmiz/items/291f16f619a939bd7b87)

文字列って何バイト？1バイト = 16進数2桁で、GoはUTF-8なので1~4バイト。UTF-8ではアルファベットは1バイトだけれど、ひらがなの「あ」は3バイトで、絵文字の「😨」は4バイト。

[Goのruneを理解するためのUnicode知識 - Qiita](https://qiita.com/seihmd/items/4a878e7fa340d7963fee)

文字列にfor文でインデックスアクセスした場合は、1バイトごとに表示される。つまり「あ」は"e3 81 82"のように3つ表示される。

バイトごとではなく文字ごとにアクセスしたい場合は、runeを使う。

---

文字列をバイトの配列にキャストしようとすると、次のエラーになった。

```go
// cannot convert "tekihei2317" (untyped string constant) to type byte
[32]byte("tekihei2317")
```

バイトの配列にはキャストできないけど、バイトのスライスにはキャストできた。そもそも、固定長の文字列をどうやって持つべきかがよくわからない。とりあえず[32]byteみたいな感じにしてみた。

これだと、1バイトを超える大きさの文字を格納することができない。文字コードが難しい。

```go
// OK
{'a', 'b'}

// NG
{'あ', 'い'}
```

reflect.DeepEqualは、配列とスライスを比較するとfalseになるっぽい。その場合は、スライスに[:]をつけて配列に変換する？

Printfの書式の%vと%+vの違いは？

[2つのsliceがすべて同じ要素をもっているかどうかを比較するショートカット - Qiita](https://qiita.com/taksatou@github/items/48b22d6d37e99a6c21cc)

---

テーブルは、ページへの参照を配列で持ちます。ページには、複数の行が含まれています。

ポインタとintって足し算できないんだったっけ。

```go
func rowSlot(table *Table, rowNum uint32) unsafe.Pointer {
	pageNum := rowNum / ROWS_PER_PAGE
	page := table.pages[pageNum]

	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * uint32(ROW_SIZE)

  // invalid operation: page + int(byteOffset) (mismatched types unsafe.Pointer and int)
	return page + int(byteOffset)
}
```

ポインタを進める時はこうすればいいらしい。

```go
return unsafe.Pointer(uintptr(page) + uintptr(byteOffset))
```
