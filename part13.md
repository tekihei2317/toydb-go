# memo

## Part 13

リーフノードを分割した後に、親ノードを更新する処理を実装する。次の2つのステップで行う。

- 親ノードのキーを、左側のノードの最大値で更新する
  - 分割したノードが一番右側だったらどうするんだろう
- 更新したキーの後ろに、右側のノードのポインタとキーを挿入する

まずは最初のステップを実装して、動作確認する。

13個なので、14個で分割する

```text
リーフノードの分割が起こった

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
