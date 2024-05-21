package service

import (
	"ManyACG/adapter"
	"ManyACG/dao"
	es "ManyACG/errors"
	"ManyACG/model"
	"ManyACG/types"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateArtwork(ctx context.Context, artwork *types.Artwork) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if artworkModel != nil {
		return nil, es.ErrArtworkAlreadyExist
	}
	if dao.CheckDeletedByURL(ctx, artwork.SourceURL) {
		return nil, es.ErrArtworkDeleted
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		// 创建 Tag
		tagIDs := make([]primitive.ObjectID, len(artwork.Tags))
		for i, tag := range artwork.Tags {
			tagModel, err := dao.GetTagByName(ctx, tag)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if tagModel != nil {
				tagIDs[i] = tagModel.ID
				continue
			}
			tagModel = &model.TagModel{
				Name: tag,
			}
			tagRes, err := dao.CreateTag(ctx, tagModel)
			if err != nil {
				return nil, err
			}
			tagIDs[i] = tagRes.InsertedID.(primitive.ObjectID)
		}

		// 创建 Artist
		var artist_id primitive.ObjectID
		artistModel, err := dao.GetArtistByUID(ctx, artwork.Artist.UID)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if artistModel != nil {
			artistModel.Name = artwork.Artist.Name
			artist_id = artistModel.ID
		} else {
			artistModel = &model.ArtistModel{
				Type:     artwork.Artist.Type,
				UID:      artwork.Artist.UID,
				Username: artwork.Artist.Username,
				Name:     artwork.Artist.Name,
			}
			res, err := dao.CreateArtist(ctx, artistModel)
			if err != nil {
				return nil, err
			}
			artist_id = res.InsertedID.(primitive.ObjectID)
		}

		// 创建 Artwork
		artworkModel = &model.ArtworkModel{
			Title:       artwork.Title,
			Description: artwork.Description,
			R18:         artwork.R18,
			SourceType:  artwork.SourceType,
			SourceURL:   artwork.SourceURL,
			ArtistID:    artist_id,
			Tags:        tagIDs,
		}
		res, err := dao.CreateArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}

		// 创建 Picture
		pictureModels := make([]*model.PictureModel, len(artwork.Pictures))
		for i, picture := range artwork.Pictures {
			pictureModel := &model.PictureModel{
				Index:        picture.Index,
				ArtworkID:    res.InsertedID.(primitive.ObjectID),
				Thumbnail:    picture.Thumbnail,
				Original:     picture.Original,
				Width:        picture.Width,
				Height:       picture.Height,
				Hash:         picture.Hash,
				BlurScore:    picture.BlurScore,
				TelegramInfo: (*model.TelegramInfo)(picture.TelegramInfo),
				StorageInfo:  (*model.StorageInfo)(picture.StorageInfo),
			}
			pictureModels[i] = pictureModel
		}
		pictureRes, err := dao.CreatePictures(ctx, pictureModels)
		if err != nil {
			return nil, err
		}
		pictureIDs := make([]primitive.ObjectID, len(pictureRes.InsertedIDs))
		for i, id := range pictureRes.InsertedIDs {
			pictureIDs[i] = id.(primitive.ObjectID)
		}

		// 更新 Artwork 的 Pictures
		_, err = dao.UpdateArtworkPicturesByID(ctx, res.InsertedID.(primitive.ObjectID), pictureIDs)
		if err != nil {
			return nil, err
		}
		artworkModel, err = dao.GetArtworkByID(ctx, res.InsertedID.(primitive.ObjectID))
		if err != nil {
			return nil, err
		}
		return artworkModel, nil
	})
	if err != nil {
		return nil, err
	}
	artwork.CreatedAt = result.(*model.ArtworkModel).CreatedAt.Time()
	return artwork, nil
}

func GetArtworkByURL(ctx context.Context, sourceURL string) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel)
}

func GetArtworkByMessageID(ctx context.Context, messageID int) (*types.Artwork, error) {
	pictureModel, err := dao.GetPictureByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	artworkModel, err := dao.GetArtworkByID(ctx, pictureModel.ArtworkID)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel)
}

func GetArtworkByID(ctx context.Context, id primitive.ObjectID) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel)
}

func GetArtworkIDByPicture(ctx context.Context, picture *types.Picture) (primitive.ObjectID, error) {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return pictureModel.ArtworkID, nil
}

func GetRandomArtworks(ctx context.Context, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	artworkModels, err := dao.GetArtworksByR18(ctx, r18, int64(limit))
	if err != nil {
		return nil, err
	}
	artworks := make([]*types.Artwork, len(artworkModels))
	errChan := make(chan error, len(artworkModels))
	for i, artworkModel := range artworkModels {
		go func(i int, artworkModel *model.ArtworkModel) {
			artistModel, err := dao.GetArtistByID(ctx, artworkModel.ArtistID)
			if err != nil {
				errChan <- err
				return
			}
			tags := make([]string, len(artworkModel.Tags))
			for j, tagID := range artworkModel.Tags {
				tagModel, err := dao.GetTagByID(ctx, tagID)
				if err != nil {
					errChan <- err
					return
				}
				tags[j] = tagModel.Name
			}
			pictures := make([]*types.Picture, len(artworkModel.Pictures))
			for j, pictureID := range artworkModel.Pictures {
				pictureModel, err := dao.GetPictureByID(ctx, pictureID)
				if err != nil {
					errChan <- err
					return
				}
				pictures[j] = pictureModel.ToPicture()
			}
			artworks[i] = &types.Artwork{
				Title:       artworkModel.Title,
				Description: artworkModel.Description,
				R18:         artworkModel.R18,
				CreatedAt:   artworkModel.CreatedAt.Time(),
				SourceType:  artworkModel.SourceType,
				SourceURL:   artworkModel.SourceURL,
				Artist:      artistModel.ToArtist(),
				Tags:        tags,
				Pictures:    pictures,
			}
			errChan <- nil
		}(i, artworkModel)
	}
	for range artworkModels {
		err := <-errChan
		if err != nil {
			return nil, err
		}
	}
	return artworks, nil
}

// 通过标签获取作品, 标签名使用全字匹配
//
// tags: 二维数组, tags = [["tag1", "tag2"], ["tag3", "tag4"]] 表示 (tag1 || tag2) && (tag3 || tag4)
func GetArtworksByTags(ctx context.Context, tags [][]string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	if len(tags) == 0 {
		return GetRandomArtworks(ctx, r18, limit)
	}
	tagIDs := make([][]primitive.ObjectID, len(tags))
	for i, tagGroup := range tags {
		tagIDs[i] = make([]primitive.ObjectID, len(tagGroup))
		for j, tagName := range tagGroup {
			tagModel, err := dao.GetTagByName(ctx, tagName)
			if err != nil {
				return nil, err
			}
			tagIDs[i][j] = tagModel.ID
		}
	}
	artworkModels, err := dao.GetArtworksByTags(ctx, tagIDs, r18, int64(limit))
	if err != nil {
		return nil, err
	}
	artworks := make([]*types.Artwork, len(artworkModels))
	for i, artworkModel := range artworkModels {
		artworks[i], err = adapter.ConvertToArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
	}
	return artworks, nil
}

func UpdateArtworkR18ByURL(ctx context.Context, sourceURL string, r18 bool) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	_, err = dao.UpdateArtworkR18ByID(ctx, artworkModel.ID, r18)
	if err != nil {
		return err
	}
	return nil
}

func DeleteArtworkByURL(ctx context.Context, sourceURL string) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err := dao.DeleteArtworkByID(ctx, artworkModel.ID)
		if err != nil {
			return nil, err
		}

		_, err = dao.DeletePicturesByArtworkID(ctx, artworkModel.ID)
		if err != nil {
			return nil, err
		}

		_, err = dao.CreateDeleted(ctx, &model.DeletedModel{
			SourceURL: sourceURL,
			ArtworkID: artworkModel.ID,
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
