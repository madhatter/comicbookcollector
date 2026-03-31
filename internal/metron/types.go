package metron

type IssueSearchResponse struct {
	Count   int     `json:"count"`
	Results []Issue `json:"results"`
}

type Issue struct {
	ID          int         `json:"id"`
	Series      Series      `json:"series"`
	Number      string      `json:"number"`
	CoverDate   string      `json:"cover_date"`
	StoreDate   string      `json:"store_date"`
	Description string      `json:"desc"`
	Price       string      `json:"price"`
	PageCount   int         `json:"page_count"`
	UPC         string      `json:"upc"`
	Credits     []Credit    `json:"credits"`
	Characters  []Character `json:"characters"`
	Teams       []Team      `json:"teams"`
	Arcs        []Arc       `json:"arcs"`
}

type Credit struct {
	Creator Creator `json:"creator"`
	Role    []Role  `json:"role"`
}

type Creator struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Character struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Arc struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Series struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Volume int    `json:"volume"`
}
