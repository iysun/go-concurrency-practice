package main

import (
	"net/url"
	"regexp"
	"strings"
)

// extractLinks 使用正则表达式从 HTML 中提取所有链接
func extractLinks(html, orgUrl string) []string {
	// 正则表达式匹配 <a> 标签中的 href 属性
	// 支持双引号、单引号、无引号三种情况
	re := regexp.MustCompile(`<a\s+[^>]*?href\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))[^>]*>`)

	matches := re.FindAllStringSubmatch(html, -1)

	var links []string
	seen := make(map[string]bool)

	// 解析 base URL，用于处理相对路径
	baseURL, err := url.Parse(orgUrl)
	if err != nil {
		return links
	}

	for _, match := range matches {
		// match[1] 双引号中的值, match[2] 单引号中的值, match[3] 无引号的值
		href := ""
		for i := 1; i <= 3; i++ {
			if match[i] != "" {
				href = strings.TrimSpace(match[i])
				break
			}
		}

		// 过滤无效链接
		if !isValidLink(href) {
			continue
		}

		// 解析并规范化 URL
		absoluteURL := resolveURL(href, baseURL)
		if absoluteURL == "" {
			continue
		}

		// 只保留同域名的链接
		if !isSameDomain(absoluteURL, baseURL) {
			continue
		}

		// 去重
		if !seen[absoluteURL] {
			links = append(links, absoluteURL)
			seen[absoluteURL] = true
		}
	}

	return links
}

// isValidLink 检查链接是否有效
func isValidLink(href string) bool {
	if href == "" {
		return false
	}

	// 过滤特殊协议和锚点
	invalidPrefixes := []string{"#", "javascript:", "mailto:", "tel:", "data:", "file:"}
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(strings.ToLower(href), prefix) {
			return false
		}
	}

	return true
}

// resolveURL 解析相对 URL 为绝对 URL
func resolveURL(href string, base *url.URL) string {
	parsedURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// 将相对 URL 解析为绝对 URL
	absoluteURL := base.ResolveReference(parsedURL)

	// 清理 URL：移除 fragment、query string、尾部斜杠等
	absoluteURL.Fragment = ""
	// absoluteURL.RawQuery = ""  // 可选：是否需要保留查询参数

	// 标准化路径
	if absoluteURL.Path == "" {
		absoluteURL.Path = "/"
	}

	// 可选：将路径转为小写
	// absoluteURL.Path = strings.ToLower(absoluteURL.Path)

	return absoluteURL.String()
}

// isSameDomain 检查 URL 是否属于同一域名
func isSameDomain(link string, baseURL *url.URL) bool {
	parsedLink, err := url.Parse(link)
	if err != nil {
		return false
	}

	return parsedLink.Host == baseURL.Host
}
