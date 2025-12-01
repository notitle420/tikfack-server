package dmmapi

// JSON レスポンス構造体

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Actress struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Maker struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Director struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Series struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ... Maker, Series, Director も同様の構造体を定義

type Item struct {
	ContentID string `json:"content_id"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	URL       string `json:"URL"`
	ImageURL  struct {
		Large string `json:"large"`
	} `json:"imageURL"`
	SampleMovieURL *struct {
		Size720480 string `json:"size_720_480"`
	} `json:"sampleMovieURL,omitempty"`
	Prices struct {
		Price      string `json:"price,omitempty"`
		Deliveries *struct {
			Delivery []struct {
				Type      string `json:"type"`
				Price     string `json:"price"`
				ListPrice string `json:"list_price"`
			}
		} `json:"deliveries,omitempty"`
	} `json:"prices"`
	Review *struct {
		Count   int    `json:"count"`
		Average string `json:"average"` // JSONではstringとして返ってくる
	} `json:"review,omitempty"`
	ItemInfo struct {
		Actress  []Actress  `json:"actress,omitempty"`
		Genre    []Genre    `json:"genre,omitempty"`
		Maker    []Maker    `json:"maker,omitempty"`
		Series   []Series   `json:"series,omitempty"`
		Director []Director `json:"director,omitempty"`
	} `json:"iteminfo"`
}

type Result struct {
	Status        int    `json:"status"`
	ResultCount   int    `json:"result_count"`
	TotalCount    int    `json:"total_count"`
	FirstPosition int    `json:"first_position"`
	Items         []Item `json:"items"`
}

type Response struct {
	Result Result `json:"result"`
}
