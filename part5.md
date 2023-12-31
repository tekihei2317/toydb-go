# memo

## Part 5

現状はインメモリのため、ファイルに保存してプログラムの終了後にもデータが残っているようにする。実装としては、ページの内容をそのままディスクに書き込めばOK。このためにページャーを作る。ページャーは位置を渡すとそのページをキャッシュまたはディスクから取得し、取得したデータをキャッシュしてくれる。

ファイルの読み書きが必要なので、Goでどのようにするのかを調べる。元のコードでは、lseek・read・open・writeを使っている。

- 指定したバイト数だけ進める（そこから読み始める）
- 指定したバイト数だけ読み取る
- 指定したバイト数だけ書き込む

の3つができればOK。

```c
// get_page: ファイルからページの内容を読み取り、 ページャーにキャッシュする

// ファイルディスクリプタを、指定したバイトの数だけ進める
lseek(pager->file_descriptor, page_num * PAGE_SIZE, SEEK_SET);

// ファイルから指定したサイズのデータを読みとり、バッファに格納する
ssize_t bytes_read = read(pager->file_descriptor, page, PAGE_SIZE);

// pager_open: ページャーを初期化する

// ファイルを開いて、ファイルディスクリプタを取得する
int fd = open(filename,
  O_RDWR | 	// Read/Write mode
    O_CREAT,	// Create file if it does not exist
  S_IWUSR |	// User write permission
    S_IRUSR	// User read permission
);

// ファイルのバイト数を取得する
off_t file_length = lseek(fd, 0, SEEK_END);

// pager_flush: ページの内容をディスクに書き込む

// 省略
off_t offset = lseek(pager->file_descriptor, page_num * PAGE_SIZE, SEEK_SET);

// バッファの内容を、ファイルに書き込む
ssize_t bytes_written = write(
  pager->file_descriptor, pager->pages[page_num], size
);
```

まず、ファイル全体の読み取りは`os.ReadFile`が使える。

ファイルの途中から読み取りたい場合は、`os.Open`で`os.File`のインスタンスを作ってから、`File.Seek`でオフセットを進めて、`File.Read`で読み取る。`File.Read`はバイトのスライスを渡し、最大で`len(b)`だけ読み取る。

ファイルがなければ作成する場合は、`os.OpenFile`を使うのが一番楽そう。`O_RDWR`と`O_CREAT`のフラグをつけてファイルを開くと、なかった場合に作成してくれる。`os.Open`は`O_RDONLY`なことに注意する。途中から書き込みたい場合は、読み取りの場合と同様に`File.Seek`を使う。

// ファイルディスクリプタとは？

- [x] dbOpenを作る
- [x] pagerに書き込むように変更する（table.insertRow）
- [x] dbCloseを実装する
- [x] pager.flushを実装する（ファイルに書き込む処理）
- [x] 読み取る時に、ディスクから読み取るようにする（ページャにキャッシュする）

読み取る時にディスクから読み取らなかったらどうなるんだろう。

ページ、ディスクから読み込み済みかどうかのデータが必要かな。ポインタにしてnilにするのは微妙な気がする。nilが、insertで書き込んだという情報と、selectで読み込んだという2つの情報を持っているため。でもこれらは区別しなくてもいいのかもしれない。

書き込む時に全てのデータを書き込む？訳ではないか。nilじゃないものを書き込んでいる。これは、selectで取得したディスクのデータと、insertで書き込んだデータの2つがある。

## Part 6

B-Treeの実装をするための準備パート。残りの章はB-Treeの実装っぽい。

- テーブルの開始、終了を表すカーソルを作成する
- カーソルが指す行にアクセスする
- カーソルを次の行に進める

これらを実装する。

行が保存されている位置のポインタを返す`rowSlot`関数が、`cursorValue`関数に書き換えられる。

### リファクタリング

コンポーネントの依存関係がわかりにくくなってきたので、モジュールに分割してみた。main以外のパッケージを作るには、ディレクトリを作る必要がある。例えば、`table`パッケージを作るには`table/`ディレクトリを作る。

ひとまず次のように分割した。

- `table`パッケージ（名前微妙かも）
  - テーブルに関する操作を含む
  - 含まれるコンポーネントはテーブル、カーソル
- `persistence`パッケージ
  - ファイルやメモリへの書き込み・読み込みを担当する
  - 含まれるコンポーネントはページャ

`main`パッケージのコードはそのままにしている。クエリのパース処理、クエリを実行する処理などに分割できると思う。そういえばバイトコードを作る処理は実装しないような気がする（複雑なSQLに対応しないため）。

ちょっと微妙なところは、ページャに行のサイズ（`ROW_SIZE`）をハードコーディングしているところ。実際はテーブルから計算した数値を渡す必要があると思う。テーブルのスキーマが変わった場合とかも考えると、ページャの実装が難しそうだなと思った。

## Part 7

SQLiteは、インデックスの実装にB-treeを、テーブルの実装にB+ treeを使っている。これからはテーブルの実装をするので、B+ treeの実装をする。

B-treeとB+ treeの違いは、B+ treeが内部ノードに値を持っていないこと。B+ treeは、リーフノードのみが値を持っており、内部ポードはキーとポインタしか持っていない。

B+ treeに挿入するアルゴリズムは次のような感じ。

- リーフノードの挿入位置を見つけて挿入する
- リーフノードの容量（X）を超えたら、リーフノードを分割する
- ノードを分割すると、その境界を表すキーを親ノードに挿入する
- キーの挿入によって、親ノードの容量（Y）を超えたら分割する。これを親ノードをたどりながら、容量を超えなくなるまで再帰的に繰り返す。
- 最終的にルートノードにたどり着いて分割した場合は、新たにルートノードを作って深さが1増える

もてる子ノードの数の上限をorderという。例えば子ノードの上限が3の場合は、order 3のB-treeという。

内部ノードの容量Yは、子ノードを上限まで持った時に最大になるので、`order - 1`になる。リーフノードの容量Xは、`order`と同じになる？（分かっていない）

次のサイトで動作を見てみるとなんとなく理解できた。 [B+ Tree Visualization](https://www.cs.usfca.edu/~galles/visualization/BPlusTree.html)

分からなかったのは、B-treeとページの対応。ノードがページに対応するとあったけど、どういう意味なのかが分からなかった（ノードに直接データを入れるのか？リーフノードにページ番号が入っているのか？ページを分割するのか？、こんがらがってる）

## Part 8

ページは、これまでと同じようにページ番号順にファイルに保存されるっぽい。

B-treeの役割は、検索するときはキーから行が保存されている位置を見つけること。位置が分かれば、ページャからデータを取得できる。挿入するときは、挿入する位置を見つけたり、再帰的にノードを分割しながらページャのデータを変更しそう。

ノードがどのようなデータを持っているのかが気になる。リーフノードは行を持っているらしい。つまり、リーフノードがページャの役割をしているっぽい。

部分的なページの読み書きをサポートしなくて良くなったとあるけど、どういうことだろう。←ページごとに行数を持っているため、全部書き込んだとしてもどこまでが行なのか分かるから。

- [x] ページャにページの数を持たせるようにする（getPageでnumPagesを変更してるのが気になる）、ファイルから読み込むときにnumPagesを設定する。
- [x] カーソルを(ページ番号,セル番号)のペアに変更する。合わせて、tableStartとtableEndも変更する。tableEndでは、ページからセル数を取得する。
  - ↑tableEndの実装はノードが増えてくるとかなり変わりそう
- [x] cursorValue、cursorAdvanceを実装する。
- [ ] ページの内容をそのままファイルに書き込むように変更する

cursorValueはどうやって実装すればいいだろう。使っているのは行を挿入するところだけ。ページの開始位置と終了位置が分かればいいので、ヘッダーとキーのサイズを考慮するようにすれば良さそう。
