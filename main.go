package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/caseymrm/menuet"
)

// Post represents one HackerNews post/submission
type Post struct {
	ID               int64
	Rank			 int64
	Link             string
	Title            string
	Points           int64
	CommentCount     int64
	Username         string
	AuthHash         string
	CanUpvote        bool
	TimestampString  string
	SiteStr          string
	Timestamp        time.Time
}

// Item returns a short menu item for the Post, trucated as requested
func (p *Post) Item(truncate int) menuet.MenuItem {
	text := p.Title
	if len(text) > truncate-2 {
		text = fmt.Sprintf("%s...", p.Title[0:truncate-3])
	}
	item := menuet.MenuItem{
		Text:       text,
		FontWeight: menuet.WeightRegular,
		Clicked:    p.openLink,
	}
	return item
}

func (p *Post) openLink() {
	exec.Command("open", p.Link).Run()
}

// User represents one HackerNews user
type User struct {
	Username         string
	Karma            int64
	CreatedString	 string
	Created          time.Time
}

var posts []Post
var users map[string]User

func checkHackerNews() {
	menuet.App().SetMenuState(&menuet.MenuState{
		Title: "|Y| HN",
	})	
	ticker := time.NewTicker(10 * time.Minute)
	for ; true; <-ticker.C {
		err := fetchAllPosts()
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		
		menuet.App().MenuChanged()
	}
}

func menuItems() []menuet.MenuItem {
	items := make([]menuet.MenuItem, 0, 2*len(posts)+3)
	items = append(items, menuet.MenuItem{
		Text:     "Recent posts",
		FontSize: 12,
	})
	for _, post := range posts {
		post := post

		items = append(items, post.Item(50))
	}
	return items
}


func main() {
	go checkHackerNews()
	app := menuet.App()
	app.Name = "HackerNews Menuet"
	app.Label = "com.github.unkrich.hackernews-menuet"
	app.Children = menuItems
	app.AutoUpdate.Version = "v0.1"
	app.AutoUpdate.Repo = "unkrich/hackernews-menuet"
	app.RunApplication()
}