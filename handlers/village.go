package handlers

type Village struct {
	ID    string `bson:"_id"`
	Name  string
	Likes int32 `json:"likes,omitempty"`
}
