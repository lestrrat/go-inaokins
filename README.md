# inaokins

This is a bot that helps you interface with technical journal extraordinaire, @inao.

# FEATURES

## Auto-remind assigned issues

When @inao assigns you a task through an issue, it is usually notified to you
via an email. This email is usually quickly lost, and then you run the risk of
angering the great @inao.

To counter this situation, the bot periodically polls issues in the specified
repo, and looks for open issues that are assigned to you.

While the the issue is open and assigned to you, it will keep notifying you
via Slack messages, every hour (except for 7pm - 7am).

You must explicitly close the issue or unassign yourself in order to stop it.

### Snooze

You may snooze a reminder by hand for up to 24 hours. Example: snooze issue
`#34` for 24 hours

```text
snooze #34 
```

# TODO

## Parse the content of @inao's messages for due dates

Find the due dates, and set appropriate milestones so that it's easy to track
which issues must be resolved by a particular date.

Once the deadline gets close, we should be sending more and more threatening
messages.

# CONFIGURATION

## inaokinsのURL

inaokinsはGoogle App Engine上で動作します。go1.8で動作しますので、2017年6月27日以降のGoogle App Engine SDKを利用する必要があります。

デフォルトの設定ではあなのGoogle Cloud Platform上でのプロジェクト名が`myproject-1234`だった場合、inaokinsは以下のURLにデプロイされます。

```
https://inaokins-dot-myproject-1234.appspot.com
```

## Webhookの設定

Githubで稲尾さんと共有しているレポジトリのWebhookを設定して、GAE上でデプロイされているinaokinsに通知が送られるようにします。

GAEのURLに`/webhook/github`を追加したものを該当レポジトリの設定に記入し、Content-Typeも`application/json`に変更してください。Secretも必須です。

[](./assets/webhook.png)

通知に必要なイベントは"Issue"だけです。

Secretに設定したパスワードは、app.yaml内の`GITHUB_SECRET`に設定してください。

## Slackの設定

SlackではAppを作成し、bot Tokenを入手します。そのうち「公式」Appをリリースするかもしれません。
bot Tokenはapp.yaml内の`SLACK_TOKEN`に設定してください

## Google App Engineへのデプロイ

```
gcloud app deploy
```
