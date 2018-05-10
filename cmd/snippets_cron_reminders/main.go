package main

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/smtp"
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

	backlogWeeks := 6

	// Week for which to generate snippets emails.
	week := *dates.LastIsoWeek().Seek(-backlogWeeks)
	thisWeek := dates.LastIsoWeek()

	db, err := gorm.Open("postgres", *dbAddress)
	if err != nil {
		panic(err)
	}

	// Query relevant data from the users table.
	var users []schema.Post
	if r := db.Select("user_name").Where("(year = ? and week >= ?) or year > ?", week.Year, week.Week, week.Year).Group("user_name").Find(&users); r.Error != nil {
		panic(r.Error)
	}

	var usersList []string
	for _, user := range users {
		usersList = append(usersList, user.UserName)
	}

	var lastPosts []schema.Post
	if r := db.Where("user_name in (?) and week = ?", usersList, thisWeek.Week).Find(&lastPosts); r.Error != nil {
		panic(r.Error)
	}

	type Snippet struct {
		UserName     string
		BodyThisWeek []string
		BodyNextWeek []string
	}

	currentSnippetsMap := map[string]Snippet{}
	for _, post := range lastPosts {
		currentSnippetsMap[post.UserName] = Snippet{
			UserName:     post.UserName,
			BodyThisWeek: splitLines(post.BodyThisWeek),
			BodyNextWeek: splitLines(post.BodyNextWeek),
		}
	}

	var usersInfo []schema.User
	if r := db.Where("user_name in (?)", usersList).Find(&usersInfo); r.Error != nil {
		panic(r.Error)
	}

	for _, user := range usersInfo {
		// Override this line for testing.
		emailAddress := user.EmailAddress

		// Render email body.
		body := bytes.NewBuffer([]byte{})
		if err := snippetsEmailBody.Execute(body, struct {
			SnippetsUrl    string
			EmailAddress   string
			UserName       string
			RealName       string
			BacklogWeeks   int
			CurrentSnippet Snippet
		}{
			SnippetsUrl:    *snippetsUrl,
			EmailAddress:   emailAddress,
			UserName:       user.UserName,
			RealName:       user.RealName,
			BacklogWeeks:   backlogWeeks,
			CurrentSnippet: currentSnippetsMap[user.UserName],
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
	`Subject: Snippets reminder
To: {{.EmailAddress}}
Content-Type: text/html; charset=utf-8

<!DOCTYPE html>
<html>
	<head>
		<title>Snippets</title>
	</head>
	<body>
		<p>Hello {{.RealName}},</p>

		<p>You are receiving this email, because you wrote on
		<a href="{{.SnippetsUrl}}">Snippets</a> during any of the past
		{{.BacklogWeeks}} weeks.</p>

		{{if .CurrentSnippet.BodyThisWeek}}
			<p>You currently wrote the following:</p>
			<ul>
				<li>What have you been up to this week?</li>
				<ul>
					{{range .CurrentSnippet.BodyThisWeek}}
						<li>{{.}}</li>
					{{end}}
				</ul>
				{{if .CurrentSnippet.BodyNextWeek}}
					<li>What are your plans for next week?</li>
					<ul>
						{{range .CurrentSnippet.BodyNextWeek}}
							<li>{{.}}</li>
						{{end}}
					</ul>
				{{end}}
			</ul>
		{{else}}
			<p>You currently didn't write any snippets this week.</p>
		{{end}}

		<p>Your snippet will be sent on Monday to your subscribers.
		Please make sure they are completed by then.</p>
	</body>
</html>`))
