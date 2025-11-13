package bilibili_crawler

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseVideoInfo 解析视频信息
func parseVideoInfo(html string, videoURL string) (*VideoInfo, error) {
	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	info := &VideoInfo{URL: videoURL}

	// 从URL中提取BV号
	bvPattern := regexp.MustCompile(`BV[a-zA-Z0-9]+`)
	if match := bvPattern.FindString(videoURL); match != "" {
		info.BV = match
	}

	// 提取标题
	if title := doc.Find("title").Text(); title != "" {
		info.Title = strings.TrimSpace(regexp.MustCompile(`_哔哩哔哩_bilibili`).ReplaceAllString(title, ""))
	}

	// 提取标签
	if keywords, exists := doc.Find("meta[itemprop='keywords']").Attr("content"); exists {
		contentWithoutTitle := strings.Replace(keywords, info.Title+",", "", 1)
		keywordsList := strings.Split(contentWithoutTitle, ",")
		if len(keywordsList) > 4 {
			info.Tags = strings.Join(keywordsList[:len(keywordsList)-4], ",")
		} else {
			info.Tags = strings.Join(keywordsList, ",")
		}
	}

	// 提取meta description中的数据
	if metaDesc, exists := doc.Find("meta[itemprop='description']").Attr("content"); exists {
		// 提取播放量、弹幕量、点赞数、投币数、收藏数、转发数
		numbersPattern := regexp.MustCompile(`视频播放量 (\d+)、弹幕量 (\d+)、点赞数 (\d+)、投硬币枚数 (\d+)、收藏人数 (\d+)、转发人数 (\d+)`)
		if match := numbersPattern.FindStringSubmatch(metaDesc); len(match) > 6 {
			info.Views, _ = strconv.Atoi(match[1])
			info.Danmaku, _ = strconv.Atoi(match[2])
			info.Likes, _ = strconv.Atoi(match[3])
			info.Coins, _ = strconv.Atoi(match[4])
			info.Favorites, _ = strconv.Atoi(match[5])
			info.Shares, _ = strconv.Atoi(match[6])
		}

		// 提取作者名称
		authorPattern := regexp.MustCompile(`视频作者\s*([^,]+)`)
		if match := authorPattern.FindStringSubmatch(metaDesc); len(match) > 1 {
			info.Author = strings.TrimSpace(match[1])
		}

		// 提取作者简介
		authorDescPattern := regexp.MustCompile(`作者简介\s*([^,]+)`)
		if match := authorDescPattern.FindStringSubmatch(metaDesc); len(match) > 1 {
			info.AuthorDesc = strings.TrimSpace(match[1])
		}

		// 提取视频简介
		metaParts := strings.Split(metaDesc, ",")
		if len(metaParts) > 0 {
			info.VideoDesc = strings.TrimSpace(metaParts[0])
		}
	}

	// 提取发布时间
	if publishDate, exists := doc.Find("meta[itemprop='uploadDate']").Attr("content"); exists {
		info.PublishDate = publishDate
	}

	// 从script标签中提取更多信息
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		// 提取 window.__playinfo__ 中的时长信息
		if strings.Contains(text, "window.__playinfo__") {
			durationPattern := regexp.MustCompile(`"duration":(\d+)`)
			if match := durationPattern.FindStringSubmatch(text); len(match) > 1 {
				duration, _ := strconv.Atoi(match[1])
				info.Duration = duration
			}
		}

		// 提取 window.__INITIAL_STATE__ 中的信息
		if strings.Contains(text, "window.__INITIAL_STATE__") {
			// 提取作者ID
			authorIDPattern := regexp.MustCompile(`"mid":(\d+)`)
			if match := authorIDPattern.FindStringSubmatch(text); len(match) > 1 {
				info.AuthorID = match[1]
			}

			// 提取视频AID
			aidPattern := regexp.MustCompile(`"aid":(\d+)`)
			if match := aidPattern.FindStringSubmatch(text); len(match) > 1 {
				info.Aid = match[1]
			}
		}
	})

	return info, nil
}

// isURL 判断是否为URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// isBVID 判断是否为BV号
func isBVID(s string) bool {
	matched, _ := regexp.MatchString(`^BV[a-zA-Z0-9]+$`, s)
	return matched
}

// isAVID 判断是否为AV号
func isAVID(s string) bool {
	matched, _ := regexp.MatchString(`^\d+$`, s)
	return matched
}

// FormatDuration 将秒数格式化为时:分:秒
func FormatDuration(seconds int) string {
	if seconds == 0 {
		return "未知"
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return strings.TrimSpace(strconv.Itoa(hours) + ":" + padZero(minutes) + ":" + padZero(secs))
	}
	return strings.TrimSpace(strconv.Itoa(minutes) + ":" + padZero(secs))
}

// padZero 补零
func padZero(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// FormatNumber 格式化数字（添加千分位）
func FormatNumber(n int) string {
	if n == 0 {
		return "0"
	}
	if n >= 10000 {
		return strings.TrimSpace(strconv.FormatFloat(float64(n)/10000, 'f', 1, 64) + "万")
	}

	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}

	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
