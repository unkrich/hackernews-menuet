package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
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
	item.Children = func() []menuet.MenuItem {
		return p.ChildActions()
	}
	return item
}

func (p *Post) ChildActions() []menuet.MenuItem {
	// Title wrapped
	lines := wrap(p.Title, 42)
	items := make([]menuet.MenuItem, 0, len(lines)+1)
	for _, line := range lines {
		items = append(items, menuet.MenuItem{
			Text:       line,
			Clicked:    p.openLink,
			FontWeight:    menuet.WeightMedium,
		})
	}
	// Username
	items = append(items, menuet.MenuItem{
		Text:          fmt.Sprintf("By: %s", p.Username),
		Clicked:       p.openUserProfile,
		FontWeight:    menuet.WeightUltraLight,
	})
	// Timestamp posted
	items = append(items, menuet.MenuItem{
		Text: fmt.Sprintf("%s", p.TimestampString),
		FontWeight: menuet.WeightUltraLight,
	})

	items = append(items, menuet.MenuItem{
		Type: menuet.Separator,
	})

	//Comments
	items = append(items, menuet.MenuItem{
		Text:      fmt.Sprintf("Comments: %d", p.CommentCount),
		Clicked:   p.openComments,
	})
	// Upvote - disabled action because not signed in, no authHash
	if p.CanUpvote {
		items = append(items, menuet.MenuItem{
			Text:      fmt.Sprintf("Upvotes: %d (Click to Upvote)", p.Points),
			Clicked:   p.upvote,
		})
	} else {
		items = append(items, menuet.MenuItem{
			Text:      fmt.Sprintf("Upvotes: %d", p.Points),
		})
	}

	return items
}

// Post Actions
func (p *Post) upvote() {
	exec.Command("curl", p.voteHref()).Run()
	p.CanUpvote = false
	menuet.App().MenuChanged()
}

func (p *Post) openLink() {
	exec.Command("open", p.Link).Run()
}

func (p *Post) openUserProfile() {
	exec.Command("open", p.userHref()).Run()
}

func (p *Post) openComments() {
	exec.Command("open", p.commentsHref()).Run()
}

// Post links formatted
func (p *Post) commentsHref() string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", p.ID)
}

func (p *Post) voteHref() string {
	return fmt.Sprintf("https://news.ycombinator.com/vote?id=%d&auth=%s", p.ID, p.AuthHash)
}

func (p *Post) userHref() string {
	return fmt.Sprintf("https://news.ycombinator.com/user?id=%s", p.Username)
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

		err = fetchUsers()
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


// Helper
func wrap(text string, width int) []string {
	lines := make([]string, 0, len(text)/width)
	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}
	current := words[0]
	remaining := width - len(current)
	for _, word := range words[1:] {
		if len(word)+1 > remaining {
			lines = append(lines, current)
			current = word
			remaining = width - len(word)
		} else {
			current += " " + word
			remaining -= 1 + len(word)
		}
	}
	lines = append(lines, current)
	return lines
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