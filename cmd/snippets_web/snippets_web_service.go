package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/ProdriveTechnologies/snippets/pkg/dates"
	"github.com/ProdriveTechnologies/snippets/pkg/schema"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type SnippetsWebService struct {
	database  *gorm.DB
	templates *template.Template
	selfUrl   string
}

func NewSnippetsWebService(database *gorm.DB, templates *template.Template, selfUrl string, router *mux.Router) *SnippetsWebService {
	sws := &SnippetsWebService{
		database:  database,
		templates: templates,
		selfUrl:   selfUrl,
	}
	router.HandleFunc("/", sws.handleLandingPage)
	router.HandleFunc("/others", sws.handleOthersList)
	router.HandleFunc("/{user_name:[a-z]+}/{year:[0-9]{4}}-W{week:[0-9]{2}}", sws.handleSnippetView)
	router.HandleFunc("/{user_name:[a-z]+}/subscribe", sws.handleSubscribe)
	router.HandleFunc("/{user_name:[a-z]+}/unsubscribe", sws.handleUnsubscribe)
	return sws
}

func getCurrentUser(req *http.Request) string {
	return req.Header.Get("X-Auth-Subject")
}

func (sws *SnippetsWebService) handleErrorPage(w http.ResponseWriter, req *http.Request, message string, code int) {
	log.Print(message)
	w.WriteHeader(code)
	if err := sws.templates.ExecuteTemplate(w, "error.html", struct {
		Message string
	}{
		Message: message,
	}); err != nil {
		log.Print(err)
	}
}

func (sws *SnippetsWebService) handleLandingPage(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, fmt.Sprintf("%s%s/%s", sws.selfUrl, getCurrentUser(req), dates.LastIsoWeek()), http.StatusSeeOther)
}

func (sws *SnippetsWebService) handleOthersList(w http.ResponseWriter, req *http.Request) {
	var users []schema.User
	if r := sws.database.Where("user_name != ?", getCurrentUser(req)).Find(&users); r.Error != nil {
		sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
		return
	}

	if err := sws.templates.ExecuteTemplate(w, "others.html", struct {
		Users    []schema.User
		LastWeek dates.IsoWeek
	}{
		Users:    users,
		LastWeek: *dates.LastIsoWeek().Seek(-1),
	}); err != nil {
		log.Print(err)
	}
}

func (sws *SnippetsWebService) createOrUpdateUser(req *http.Request) error {
	return sws.database.Assign(schema.User{
		RealName:     req.Header.Get("X-Auth-Name"),
		EmailAddress: req.Header.Get("X-Auth-Email"),
	}).FirstOrCreate(&schema.User{
		UserName: getCurrentUser(req),
	}).Error
}

// Converts HTML code submitted by the snippet edit form into a list of
// plain-text strings.
func extractListElementsFromHtml(code string) string {
	// Turn HTML tags into newlines.
	suppress := false
	out := ""
	for _, c := range code {
		switch c {
		case '<':
			suppress = true
			out += "\n"
		case '>':
			suppress = false
		default:
			if !suppress {
				out += string(c)
			}
		}
	}

	// Remove empty lines and unnecessary whitespace from the input.
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.Join(strings.Fields(html.UnescapeString(line)), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (sws *SnippetsWebService) handleSnippetEdit(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	week := dates.ParseIsoWeek(vars["year"], vars["week"])
	if week == nil {
		http.NotFound(w, req)
		return
	}

	userName := vars["user_name"]
	if userName != getCurrentUser(req) {
		sws.handleErrorPage(w, req, "Snippets from other users cannot be edited", http.StatusForbidden)
		return
	}

	req.ParseForm()
	bodyThisWeek := extractListElementsFromHtml(req.Form.Get("body_this_week"))
	bodyNextWeek := extractListElementsFromHtml(req.Form.Get("body_next_week"))
	if bodyThisWeek == "" && bodyNextWeek == "" {
		// No body provided. Delete the snippet if one exists.
		if r := sws.database.Where("user_name = ? AND year = ?  AND week = ?", userName, week.Year, week.Week).Delete(&schema.Post{}); r.Error != nil {
			sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Create or update the snippet.
		if err := sws.createOrUpdateUser(req); err != nil {
			sws.handleErrorPage(w, req, err.Error(), http.StatusInternalServerError)
			return
		}

		if r := sws.database.Assign(map[string]string{
			"body_this_week": bodyThisWeek,
			"body_next_week": bodyNextWeek,
		}).FirstOrCreate(&schema.Post{
			UserName: userName,
			Year:     week.Year,
			Week:     week.Week,
		}); r.Error != nil {
			sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
}

func splitLines(input string) []string {
	var lines []string
	for _, line := range strings.Split(input, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func (sws *SnippetsWebService) handleSnippetView(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		sws.handleSnippetEdit(w, req)
		return
	}

	vars := mux.Vars(req)
	week := dates.ParseIsoWeek(vars["year"], vars["week"])
	if week == nil {
		http.NotFound(w, req)
		return
	}

	userName := vars["user_name"]
	var post schema.Post
	if r := sws.database.Where("user_name = ? AND year = ? AND week = ?", userName, week.Year, week.Week).Take(&post); r.Error != nil && !gorm.IsRecordNotFoundError(r.Error) {
		sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
		return
	}

	template := ""
	realName := ""
	subscribed := false
	currentUser := getCurrentUser(req)
	if userName == currentUser {
		template = "snippet_edit.html"
	} else {
		template = "snippet_view.html"

		// Obtain real name.
		var user schema.User
		if r := sws.database.Where("user_name = ?", userName).Take(&user); r.Error != nil {
			if gorm.IsRecordNotFoundError(r.Error) {
				http.NotFound(w, req)
			} else {
				sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		realName = user.RealName

		// Obtain subscription.
		var subscription schema.Subscription
		if r := sws.database.Where("subscriber = ? AND subscribee = ?", currentUser, userName).Take(&subscription); r.Error == nil {
			subscribed = true
		} else if !gorm.IsRecordNotFoundError(r.Error) {
			sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
			return
		}
		realName = user.RealName
	}

	lastWeek := dates.LastIsoWeek()
	if err := sws.templates.ExecuteTemplate(w, template, struct {
		RealName            string
		PreviousWeek        *dates.IsoWeek
		CurrentWeek         dates.IsoWeek
		CurrentWeekFirstDay string
		CurrentWeekLastDay  string
		NextWeek            *dates.IsoWeek
		LastWeek            dates.IsoWeek
		BodyThisWeek        []string
		BodyNextWeek        []string
		Subscribed          bool
	}{
		RealName:            realName,
		PreviousWeek:        week.Seek(-1),
		CurrentWeek:         *week,
		CurrentWeekFirstDay: week.FirstDay(),
		CurrentWeekLastDay:  week.LastDay(),
		NextWeek:            week.Seek(1),
		LastWeek:            lastWeek,
		BodyThisWeek:        splitLines(post.BodyThisWeek),
		BodyNextWeek:        splitLines(post.BodyNextWeek),
		Subscribed:          subscribed,
	}); err != nil {
		log.Print(err)
	}
}

func (sws *SnippetsWebService) handleSubscribe(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		sws.handleErrorPage(w, req, "Expected POST request", http.StatusMethodNotAllowed)
		return
	}

	if err := sws.createOrUpdateUser(req); err != nil {
		sws.handleErrorPage(w, req, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(req)
	userName := vars["user_name"]
	if r := sws.database.FirstOrCreate(&schema.Subscription{
		Subscriber: getCurrentUser(req),
		Subscribee: userName,
	}); r.Error != nil {
		sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
}

func (sws *SnippetsWebService) handleUnsubscribe(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		sws.handleErrorPage(w, req, "Expected POST request", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(req)
	userName := vars["user_name"]
	if r := sws.database.Where("subscriber = ? AND subscribee = ?", getCurrentUser(req), userName).Delete(&schema.Subscription{}); r.Error != nil {
		sws.handleErrorPage(w, req, r.Error.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, req.Referer(), http.StatusSeeOther)
}
