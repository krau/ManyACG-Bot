package types

type Permission string

const (
	PostArtwork   Permission = "post_artwork"
	DeleteArtwork Permission = "delete_artwork"
	DeletePicture Permission = "delete_picture"
	FetchArtwork  Permission = "fetch_artwork"
)