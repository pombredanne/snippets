package main

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/smtp"
	"sort"
	"strings"

	"github.com/ProdriveTechnologies/snippets/pkg/dates"
	"github.com/ProdriveTechnologies/snippets/pkg/schema"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func splitLines(input string) []string {
	var lines []string
	for _, line := range strings.Split(input, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func main() {
	var (
		dbAddress     = flag.String("db.address", "", "Database server address.")
		smtpFrom      = flag.String("smtp.from", "", "Source email address.")
		smtpSmarthost = flag.String("smtp.smarthost", "", "SMTP server to use for sending emails.")
		snippetsUrl   = flag.String("snippets.url", "", "URL of the Snippets site.")
	)
	flag.Parse()

	// Week for which to generate snippets emails.
	week := *dates.LastIsoWeek().Seek(-1)

	db, err := gorm.Open("postgres", *dbAddress)
	if err != nil {
		panic(err)
	}

	// Query relevant data from the subscriptions table.
	var subscriptions []schema.Subscription
	if r := db.Find(&subscriptions); r.Error != nil {
		panic(r.Error)
	}
	users := map[string]bool{}
	usersWithSubscribers := map[string]bool{}
	usersWithSubscribees := map[string][]string{}
	for _, subscription := range subscriptions {
		users[subscription.Subscriber] = true
		users[subscription.Subscribee] = true
		usersWithSubscribers[subscription.Subscribee] = true
		usersWithSubscribees[subscription.Subscriber] = append(usersWithSubscribees[subscription.Subscriber], subscription.Subscribee)
	}

	// Query relevant data from the users table.
	var usersList []string
	for user, _ := range users {
		usersList = append(usersList, user)
	}
	var usersData []schema.User
	if r := db.Where("user_name IN (?)", usersList).Find(&usersData); r.Error != nil {
		panic(r.Error)
	}
	usersMap := map[string]schema.User{}
	for _, user := range usersData {
		usersMap[user.UserName] = user
	}

	// Query relevant data from the posts table.
	var usersWithSubscribersList []string
	for user, _ := range usersWithSubscribers {
		usersWithSubscribersList = append(usersWithSubscribersList, user)
	}
	var postsData []schema.Post
	if r := db.Where("user_name IN (?) AND year = ?  AND week = ?", usersWithSubscribersList, week.Year, week.Week).Find(&postsData); r.Error != nil {
		panic(r.Error)
	}
	postsMap := map[string]schema.Post{}
	for _, post := range postsData {
		postsMap[post.UserName] = post
	}

	for subscriber, subscribees := range usersWithSubscribees {
		user := usersMap[subscriber]

		// Override this line for testing.
		emailAddress := user.EmailAddress

		type Snippet struct {
			UserName     string
			RealName     string
			BodyThisWeek []string
			BodyNextWeek []string
		}

		// Fetch snippets.
		sort.Strings(subscribees)
		var snippets []Snippet
		var didNotWriteSnippets []schema.User
		for _, subscribee := range subscribees {
			subscribeeUser := usersMap[subscribee]
			if post, ok := postsMap[subscribee]; ok {
				snippets = append(snippets, Snippet{
					UserName:     subscribeeUser.UserName,
					RealName:     subscribeeUser.RealName,
					BodyThisWeek: splitLines(post.BodyThisWeek),
					BodyNextWeek: splitLines(post.BodyNextWeek),
				})
			} else {
				didNotWriteSnippets = append(didNotWriteSnippets, subscribeeUser)
			}
		}

		// Render email body.
		body := bytes.NewBuffer([]byte{})
		if err := snippetsEmailBody.Execute(body, struct {
			SnippetsUrl         string
			EmailAddress        string
			RealName            string
			Week                dates.IsoWeek
			Snippets            []Snippet
			DidNotWriteSnippets []schema.User
		}{
			SnippetsUrl:         *snippetsUrl,
			EmailAddress:        emailAddress,
			RealName:            user.RealName,
			Week:                week,
			Snippets:            snippets,
			DidNotWriteSnippets: didNotWriteSnippets,
		}); err != nil {
			panic(err)
		}

		// Send email.
		if err := smtp.SendMail(*smtpSmarthost, nil, *smtpFrom, []string{emailAddress}, body.Bytes()); err != nil {
			log.Print("Failed to send email to ", emailAddress, ": ", err)
		}
	}
}

var snippetsEmailBody = template.Must(template.New("email").Parse(
	`Subject: Snippets for {{.Week}}
To: {{.EmailAddress}}
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
	<head>
		<title>Snippets</title>
	</head>
	<body>
		<p>Hello {{.RealName}},</p>

		<p>You are receiving this email, because you are subscribed to
		one or more people on <a href="{{.SnippetsUrl}}">Snippets</a>.
		This email contains copies of snippets that people you are subscribed to have written last week.</p>

		{{$Week := .Week}}
		{{$SnippetsUrl := .SnippetsUrl}}
		{{range .Snippets}}
			<hr/>
			{{if .BodyThisWeek}}
				<h2>What has {{.RealName}} been up to last week?</h2>
				<ul>
					{{range .BodyThisWeek}}
						<li>{{.}}</li>
					{{end}}
				</ul>
			{{end}}

			{{if .BodyNextWeek}}
				<h2>What are {{.RealName}}'s plans for this week?</h2>
				<ul>
					{{range .BodyNextWeek}}
						<li>{{.}}</li>
					{{end}}
				</ul>
			{{end}}
			<p><a href="{{$SnippetsUrl}}{{.UserName}}/{{$Week}}">link</a></p>
		{{end}}

		{{if .DidNotWriteSnippets}}
			<hr/>
			<h2>People who did not write a snippet last week</h2>

			<ul>
				{{range .DidNotWriteSnippets}}
					<li><a href="{{$SnippetsUrl}}{{.UserName}}/{{$Week}}">{{.RealName}}</a></li>
				{{end}}
			</ul>
		{{end}}
	</body>
</html>`))
