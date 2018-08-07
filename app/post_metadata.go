// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"image"
	"io"
	"strings"

	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils/markdown"
)

func (a *App) PreparePostListForClient(originalList *model.PostList) (*model.PostList, *model.AppError) {
	list := &model.PostList{
		Posts: make(map[string]*model.Post),
		Order: originalList.Order,
	}

	for id, originalPost := range originalList.Posts {
		post, err := a.PreparePostForClient(originalPost)
		if err != nil {
			return originalList, err
		}

		list.Posts[id] = post
	}

	return list, nil
}

func (a *App) PreparePostForClient(originalPost *model.Post) (*model.Post, *model.AppError) {
	post := originalPost.Clone()

	var err *model.AppError

	needReactionCounts := post.ReactionCounts == nil
	needEmojis := post.Emojis == nil
	needImageDimensions := post.ImageDimensions == nil
	needOpenGraphData := post.OpenGraphData == nil

	var reactions []*model.Reaction
	if needReactionCounts || needEmojis {
		reactions, err = a.GetReactionsForPost(post.Id)
		if err != nil {
			return post, err
		}
	}

	if needReactionCounts {
		post.ReactionCounts = model.CountReactions(reactions)
	}

	if post.FileInfos == nil {
		fileInfos, err := a.GetFileInfosForPost(post.Id, false)
		if err != nil {
			return post, err
		}

		post.FileInfos = fileInfos
	}

	if needEmojis {
		emojis, err := a.getCustomEmojisForPost(post.Message, reactions)
		if err != nil {
			return post, err
		}

		post.Emojis = emojis
	}

	post = a.PostWithProxyAddedToImageURLs(post)

	if needImageDimensions || needOpenGraphData {
		link := getFirstLinkInString(post.Message)

		og, dimensions, err := getLinkMetadata(link)
		if err != nil {
			return post, err
		}

		if needImageDimensions {
			if dimensions != nil {
				post.ImageDimensions = []*model.PostImageDimensions{dimensions}
			} else {
				post.ImageDimensions = []*model.PostImageDimensions{}
			}
		}

		if needOpenGraphData {
			if og != nil {
				post.OpenGraphData = []*opengraph.OpenGraph{og}
			} else {
				post.OpenGraphData = []*opengraph.OpenGraph{}
			}
		}
	}

	return post, nil
}

func (a *App) getCustomEmojisForPost(message string, reactions []*model.Reaction) ([]*model.Emoji, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCustomEmoji {
		// Only custom emoji are returned
		return []*model.Emoji{}, nil
	}

	names := model.EMOJI_PATTERN.FindAllString(message, -1)

	for _, reaction := range reactions {
		names = append(names, reaction.EmojiName)
	}

	if len(names) == 0 {
		return []*model.Emoji{}, nil
	}

	names = model.RemoveDuplicateStrings(names)

	for i, name := range names {
		names[i] = strings.Trim(name, ":")
	}

	return a.GetMultipleEmojiByName(names)
}

func getFirstLinkInString(str string) string {
	var url string

	markdown.Inspect(message, func(blockOrInline interface{}) bool {
		switch v := blockOrInline.(type) {
		case *markdown.ReferenceLink:
			url = v.ReferenceDefinition.Desination()
			return false
		case *markdown.InlineLink:
			url = v.Desination()
			return false
		case *markdown.Autolink:
			url = v.Desination()
			return false
		default:
			return true
		}
	})

	return url
}

func getLinkMetadata(requestURL string) (*opengraph.OpenGraph, *PostImageDimensions, error) {
	res, err := a.HTTPClient(false).Get(requestURL)
	if err != nil {
		return nil, nil, err
	}
	defer consumeAndClose(res)

	return parseLinkMetadata(requestURL, res.Body, res.Header.Get("Content-Type"))
}

func parseLinkMetadata(requestURL string, body io.Reader, contentType string) (*opengraph.OpenGraph, *model.PostImageDimensions, error) {
	if strings.HasPrefix(contentType, "image") {
		dimensions, err := parseImageDimensions(requestURL, body)
		return nil, dimensions, err
	} else if strings.HasPrefix(contentType, "text/html") {
		return ParseOpenGraphMetadata(requestURL, body, contentType), nil, nil
	} else {
		// Not an image or web page with OpenGraph information
		return nil, nil, nil
	}
}

func parseImageDimensions(requestURL string, body io.Reader) (*model.PostImageDimensions, error) {
	config, _, err := image.DecodeConfig(body)
	if err != nil {
		return nil, err
	}

	dimensions := &model.ImageDimensions{
		URL:    requestURL,
		Width:  config.Width,
		Height: config.Height,
	}

	return dimensions, nil
}
