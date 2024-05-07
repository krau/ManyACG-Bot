package pixiv

import (
	"regexp"

	"github.com/imroc/req/v3"
)

var ReqClient *req.Client

var (
	pixivSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`pixiv\.net/(?:artworks/|i/|member_illust\.php\?(?:[\w=&]*\&|)illust_id=)(\d+)`)
	numberRegexp         *regexp.Regexp = regexp.MustCompile(`\d+`)
)
