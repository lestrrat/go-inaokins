package webhook

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/google/go-github/github"
	"github.com/lestrrat/go-inaokins/webhook/internal/httputil"
	"github.com/lestrrat/go-slack"
	"github.com/lestrrat/go-slack/objects"
	"github.com/pkg/errors"
	"github.com/rjz/githubhook"
)

var secret []byte
var slackToken string
var messages []string
var msglen big.Int

func init() {
	secret = []byte(os.Getenv("GITHUB_SECRET"))
	slackToken = os.Getenv("SLACK_TOKEN")

	msgbuf, err := ioutil.ReadFile("messages.json")
	if err != nil {
		panic("failed to read messages.json")
	}
	if err := json.Unmarshal(msgbuf, &messages); err != nil {
		panic("failed to parse messages.json")
	}
	if len(messages) == 0 {
		panic("messages is empty")
	}
	msglen.SetInt64(int64(len(messages)))

	http.HandleFunc("/webhook/github", handleGithubWebhook)
	http.HandleFunc("/webhook/remind", handleRemindIssues)
}

func handleGithubWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := githubhook.Parse(secret, r)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	ctx := appengine.NewContext(r)

	switch hook.Event {
	case "issues":
		var payload github.IssuesEvent

		if err := json.Unmarshal(hook.Payload, &payload); err != nil {
			log.Debugf(ctx, "%s", hook.Payload)
			httputil.Error(w, http.StatusInternalServerError, errors.Wrap(err, "failed to unmarshal payload"))
			return
		}

		if err := handleIssue(ctx, &payload); err != nil {
			httputil.Error(w, http.StatusInternalServerError, errors.Wrap(err, "failed to handle issue"))
			return
		}
	default:
		// Silently ignore this payload
		w.WriteHeader(http.StatusOK)
	}
}

func handleIssue(ctx context.Context, payload *github.IssuesEvent) error {
	switch *payload.Action {
	case "assigned", "opened", "reopened":
		return startReminders(ctx, payload.Issue)
	case "unassigned", "closed":
		return stopReminders(ctx, payload.Issue)
	}
	return nil
}

// This is a wrapper so we can safely shove it in datastore
type Reminder struct {
	Issue []byte
}

func startReminders(ctx context.Context, issue *github.Issue) error {
	// key for this issue.
	k := datastore.NewKey(ctx, "issue", *issue.URL, 0, nil)

	// Store the issue
	buf, err := json.Marshal(issue)
	if err != nil {
		return errors.Wrap(err, `failed to serialize issue`)
	}

	if _, err := datastore.Put(ctx, k, &Reminder{Issue: buf}); err != nil {
		return errors.Wrap(err, `failed to store issue`)
	}

	return nil
}

func stopReminders(ctx context.Context, issue *github.Issue) error {
	// key for this issue.
	k := datastore.NewKey(ctx, "issue", *issue.URL, 0, nil)

	// Store the issue
	if err := datastore.Delete(ctx, k); err != nil {
		return errors.Wrap(err, `failed to delete issue`)
	}

	return nil
}

func chooseRemindMessage() string {
	idx, err := rand.Int(rand.Reader, &msglen)
	if err != nil {
		return messages[0]
	}

	return messages[int(idx.Int64())]
}

func handleRemindIssues(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	slackcl := slack.New(slackToken, slack.WithClient(urlfetch.Client(ctx)))

	q := datastore.NewQuery("issue")
	for t := q.Run(ctx); ; {
		var reminder Reminder
		if _, err := t.Next(&reminder); err != nil {
			if err == datastore.Done {
				break
			}
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		var issue github.Issue
		if err := json.Unmarshal(reminder.Issue, &issue); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}

		var attachment objects.Attachment
		attachment.Color = "danger"
		attachment.Text = chooseRemindMessage()
		attachment.Fields.Append(&objects.AttachmentField{
			Title: *issue.Title,
			Value: *issue.HTMLURL,
			Short: true,
		})
		if _, err := slackcl.Chat().PostMessage("#random").Attachment(&attachment).Do(ctx); err != nil {
			httputil.Error(w, http.StatusInternalServerError, err)
			return
		}
		log.Debugf(ctx, "%s", reminder.Issue)
	}

	w.WriteHeader(http.StatusOK)
}
