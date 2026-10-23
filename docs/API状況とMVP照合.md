# API状況とMVP照合

確認日：2026-10-24。根拠はsrc/handler/router.go、handler/service/repository/modelの実装、これまでのテスト結果。今回は資料更新とコード照合であり、稼働サーバーへの通し確認は行っていない。

## 結論

当初の「視聴作品・ジャンルをきっかけに相手を紹介し、共通点や相手経由の作品発見をJSONで提供する」というフェーズ1の主要機能は、コードとしてそろった。相手のプロフィールと2人の共通点・推薦作品は別APIで取得する。全実DBテストの最新コードでの一括実行、HTTPから実DBまでの一連の確認、性能・接続後処理等の仕上げは残るため、MVP検証完了とはまだ区別する。

## 現在のAPI一覧

すべてGET。IDは正のuint64。配列要素の作品情報はseries_id・title・genre_id。

|パス|正常応答・用途|エラー・注意点|
|---|---|---|
|/health|200、{"status":"ok"}|HTTPの生存確認。DB疎通は検査しない|
|/demo-users|200、users配列にuser_id。ランダム最大5人|取得失敗500。0件時はusers:nullのまま|
|/users/:userID/watch-series|200、series配列。作品ID順最大5作品|不正ID400、取得失敗500。不在と視聴0件はともに200・空配列|
|/users/:userID/watch-genre-counts|200、genre配列。genre_id・series_count|全視聴作品を集計、空ジャンル除外。件数降順・同数ならID順。不正400、失敗500。不在も空配列|
|/users/:userID/viewing-profile|200、user_id・series・genres|不正400、不在404、失敗500。seriesは最大5件、genresは全視聴作品の集計|
|/users/:userID/candidate-users|200、users配列にuser_id・reason|共通作品→共通ジャンル→その他で最大10人。不正400、不在404、失敗500。0件は空配列|
|/users/:userID/user-recommendation-detail/:otherUserID|200、user_id・other_user_id・common_series・common_genres・recommended_series|どちらかのID不正/同一IDは400、どちらか不存在404、取得失敗500。部分成功は返さない|

旧案の/candidates・/connectionsは登録されていない。現在の実装URLを使用する。

### 候補と詳細の意味

- reasonはshared_series・shared_genre・other。serviceが取得区分から付ける。相性スコアではない。
- 候補は視聴記録のあるユーザーから選び、自分・重複を除外する。下位区分は上位区分に該当する人を除外し、残り人数だけ取得する。全区分で不足すれば存在する分だけ返す。
- 同区分内はORDER BY RAND()。毎回必ず違う候補を保証しない。
- common_seriesは全視聴記録の共通作品を作品ID順、現実装は件数上限なし。
- common_genresは[{"genre_id":"movie"}]の形。空文字を除き重複なし・ID順、上限なし。共通作品に由来するジャンルも含む。
- recommended_seriesはotherUserID向けの既存推薦候補。ランダム最大5件、相手本人の視聴済みを除外する。閲覧者の視聴済みは除外せず、閲覧者への適合を保証しない。
- 相手の視聴作品・ジャンル集計は、otherUserIDでviewing-profileを呼ぶ。推薦詳細レスポンスにプロフィール全体は含まれない。
- 詳細APIは任意の異なる実在ユーザー2人を指定できる。直前の紹介候補に含まれていたかの検証・紹介履歴保存はしない。

## 当初要件との対応

|やりたかったこと|実装による対応|判定|
|デモで基準ユーザーを選ぶ|demo-users|実装済み|
|共通作品→共通ジャンル→その他の優先順|CandidateServiceと3つのrepositoryメソッド|実装済み|
|区分内ランダム、最大10人、不足補充、自分・重複除外|各SQLの条件・LIMIT、serviceで残り人数を計算|実装済み。性能と通し確認は残る|
|紹介理由を返す|候補のreason|実装済み|
|相手の視聴作品・ジャンルを返す|相手IDのviewing-profile|実装済み。詳細とは別API|
|2人の共通点を返す|詳細のcommon_series・common_genres|実装済み|
|相手への既存推薦から作品を発見する|詳細のrecommended_series|実装済み|
|Goのstruct/interface/DI・層分けを学ぶ|mainで注入、使用側interface、共通model、serviceの集約・分岐、スタブテスト|実装に反映済み|
|JSON APIのみ、ログイン・画面なし|バックエンドのみ|範囲どおり|

LIKE・相互LIKE・認証・画面は未実装だが、今回のフェーズ1の欠落ではない。性別による絞り込みや相性スコアも採用していない。

## 検証状況と限界

- 直近のgo test -short ./... -count=1とgo vet ./...は成功。handlerはルート経由、serviceはスタブで正常・異常・ID/context・取得順を検証。
- 同一IDは実serviceとhandlerをつなぎ、DB依存なしで400を返すテストも成功。
- 共通作品候補・その他候補・共通作品詳細・共通ジャンル詳細・推薦作品・存在確認は個別の実DBテスト成功を確認済み。
- 共通ジャンル候補（shared_genre_test.go）は作成済みだが、このチャットでは実DB実行結果未確認。
- 共通ジャンル詳細の空ジャンル除外ケースは、該当データがなくSKIP。SQL上の条件確認と、実データでの実証を区別する。
- 現データは視聴・推薦の重複が除去済み。推薦テストは返却作品に視聴済みがないことを確認するが、重複を意図的に投入した検証は行っていない。
- 全実DBテストの最新コードでの一括実行、実サーバー→DBの通し確認、RANDを含む性能評価は未完了。

## 次の確認と残作業

1. srcでgo test ./... -count=1を実行し、全実DBテストとSKIPの範囲を記録する。
2. go run .で起動し、demo-users→candidate-users→相手のviewing-profile→user-recommendation-detailを通して確認する。不正ID・不存在・同一IDも確認する。
3. ConnectDBのPing失敗時のdb.Close、SELECTの明示的な期限設定を本人が実装し、必要なテストを追加する。
4. src/.envのREAD_DB_*とCompose側のDB_READ_*等の二重化を整理するか、用途別に維持する方針を確定する。秘密値は出力しない。
5. 代表的なユーザーで候補SQLの時間・実行計画を確認する。既存の個別テストの短い実行時間だけで性能完了としない。
6. 共通点の件数上限なし、demo-usersのnullなど、現仕様をMVPとして採用するか最終確認し、検証結果を更新する。

アプリ本体は本人、レビュー・テスト・docsはアシスタントが担当。既存データ・ボリュームを維持し、無断でコミット・ステージしない。
