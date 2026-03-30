# portal-api

## はじめに

このプロジェクトはGo言語でWeb APIを開発するものです。

## 技術スタック

* 言語: Go言語
* フレームワーク: Gin
* PaaS: Render.com
* Database: Turso
* Security: bluemonday, API Key Auth, Rate Limiting
* Test: testify, httptest

## 存在するAPI

* 学習記録API

## コンセプト

### 学習記録API

自分のポートフォリオサイトに学習記録を載せるためです。
手作業でポートフォリオサイトに直接入力するのも手ですが、データ数が多いと大変なので、
CLIツールを窓口にして、このAPIを介してポートフォリオサイトに出力するというシステムを考えました。

## 概要

### 学習記録API

CLIツールから飛んできたデータをTursoに記録する。

## 技術選定理由

Render.comは15分間アクセスが無いとスリープするようなので、そのスリープから復帰する(コールドスタート)に時間がかかるようなのでそれを出来る限り短縮するためにGo言語を選択しました。

## 制約

使うことのできるユーザは開発者(私)のみです。

## ER図

### 学習記録API

<img src="./doc/studylog_er.svg" alt="ER図">

## シーケンス図

<img src="./doc/studylog_sequence.svg">
