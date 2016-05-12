package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"net/http"
	"html/template"
	"log"
	"strconv"
	"fmt"
)

const (
	GIT_REPO = "GIT_REPO_NAME"
	GIT_USER = "GIT_USER"
	AUTH_TOKEN = "GITHUB_AUTH_TOKEN"
)

var (
	client *github.Client
)

// An entry in the dev-log table
type DevLogEntry struct {
	AuthorName      string
	AuthorURL       string
	AuthorAvatarURL string
	Rev             string
	RevLink         string
	Message         string
}

// Information about the dev log
type DevLogInfo struct {
	Entries     []DevLogEntry
	NextPageURL string
	PrevPageURL string
}

func init() {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: AUTH_TOKEN})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client = github.NewClient(tc)
}

func main() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("templates/index.html")

		p := req.URL.Query().Get("p")
		page := 1
		if p != "" {
			page, err = strconv.Atoi(p)
			if err != nil {
				// silently ignore the error
				page = 1
			}
		}

		if err != nil {
			log.Fatal(err)
		}
		info, err := getLog(page, req.Host)

		if err != nil {
			log.Fatal(err)
		}

		t.Execute(res, info)
	})

	http.ListenAndServe(":8080", nil)
}

func getLog(page int, host string) (DevLogInfo, error) {

	rcs, res, err := client.Repositories.ListCommits(GIT_USER, GIT_REPO, &github.CommitsListOptions{
		ListOptions: github.ListOptions{Page: page, PerPage: 20}})
	if err != nil {
		return DevLogInfo{}, err
	} else {
		entries := make([]DevLogEntry, len(rcs))
		for i, v := range rcs {
			//fmt.Println(v)
			an := *v.Commit.Author.Name
			if an == "" {
				an = "Unknown"
			}

			au := "#"
			aau := "https://i2.wp.com/assets-cdn.github.com/images/gravatars/gravatar-user-420.png?ssl=1"

			// Some what of a hack to deal with inconsistencies caused by Git/GitHub
			if v.Author != nil {
				if v.Author.HTMLURL != nil {
					au = *v.Author.HTMLURL
				}

				if v.Author.AvatarURL != nil {
					aau = *v.Author.AvatarURL
				}
			}

			rev := *v.SHA
			revl := *v.HTMLURL
			m := *v.Commit.Message

			// Trim it down to 64
			if len(m) > 64 {
				m = m[:64] + "..."
			}

			entries[i] = DevLogEntry{
				AuthorName: an,
				AuthorURL: au,
				AuthorAvatarURL: aau,
				Rev: rev[:8],
				RevLink: revl,
				Message: m,
			}

		}

		info := DevLogInfo{Entries: entries}

		// Probably a nicer, cleaner way of doing this, but it works :)
		if res.NextPage != 0 {
			info.NextPageURL = fmt.Sprint("http://", host, "/?p=", res.NextPage)
		}

		if res.PrevPage != 0 {
			info.PrevPageURL = fmt.Sprint("http://", host, "/?p=", res.PrevPage)
		}

		return info, nil
	}
}

