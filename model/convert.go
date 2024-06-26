package model

import "ManyACG/types"

func (picture *PictureModel) ToPicture() *types.Picture {
	return &types.Picture{
		Index:        picture.Index,
		Thumbnail:    picture.Thumbnail,
		Original:     picture.Original,
		Width:        picture.Width,
		Height:       picture.Height,
		Hash:         picture.Hash,
		BlurScore:    picture.BlurScore,
		TelegramInfo: (*types.TelegramInfo)(picture.TelegramInfo),
		StorageInfo:  (*types.StorageInfo)(picture.StorageInfo),
	}
}

func (artist *ArtistModel) ToArtist() *types.Artist {
	return &types.Artist{
		Name:     artist.Name,
		Type:     artist.Type,
		UID:      artist.UID,
		Username: artist.Username,
	}
}
