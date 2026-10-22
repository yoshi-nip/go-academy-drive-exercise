# Goのinterface：使う側が必要な契約を定義する

記録：2026-10-23。今回のUserHandlerとUserRepositoryを題材にした学習メモ。実装完了の報告ではなく、設計の意味と次の学習方針をまとめる。

## 今回の契約

handler/users.goに定義するinterface：

```go
type DemoUserReader interface {
    GetDemoUsers(ctx context.Context) ([]repository.DemoUserID, error)
}

type UserHandler struct {
    reader DemoUserReader
}
```

これは処理の実装ではない。「この名前・引数・戻り値を持つメソッドを呼べる」という条件を表す。単に同名の関数がパッケージにあるだけでは満たさない。

repository/users.goの具体的な実装：

```go
func (r *UserRepository) GetDemoUsers(
    ctx context.Context,
) ([]DemoUserID, error) {
    // 保持しているreadDBを使って取得する
}
```

*UserRepositoryのメソッドが契約に一致するため、次の引数として渡せる。

```go
repo := repository.NewUserRepository(db)
h := handler.NewUserHandler(repo)
```

この呼び出しではコンパイラが型の適合を確認する。実行時にメソッドを探して適合を判定する仕組みではない。repository側にimplements宣言は不要で、handlerのinterfaceをimportする必要もない。

## struct・interface・ポインタの違い

- UserRepository struct：DBを保持し、取得処理をメソッドとして持つ具体的な型。
- DemoUserReader interface：取得担当に必要なメソッドの契約。
- UserHandler.reader：その契約を満たす具体的な値を保持するフィールド。

```go
func NewUserHandler(source DemoUserReader) *UserHandler {
    return &UserHandler{reader: source}
}
```

readerは保存先のフィールド名、sourceは引数。今回sourceとして渡される具体的な値は*UserRepository。interfaceの値は、その具体的な型と値を保持する。構造体へのポインタを渡した場合、構造体そのものをコピーして保存するわけではない。

h.reader.GetDemoUsers(ctx)は、保存されている具体的な値のメソッドを呼ぶ。本物のrepositoryならSQLを実行し、テスト用の実装なら固定データや指定したエラーを返せる。

引数を*UserRepositoryにするとその具体的な型を要求する。DemoUserReaderにすると、その契約を満たす別の型も受け取れる。通常は*DemoUserReaderにしない。それは具体的な構造体へのポインタではなく、interface変数自体へのポインタである。

## なぜ使う側に定義するのか

handlerはrepo全体の実装を固定するのではなく、自分が使うために必要なメソッドだけを要求する。

repoにGetViewingProfileやFindUserByIDが増えても、デモユーザー取得だけに依存する側はGetDemoUsersだけを契約にすればよい。実装側が用意した大きなinterfaceを受け取ると、使わないメソッドまでテスト用実装に要求することがある。

利点は、利用側の必要性に合わせて依存を小さく保てること。repoは余分なメソッドを持っていても構わない。また、実装側がたまたま同じ小さいinterfaceを提供しているなら、そこに定義したことだけで直ちに悪い設計になるわけではない。配置そのものより契約の大きさと依存の向きが重要。

## 固定される部分と交換できる部分

メソッド名・引数・戻り値は固定する。交換できるのは取得の実装。

|具体的な実装|取得方法|
|---|---|
|UserRepository|MySQLから取得|
|テスト用struct|固定のID一覧、または指定したエラーを返す|
|将来のキャッシュ実装|メモリ等から取得|

interfaceはどんな変更も吸収する仕組みではない。戻り値などの契約を変えたら、接続する側も修正する必要がある。同じメソッド形でも意味が異なれば安全な差し替えとはいえないため、件数・エラー等の振る舞いも守る。

**呼び出し方と結果の型を固定することで、中の実装を交換できる。使う側は、自分に必要な契約を定義する。** これが今回学ぶinterfaceの中心。

## 現在残っているパッケージ依存

戻り値が[]repository.DemoUserIDなので、handlerは今もrepositoryパッケージをimportしている。SQLの実装は交換できるが、パッケージの依存までなくなったわけではない。この2つは区別する。

今回決めた進め方：

1. まず現在の戻り値の型のまま、struct・interfaceによるDIを完成させる。
2. 本物のrepositoryとテスト用structを差し替え、APIの挙動を保ってDBなしのテストを通す。
3. その後、別のリファクタリングで戻り値の型の所有場所と依存関係を整理する。

後続では、例えばDBやHTTPに依存しない型をmodel等のパッケージへ置き、handlerとrepositoryの両方がその型を使う案を検討する。パッケージ名や配置はまだ確定していない。interfaceだけを別ファイルへ移しても、戻り値にrepositoryの型が残れば依存は解消しない。

APIのJSON形式や取得条件は構造変更と混ぜて変えない。serviceは業務ルールが必要になった段階で導入する。
