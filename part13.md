# memo

## Part 13

リーフノードを分割した後に、親ノードを更新する処理を実装する。次の2つのステップで行う。

- 親ノードのキーを、左側のノードの最大値で更新する
  - 分割したノードが一番右側だったらどうするんだろう
- 更新したキーの後ろに、右側のノードのポインタとキーを挿入する

まずは最初のステップを実装して、動作確認する。

```text
リーフノードの分割が起こったとき

            *, 8, *
[1,2,3,4,5,7,8] [9,10,11,12,13,14,15]

さらにインサートして、分割を起こす

            *, 7, *
[1,2,3,4,5,6,7] [8,8,8,8,8,8,8,8] [9,10,11,12,13,14,15]

内部ノードのキーが7に変わっていたらOK。真ん中のノードはまだポインタを設定していないので、表示されないはず。
```

```text
insert 1 user1 person1@example.com
insert 2 user2 person2@example.com
insert 3 user3 person3@example.com
insert 4 user4 person4@example.com
insert 5 user5 person5@example.com
insert 7 user7 person7@example.com
insert 8 user8 person8@example.com
insert 9 user9 person9@example.com
insert 10 user10 person10@example.com
insert 11 user11 person11@example.com
insert 12 user12 person12@example.com
insert 13 user13 person13@example.com
insert 14 user14 person14@example.com
insert 15 user14 person14@example.com
insert 6 user6 person6@example.com
insert 8 user8 person8@example.com
insert 8 user8 person8@example.com
insert 8 user8 person8@example.com
insert 8 user8 person8@example.com
insert 8 user8 person8@example.com
insert 8 user8 person8@example.com
.btree
.exit
```

できてるっぽい。次はキーとポインタを追加する処理を実装していく。

まずは、キーとポインタを挿入する位置を見つける。

```text
8を挿入する位置を見つける。インデックス1に挿入すればいいと分かる。

            *, 7, *
[1,2,3,4,5,6,7] [8,8,8,8,8,8,8,8] [9,10,11,12,13,14,15]

ページ番号と、8を挿入する。挿入する位置より後ろにキーがある場合は、ずらす必要がある（このケースはない）。

            *, 7, *, 8, *
[1,2,3,4,5,6,7] [8,8,8,8,8,8,8,8] [9,10,11,12,13,14,15]
```

ノードを一番右側に挿入するパターンもあるので、そのパターンを確認したい。

```text
分割した直後の状態

            *, 7, *
[1,2,3,4,5,6,7] [8,9,10,11,12,13,14]

ノードを追加して分割させる。キーの追加もその時にやってる気がする（違うかも）。
この時点では、2個目の*が[8, 9, ...]のノードを指してると思う。

            *, 7, *, 14
[1,2,3,4,5,6,7] [8,9,10,11,12,13,14] [15,15,15,15,15,15,15]

一番右のポインタが、追加したノードを指すようにする。ひとつ前のポインタが、ひとつ前のノードを指すようにする。

            *, 7, *, 14, *
[1,2,3,4,5,6,7] [8,9,10,11,12,13,14] [15,15,15,15,15,15,15]
```

```text
insert 1 user1 person1@example.com
insert 2 user2 person2@example.com
insert 3 user3 person3@example.com
insert 4 user4 person4@example.com
insert 5 user5 person5@example.com
insert 6 user6 person6@example.com
insert 7 user7 person7@example.com
insert 8 user8 person8@example.com
insert 9 user9 person9@example.com
insert 10 user10 person10@example.com
insert 11 user11 person11@example.com
insert 12 user12 person12@example.com
insert 13 user13 person13@example.com
insert 14 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
insert 15 user14 person14@example.com
.btree
.exit
```

なんかうまくいってそうな感じはします。

## Part 14

内部ノードの分割を実装する。

```text
最初の状態

    [*,5,*,12,*,20,*]
[1,5] [10,12] [14,20] [35]

11を挿入する

    [*,5,*,11,*,12,*,20,*]
[1,5] [10,11] [12] [14,20] [35]

中間ノードの容量が超えたので、分割をする。中間ノードのポインタはそのままで、キーが上に行く感じ。

            [*,12,*]
    [*,5,*,11,*] [*,20,*]
[1,5] [10,11] [12] [14,20] [35]
```

中間ノードがルートノードの場合は、新しいルートノードを作る必要がある。そうではない場合は、分割した後に2つのノードの間にあるキーを、親ノードに挿入する。

```text
最初の状態

            [*,12,*]
    [*,5,*,11,*] [*,20,*]
[1,5] [10,11] [12] [14,20] [35]

8を挿入する

            [*,12,*]
    [*,5,*,8,*,11,*] [*,20,*]
[1,5] [8] [10,11] [12] [14,20] [35]

9を挿入する

            [*,12,*]
    [*,5,*,9,*,11,*] [*,20,*]
[1,5] [8,9] [10,11] [12] [14,20] [35]

4を挿入する

            [*,9,*,12,*]
    [*,4,*,5,*] [*,11,*] [*,20,*]
[1,4] [5] [8,9] [10,11] [12] [14,20] [35]

中間ノードがいっぱいになったので、分割する。分割してできた右のノードと、分割した2つのノードの間のキーが浮いた状態になるので、それらを親ノードに挿入する。
```

これで

- 中間ノードがルートノードの場合
- 中間ノードがルートノードでない場合

の2つのパターンを確認できました。

### コードを見る

長いので一つずつ理解していきます。

最初に呼び出した段階では、childは新しく追加したリーフノードで、parentはその親ノード。parentを分割するので、oldという名前をつけていると思う。

```c
void internal_node_split_and_insert(Table* table, uint32_t parent_page_num,
                          uint32_t child_page_num) {
  uint32_t old_page_num = parent_page_num;
  void* old_node = get_page(table->pager,parent_page_num);
  uint32_t old_max = get_node_max_key(table->pager, old_node);

  void* child = get_page(table->pager, child_page_num); 
  uint32_t child_max = get_node_max_key(table->pager, child);

  uint32_t new_page_num = get_unused_page_num(table->pager);
```

分割するノードが、ルートノードかどうかで条件分岐をしています。ルートノードを分割する場合は、新しくルートノードを作成し、それをparentとしています。create_new_rootの中身はとりあえず置いておきます。

そうではない場合は、old_nodeの親ノードをparentとしています。

```c
  /*
  Declaring a flag before updating pointers which
  records whether this operation involves splitting the root -
  if it does, we will insert our newly created node during
  the step where the table's new root is created. If it does
  not, we have to insert the newly created node into its parent
  after the old node's keys have been transferred over. We are not
  able to do this if the newly created node's parent is not a newly
  initialized root node, because in that case its parent may have existing
  keys aside from our old node which we are splitting. If that is true, we
  need to find a place for our newly created node in its parent, and we
  cannot insert it at the correct index if it does not yet have any keys
  */
  uint32_t splitting_root = is_node_root(old_node);

  void* parent;
  void* new_node;
  if (splitting_root) {
    create_new_root(table, new_page_num);
    parent = get_page(table->pager,table->root_page_num);
    /*
    If we are splitting the root, we need to update old_node to point
    to the new root's left child, new_page_num will already point to
    the new root's right child
    */
    old_page_num = *internal_node_child(parent,0);
    old_node = get_page(table->pager, old_page_num);
  } else {
    parent = get_page(table->pager,*node_parent(old_node));
    new_node = get_page(table->pager, new_page_num);
    initialize_internal_node(new_node);
  }
```

次に、old_nodeを新しいノードに移動していく。（old_nodeは容量を超えているのだろうか。コードを見た感じでは、まだキーを挿入していなかった感じはする。）

curは、古いノードの一番右のページ。まずはこれを、新しいノード（分割したノードの移動先）の子ノードにしている。

それから、新しいノードへ順番に移動している。どこまで移動しているかは未理解（middle keyは残しているのかな）。

```c
  uint32_t* old_num_keys = internal_node_num_keys(old_node);

  uint32_t cur_page_num = *internal_node_right_child(old_node);
  void* cur = get_page(table->pager, cur_page_num);

  /*
  First put right child into new node and set right child of old node to invalid page number
  */
  internal_node_insert(table, new_page_num, cur_page_num);
  *node_parent(cur) = new_page_num;
  *internal_node_right_child(old_node) = INVALID_PAGE_NUM;
  /*
  For each key until you get to the middle key, move the key and the child to the new node
  */
  for (int i = INTERNAL_NODE_MAX_CELLS - 1; i > INTERNAL_NODE_MAX_CELLS / 2; i--) {
    cur_page_num = *internal_node_child(old_node, i);
    cur = get_page(table->pager, cur_page_num);

    internal_node_insert(table, new_page_num, cur_page_num);
    *node_parent(cur) = new_page_num;

    (*old_num_keys)--;
  }
```

分割した後にchildを挿入している雰囲気がある。ルートノードを分割しなかった場合は、internal_node_insertを再帰的に呼び出している。

```c
  /*
  Set child before middle key, which is now the highest key, to be node's right child,
  and decrement number of keys
  */
  *internal_node_right_child(old_node) = *internal_node_child(old_node,*old_num_keys - 1);
  (*old_num_keys)--;

  /*
  Determine which of the two nodes after the split should contain the child to be inserted,
  and insert the child
  */
  uint32_t max_after_split = get_node_max_key(table->pager, old_node);

  uint32_t destination_page_num = child_max < max_after_split ? old_page_num : new_page_num;

  internal_node_insert(table, destination_page_num, child_page_num);
  *node_parent(child) = destination_page_num;

  update_internal_node_key(parent, old_max, get_node_max_key(table->pager, old_node));

  if (!splitting_root) {
    internal_node_insert(table,*node_parent(old_node),new_page_num);
    *node_parent(new_node) = *node_parent(old_node);
  }
}
```

色々なところでメモリを読み書きしている（純粋関数がない）ので、どの関数もハイコンテキストで読み解くのが大変。とりあえず書きながら動かしてみます。

### 実際のパターンを用意しながら、やってみる

まずは、分割する内部ノードがルートノードだった場合を考える。

内部ノードのキーの数の上限は3個。リーフノードは14個目を挿入すると7+7個に分かれるので、25個挿入すると分割が見れる。


```text
分割直前の状態。12を空欄にしており、12を挿入すると分割が起きるようにしている。

                                    [*,7,*,21,*,28,*]
[1,2,3,4,5,6,7] [8,9,10,11,13,14,15,16,17,18,19,20,21] [22,23,24,25,26,27,28] [29,30,31,32,33,34,35]

12を挿入する。リーフノードの分割が起きる。

                                    [*,7,*,14,*,28,*]
[1,2,3,4,5,6,7] [8,9,10,11,12,13,14] [15,16,17,18,19,20,21] [22,23,24,25,26,27,28] [29,30,31,32,33,34,35]

```

```text
insert 1 user1 person1@example.com
insert 2 user2 person2@example.com
insert 3 user3 person3@example.com
insert 4 user4 person4@example.com
insert 5 user5 person5@example.com
insert 6 user6 person6@example.com
insert 7 user7 person7@example.com
insert 8 user8 person8@example.com
insert 9 user9 person9@example.com
insert 10 user10 person10@example.com
insert 11 user11 person11@example.com
insert 13 user13 person13@example.com
insert 14 user14 person14@example.com
insert 21 user14 person14@example.com
insert 22 user14 person14@example.com
insert 23 user14 person14@example.com
insert 24 user14 person14@example.com
insert 25 user14 person14@example.com
insert 26 user14 person14@example.com
insert 27 user14 person14@example.com
insert 28 user14 person14@example.com
insert 29 user14 person14@example.com
insert 30 user14 person14@example.com
insert 31 user14 person14@example.com
insert 32 user14 person14@example.com
insert 33 user14 person14@example.com
insert 34 user14 person14@example.com
insert 35 user14 person14@example.com
insert 15 user14 person14@example.com
insert 16 user14 person14@example.com
insert 17 user14 person14@example.com
insert 18 user14 person14@example.com
insert 19 user14 person14@example.com
insert 20 user14 person14@example.com
insert 12 user12 person12@example.com
.btree
.exit
```

えー、とりあえず写したらバグって困ってます。
