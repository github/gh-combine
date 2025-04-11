package github

type Ref struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type Label struct {
	Name string `json:"name"`
}

type Labels []Label

type Pull struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Head   Ref    `json:"head"`
	Base   Ref    `json:"base"`
	Labels Labels `json:"labels"`
}

type Pulls []Pull
