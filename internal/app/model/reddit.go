package model

type RedditResponse struct {
	Memes []RedditMeme `json:"memes"`
}

type RedditMeme struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Author string `json:"author"`
}
