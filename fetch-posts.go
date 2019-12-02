package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var fetched time.Time
var usersOnce sync.Once

func fetchAllPosts() (err error) {
	if fetched.After(time.Now().Add(-9 * time.Minute)) {
		return fmt.Errorf("Called too frequently (%v > %v)", fetched, time.Now().Add(-9*time.Minute))
	}

	newPosts, err := fetchPosts()
	posts = newPosts
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[j].Rank > posts[i].Rank
	})
	return err
}

func fetchPosts() ([]Post, error) {
	var err error
	fetched = time.Now()
	url := "https://news.ycombinator.com/"
	log.Printf("Fetching %s", url)
	resp, geterr := http.Get(url)
	if geterr != nil {
		return nil, geterr
	}
	if err != nil {
		return nil, err
	}
	newPosts, err := parsePosts(resp.Body)
	log.Printf("Got %d posts", len(newPosts))
	return newPosts, err
}

func parsePosts(r io.Reader) ([]Post, error) {
	posts := make([]Post, 0)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find(".athing").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			return
		}
		parsedId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Printf("Couldn't parse ID %s: %v", id, err)
			return
		}

		rank := s.Find(".rank").Text()
		parsedRank, err := strconv.ParseInt(rank[0:len(rank)-1], 10, 64)
		if err != nil {
			log.Printf("Couldn't parse rank %s: %v", rank, err)
			return
		}

		href, exists := s.Find("a.storylink").Attr("href")
		if !exists {
			return
		}

		authHash, exists := s.Find("#up_" + id).Attr("href")
		canUpvote := false
		parsedAuthHash := ""
		if !exists && strings.Contains(authHash, "auth") {
			re := regexp.MustCompile("auth=.*&")
			parsedAuthHash = re.FindString(authHash)[5:len(authHash)-1]
			canUpvote = true
		}



		// Data points found in the immediately following <tr>
		sibling := s.Next()
		points := sibling.Find(".score")
		if len(points.Text()) < 7 {
			return
		}
		parsedPoints, err := strconv.ParseInt(points.Text()[0:len(points.Text())-7], 10, 64) // 7 is length of " points"
		if err != nil {
			log.Printf("Couldn't parse points %s: %v", points, err)
			return
		}

		timestampWords := sibling.Find(".age").Text()

		commentCount := sibling.Find("td > a:last-child").Text()
		re := regexp.MustCompile("[0-9]+")
		parsedCommentsCount, err := strconv.ParseInt(re.FindString(commentCount), 10, 64)
		if err != nil {
			log.Printf("Couldn't parse comments %s : %v", commentCount, err)
			parsedCommentsCount = 0
		}

		post := Post {
			ID:                parsedId,
			Rank:              parsedRank,
			Link:              href,
			Title:             s.Find("a.storylink").Text(),
			Points:            parsedPoints,
			CommentCount:      parsedCommentsCount,
			Username:          sibling.Find(".hnuser").Text(),
			AuthHash:          parsedAuthHash,
			CanUpvote:         canUpvote,
			SiteStr:           s.Find(".sitestr").Text(),
			TimestampString:   timestampWords,
			Timestamp:         time.Now(),
		}
		posts = append(posts, post)
	})
	sort.Slice(posts, func(i, j int) bool {
		return posts[j].Rank > posts[i].Rank
	})
	if len(posts) > 10 {
		posts = posts[0:10]
	}
	return posts, nil
}

func fetchUsers() (err error) {
	usersOnce.Do(func() {
		users = make(map[string]User)
	})

	for _, post := range posts {
		userInfo, err := fetchUserInfo(post.Username)
		users[post.Username] = userInfo[0]
		if err != nil {
			log.Printf("Error fetching %s: %v", post.Username, err)
			continue
		}
	}
	return err
}

func fetchUserInfo(username string) ([]User, error) {
	var err error
	fetched = time.Now()
	url := "https://news.ycombinator.com/user?id=" + username
	log.Printf("Fetching %s", url)
	resp, geterr := http.Get(url)
	if geterr != nil {
		return nil, geterr
	}
	if err != nil {
		return nil, err
	}
	userInfo, err := parseUserInfo(username, resp.Body)
	log.Printf("Got userInfo for %s", username)
	return userInfo, err
}

func parseUserInfo(username string, r io.Reader) ([]User, error) {
	tempUsers := make([]User, 0)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find(".athing").Each(func(i int, s *goquery.Selection) {
		createdSibling := s.Next()
		createdString := createdSibling.Text()[8: len(createdSibling.Text())]

		karmaSibling := createdSibling.Next().Text()
		re := regexp.MustCompile("[0-9]+")
		karma, err := strconv.ParseInt(re.FindString(karmaSibling), 10, 64)
		if err != nil {
			log.Printf("Couldn't parse karm %s: %v", karmaSibling, err)
			return
		}

		user := User {
			Username:      username,
			Karma:         karma,
			CreatedString: createdString,
			Created:       time.Now(),
		}
		tempUsers = append(tempUsers, user)
	})
	sort.Slice(tempUsers, func(i, j int) bool {
		return tempUsers[j].Created.Before(tempUsers[i].Created)
	})
	return tempUsers, nil
}

