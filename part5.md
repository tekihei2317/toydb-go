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

コンポーネントの依存関係がわかりにくくなってきたので、モジュールに分割してみる。main以外のパッケージを作ったことがないので、作り方を調べてみる。
