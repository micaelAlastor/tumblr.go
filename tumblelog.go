package tumblr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Reference to a blog which can be used to perform further blog actions
type BlogRef struct {
	client ClientInterface
	Name   string `json:"name"`
}

// Tumblelog struct
type Blog struct {
	BlogRef
	Url                  string `json:"url"`
	Title                string `json:"title"`
	Posts                int64  `json:"posts"`
	Ask                  bool   `json:"ask"`
	AskAnon              bool   `json:"ask_anon"`
	AskAnonPageTitle     string `json:"ask_page_title"`
	CanSendFanMail       bool   `json:"can_send_fan_mail"`
	CanSubmit            bool   `json:"can_submit"`
	CanSubscribe         bool   `json:"can_subscribe"`
	Description          string `json:"description"`
	Followed             bool   `json:"followed"`
	IsBlockedFromPrimary bool   `json:"is_blocked_from_primary"`
	IsNSFW               bool   `json:"is_nsfw"`
	ShareLikes           bool   `json:"share_likes"`
	SubmissionPageTitle  string `json:"submission_page_title"`
	Subscribed           bool   `json:"subscribed"`
	TotalPosts           int64  `json:"total_posts"`
	Updated              int64  `json:"updated"`
	UUID                 string `json:"uuid"`
}

// Convenience method converting a Blog into a JSON representation
func (b *Blog) String() string {
	return jsonStringify(*b)
}

// Retrieve information about a blog
func GetBlogInfo(client ClientInterface, name string) (*Blog, error) {
	response, err := client.Get(blogPath("/blog/%s/info", name))
	if err != nil {
		return nil, err
	}
	blog := struct {
		Response struct {
			Blog Blog `json:"blog"`
		} `json:"response"`
	}{}
	//blog := blogResponse{}
	err = json.Unmarshal(response.body, &blog)
	if err != nil {
		return nil, err
	}
	blog.Response.Blog.client = client
	return &blog.Response.Blog, nil
}

// Retrieve Blog's Avatar URI
func GetAvatar(client ClientInterface, name string) (string, error) {
	response, err := client.Get(blogPath("/blog/%s/avatar", name))
	if err != nil {
		return "", err
	}
	if location := response.Headers.Get("Location"); len(location) > 0 {
		return location, nil
	}
	if err = response.PopulateFromBody(); err != nil {
		return "", err
	}
	if l, ok := response.Result["location"]; ok {
		if location, ok := l.(string); ok {
			return location, nil
		}
	}
	return "", errors.New("Unable to detect avatar location")
}

// Create a BlogRef
func NewBlogRef(client ClientInterface, name string) *BlogRef {
	return &BlogRef{
		Name:   name,
		client: client,
	}
}

// Retrieves blog info for the given blog reference
func (b *BlogRef) GetInfo() (*Blog, error) {
	return GetBlogInfo(b.client, b.Name)
}

// Retrieves blog's posts for the given blog reference
func (b *BlogRef) GetPosts(params url.Values) (*Posts, error) {
	return GetPosts(b.client, b.Name, params)
}

// Retrieves name property
func (b *BlogRef) getName() string {
	return b.Name
}

// Helper function to allow for less verbose code
func normalizeBlogName(name string) string {
	if !strings.Contains(name, ".") {
		name = fmt.Sprintf("%s.tumblr.com", name)
	}
	return name
}

// Expects path to contain a single %s placeholder to be substituted with the result of normalizeBlogName
func blogPath(path, name string) string {
	return fmt.Sprintf(path, normalizeBlogName(name))
}
