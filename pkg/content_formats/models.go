package content_formats

const (
	QUERY_GET_FORMATS = "SELECT * FROM content_formats"
)

type ContentFormat struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}
