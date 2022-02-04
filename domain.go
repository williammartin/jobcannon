package jobcannon

// Some redundancy in these types but it keeps things simple
// to store single blobs with redundancy rather than references.

type CatalogId int
type JobId int

type ExpressionOfInterest struct {
	CatalogId  CatalogId
	JobId      JobId
	By         string
	Text       string
	Interested bool
}

type Job struct {
	Id   JobId
	By   string
	Text string
}

type Catalog struct {
	Id     CatalogId
	JobIds []JobId
}
