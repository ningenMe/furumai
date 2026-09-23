#!/usr/bin/env bash
# GitHub IssueをNotionの「タスク一覧」DBへ同期する(検索してあれば更新、なければ作成)。
# 依存はcurl/jqのみ(ubuntu-latestランナーに標準で入っている)。
#
# Issueのタイトル/本文はシェルコマンドとして組み立てない。GITHUB_EVENT_PATHの
# JSONをjqで読み、Notion APIへのリクエストボディもjq -n --argで構築する
# (シェル変数をコマンドラインに埋め込まないので、値に何が入っていても安全)。
set -euo pipefail

NOTION_API="https://api.notion.com/v1"
NOTION_VERSION="2025-09-03"

notion_api() {
  local method=$1 path=$2 body=${3:-}
  local args=(-sf -X "$method" "$NOTION_API$path"
    -H "Authorization: Bearer $NOTION_TOKEN"
    -H "Notion-Version: $NOTION_VERSION"
    -H "Content-Type: application/json")
  [ -n "$body" ] && args+=(-d "$body")
  curl "${args[@]}"
}

if [ "$GITHUB_EVENT_NAME" = "issues" ]; then
  issue=$(jq -c '.issue' "$GITHUB_EVENT_PATH")
  repo_full_name=$(jq -r '.repository.full_name' "$GITHUB_EVENT_PATH")
else
  # workflow_dispatch: inputs.issue_numberからissueをREST APIで取得する。
  issue_number=$(jq -r '.inputs.issue_number' "$GITHUB_EVENT_PATH")
  repo_full_name="$GITHUB_REPOSITORY"
  issue=$(curl -sf "https://api.github.com/repos/$repo_full_name/issues/$issue_number" \
    -H "Authorization: Bearer $GITHUB_TOKEN")
fi

title=$(jq -r '.title' <<<"$issue")
html_url=$(jq -r '.html_url' <<<"$issue")
number=$(jq -r '.number' <<<"$issue")
state=$(jq -r '.state' <<<"$issue")

if [ "$state" = "closed" ]; then status="完了"; else status="未着手"; fi

properties=$(jq -n \
  --arg title "$title" \
  --arg status "$status" \
  --arg repo "$repo_full_name" \
  --argjson number "$number" \
  --arg url "$html_url" \
  '{
    "タスク名": {title: [{text: {content: $title}}]},
    "状態": {status: {name: $status}},
    "リポジトリ": {rich_text: [{text: {content: $repo}}]},
    "Issue番号": {number: $number},
    "GitHub URL": {url: $url}
  }')

# データベースIDからdata source IDを解決する(Notion API 2025-09-03以降、
# クエリ/ページ作成はdata source単位で行う)。
data_source_id=$(notion_api GET "/databases/$NOTION_DATABASE_ID" | jq -r '.data_sources[0].id')

query_body=$(jq -n --arg url "$html_url" '{filter: {property: "GitHub URL", url: {equals: $url}}}')
existing_page_id=$(notion_api POST "/data_sources/$data_source_id/query" "$query_body" | jq -r '.results[0].id // empty')

if [ -n "$existing_page_id" ]; then
  notion_api PATCH "/pages/$existing_page_id" "$(jq -n --argjson properties "$properties" '{properties: $properties}')" >/dev/null
  echo "Updated Notion page for #$number"
else
  create_body=$(jq -n --arg dsid "$data_source_id" --argjson properties "$properties" \
    '{parent: {type: "data_source_id", data_source_id: $dsid}, properties: $properties}')
  notion_api POST "/pages" "$create_body" >/dev/null
  echo "Created Notion page for #$number"
fi
