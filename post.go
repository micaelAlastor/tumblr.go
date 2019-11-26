package tumblr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
)

// Posts represents a list of MiniPosts, which have a minimal set of information.
type Posts struct {
	client     ClientInterface
	response   Response
	Posts      []Post `json:"posts"`
	TotalPosts int64  `json:"total_posts"`
}

// MiniPost stores the basics for what is needed in a Post.
type MiniPost struct {
	Id        uint64 `json:"id"`
	Type      string `json:"type"`
	BlogName  string `json:"blog_name"`
	ReblogKey string `json:"reblog_key"`
}

// PostRef is a base struct used as a starting point for performing operations on a post.
type PostRef struct {
	MiniPost
	client ClientInterface
}

// Post holds the common fields of any post type.
type Post struct {
	PostRef
	Body             string        `json:"body"`
	CanLike          bool          `json:"can_like"`
	CanReblog        bool          `json:"can_reblog"`
	CanReply         bool          `json:"can_reply"`
	CanSendInMessage bool          `json:"can_send_in_message"`
	Caption          string        `json:"caption"`
	Date             string        `json:"date"`
	DisplayAvatar    bool          `json:"display_avatar"`
	Followed         bool          `json:"followed"`
	Format           string        `json:"format"`
	Highlighted      []interface{} `json:"highlighted"`
	Liked            bool          `json:"liked"`
	NoteCount        uint64        `json:"note_count"`
	PermalinkUrl     string        `json:"permalink_url"`
	PostUrl          string        `json:"post_url"`
	Reblog           struct {
		Comment  string `json:"comment"`
		TreeHTML string `json:"tree_html"`
	} `json:"reblog"`
	Notes             []Note   `json:"notes"`
	RecommendedColor  string   `json:"recommended_color"`
	RecommendedSource bool     `json:"recommended_source"`
	ShortUrl          string   `json:"short_url"`
	Slug              string   `json:"slug"`
	SourceTitle       string   `json:"source_title"`
	SourceUrl         string   `json:"source_url"`
	State             string   `json:"state"`
	Summary           string   `json:"summary"`
	Tags              []string `json:"tags"`
	Timestamp         uint64   `json:"timestamp"`
	FeaturedTimestamp uint64   `json:"featured_timestamp,omitempty"`
	TrackName         string   `json:"track_name,omitempty"`

	//NPF posts
	Content []NpfContent `json:"content"`
	Trail   []NpfTrail   `json:"trail"`
}

type BlogMiniInfo struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Updated     uint64 `json:"updated"`
	UUID        string `json:"uuid"`
}

type Note struct {
	Type        string `json:"type"`
	Timestamp   uint64 `json:"timestamp"`
	BlogName    string `json:"blog_name"`
	BlogUUID    string `json:"blog_uuid"`
	BlogUrl     string `json:"blog_url"`
	Followed    bool   `json:"followed"`
	AvatarShape string `json:"avatar_shape"`
	//for reply notes
	ReplyText string `json:"reply_text"`
	//for reblog notes
	PostID               string `json:"post_id"`
	ReblogParentBlogName string `json:"reblog_parent_blog_name"`
}

type NpfContent struct {
	//general
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	//for text content
	Text       string       `json:"text"`
	Formatting []Formatting `json:"formatting"`
	//for media content
	Media   NpfMediaContainer `json:"media"`
	AltText string            `json:"alt_text"`
	//for link content
	NpfLink
	//poster can be either Media or array of so we omit it for now
	Poster NpfMediaContainer `json:"poster"`
	//audio and video ignored
}

type NpfLink struct {
	Url         string `json:"url"`
	DisplayUrl  string `json:"display_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	SiteName    string `json:"site_name"`
}

type Formatting struct {
	Type string `json:"type"`
	//for link type formatting
	Url string `json:"url"`
	//for mention type formatting
	Blog BlogMiniInfo `json:"blog"`
}

type NpfMediaContainer struct {
	Media           NpfMedia
	MediaCollection []NpfMedia
	IsArray         bool
}

func (n *NpfMediaContainer) UnmarshalJSON(data []byte) error {
	switch data[0] {
	case '[':
		mediaCollection := make([]NpfMedia, 0)
		if err := json.Unmarshal(data, &mediaCollection); err != nil {
			return err
		}
		n.MediaCollection = mediaCollection
		n.IsArray = true
	case '{':
		media := NpfMedia{}
		if err := json.Unmarshal(data, &media); err != nil {
			return err
		}
		n.Media = media
		n.IsArray = false
	default:
		return errors.New("unexpected char or whatever")
	}

	return nil
}

type NpfMedia struct {
	Type   string `json:"type"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type NpfTrail struct {
	Post       TrailPost    `json:"post"`
	Blog       BlogMiniInfo `json:"blog"`
	Content    []NpfContent `json:"content"`
	BrokenBlog BrokenBlog   `json:"broken_blog"` //in case trail post is unavailable. instead of post basically
}

type TrailPost struct {
	Id string `json:"id"`
}

type BrokenBlog struct {
	Name string `json:"name"`
}

// String returns the Post as a JSON string.
func (p *Post) String() string {
	return jsonStringify(*p)
}

// GetProperty uses reflection to retrieve one-off field values.
func (p *Post) GetProperty(key string) (interface{}, error) {
	if field, exists := reflect.TypeOf(p).Elem().FieldByName(key); exists {
		return reflect.ValueOf(p).Elem().FieldByIndex(field.Index), nil
	}
	return nil, errors.New(fmt.Sprintf("Property %s does not exist", key))
}

// GetSelf returns the Post from a PostInterface.
func (p *Post) GetSelf() *Post {
	return p
}

// helper method for querying a given path which should return a list of posts
func queryPosts(client ClientInterface, path, name string, params url.Values) (*Posts, error) {
	response, err := client.GetWithParams(blogPath(path, name), params)
	if err != nil {
		return nil, err
	}
	posts := struct {
		Response Posts `json:"response"`
	}{}
	if err = json.Unmarshal(response.body, &posts); err == nil {
		posts.Response.response = response
		posts.Response.client = client
		// store
		return &posts.Response, nil
	}
	return nil, err
}

// GetPosts retrieves a blog's posts, in the API docs you can find how to filter by ID, type, etc.
func GetPosts(client ClientInterface, name string, params url.Values) (*Posts, error) {
	return queryPosts(client, "/blog/%s/posts", name, params)
}

// SetClient sets the client member of the PostRef.
func (r *PostRef) SetClient(c ClientInterface) {
	r.client = c
}
