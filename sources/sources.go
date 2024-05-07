package sources

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/config"
	"ManyACG-Bot/errors"
	"ManyACG-Bot/sources/pixiv"
	"ManyACG-Bot/sources/twitter"
	"ManyACG-Bot/types"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	Sources          = make(map[types.SourceType]Source)
	SourceURLRegexps = make(map[types.SourceType]*regexp.Regexp)
)

func InitSources() {
	if config.Cfg.Source.Pixiv.Enable {
		Sources[types.SourceTypePixiv] = new(pixiv.Pixiv)
		Sources[types.SourceTypePixiv].Init()
	}
	if config.Cfg.Source.Twitter.Enable {
		Sources[types.SourceTypeTwitter] = new(twitter.Twitter)
		Sources[types.SourceTypeTwitter].Init()
	}

	for sourceType, source := range Sources {
		SourceURLRegexps[sourceType] = source.GetSourceURLRegexp()
	}
}

func GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	for k, v := range SourceURLRegexps {
		if v.MatchString(sourceURL) {
			if Sources[k] != nil {
				return Sources[k].GetArtworkInfo(sourceURL)
			}
		}
	}
	return nil, errors.ErrSourceNotSupported
}

func GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	fileName := ""
	switch artwork.SourceType {
	case types.SourceTypePixiv:
		fileName = artwork.Title + "_" + filepath.Base(picture.Original)
	case types.SourceTypeTwitter:
		original := picture.Original
		urlSplit := strings.Split(picture.Original, "?")
		if len(urlSplit) > 1 {
			original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
		}
		tweetID := strings.Split(artwork.SourceURL, "/")[len(strings.Split(artwork.SourceURL, "/"))-1]
		fileName = tweetID + "_" + strconv.Itoa(int(picture.Index)) + filepath.Ext(original)
	default:
		fileName = artwork.Title + "_" + filepath.Base(picture.Original)
	}
	return common.ReplaceFileNameInvalidChar(fileName)
}

func FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for sourceType, reg := range SourceURLRegexps {
		if url := reg.FindString(text); url != "" {
			return Sources[sourceType].GetCommonSourceURL(url)
		}
	}
	return ""
}

// MatchesSourceURL returns whether the text contains a source URL.
func MatchesSourceURL(text string) bool {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, reg := range SourceURLRegexps {
		if reg.MatchString(text) {
			return true
		}
	}
	return false
}

func GetPixivRegularURL(original string) string {
	photoURL := strings.Replace(original, "img-original", "img-master", 1)
	photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
	photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
	return photoURL
}
