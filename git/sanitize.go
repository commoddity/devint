package git

import (
	"regexp"
	"strings"
)

// sanitizeDiff removes identifying information and PII from diff content before sending to LLM.
// It removes or redacts:
//   - Company/organization names (case-insensitive, configurable via companyName parameter)
//   - Email addresses
//   - Domain names and URLs
//   - IP addresses
//   - API keys and tokens (common patterns)
//   - File paths containing identifying terms
func sanitizeDiff(diff string, companyName string) string {
	result := diff

	// Company/organization name variations (case-insensitive)
	// Build regex pattern dynamically from config value to match all variations
	// This will match variations like "companyname", "CompanyName", "COMPANYNAME", "company-name", "company_name", etc.
	if companyName != "" {
		// First, do a comprehensive case-insensitive string replacement as a baseline
		// This catches the exact name in any case
		lowerCompanyName := strings.ToLower(companyName)
		result = replaceCompanyNameCaseInsensitive(result, lowerCompanyName)

		// Split the name into parts if it has separators, or treat as single word
		nameParts := strings.FieldsFunc(lowerCompanyName, func(r rune) bool {
			return r == '-' || r == '_' || r == ' '
		})

		var pattern string
		if len(nameParts) > 1 {
			// Multiple word parts: match with optional separators between them
			// Escape each part to handle regex special characters
			escapedParts := make([]string, len(nameParts))
			for i, part := range nameParts {
				escapedParts[i] = regexp.QuoteMeta(part)
			}
			separatorPattern := `[\s_-]?`
			pattern = `(?i)(` + strings.Join(escapedParts, separatorPattern) + `)`
		} else {
			// Single word: escape and match as-is
			escapedName := regexp.QuoteMeta(lowerCompanyName)
			pattern = `(?i)(` + escapedName + `)`
		}

		// Apply pattern to catch any remaining variations
		companyNamePattern := regexp.MustCompile(pattern)
		result = companyNamePattern.ReplaceAllString(result, "[COMPANY]")

		// Also handle file paths containing the company name
		// Use the same flexible pattern to match variations in file paths
		filePathPattern := regexp.MustCompile(`(?i)(Old File:|New File:)\s*([^\s]*` + pattern + `[^\s]*)`)
		result = filePathPattern.ReplaceAllStringFunc(result, func(match string) string {
			parts := strings.SplitN(match, " ", 2)
			if len(parts) == 2 {
				return parts[0] + " [REDACTED_PATH]"
			}
			return match
		})
	}

	// Email addresses
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	result = emailPattern.ReplaceAllString(result, "[EMAIL]")

	// URLs (http/https/ftp) - process before domain names since URLs contain domains
	urlPattern := regexp.MustCompile(`(?i)(https?|ftp)://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	result = urlPattern.ReplaceAllString(result, "[URL]")

	// Domain names (but preserve common non-identifying domains)
	// Match domains like example.com, api.example.com, etc.
	domainPattern := regexp.MustCompile(`(?i)\b([a-z0-9](?:[a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,}\b`)
	// Exclude common non-identifying domains that are safe to show
	excludedDomains := regexp.MustCompile(`(?i)(github\.com|gitlab\.com|bitbucket\.org|npmjs\.com|pypi\.org|docker\.com|example\.com|localhost|127\.0\.0\.1|test\.com|example\.org)`)

	lines := strings.Split(result, "\n")
	for i, line := range lines {
		// Skip if it's a markdown code block marker
		if strings.HasPrefix(line, "```") {
			continue
		}
		// Skip file path lines (they will be handled separately)
		if strings.HasPrefix(line, "Old File:") || strings.HasPrefix(line, "New File:") {
			continue
		}

		// Find all domain matches in this line
		matches := domainPattern.FindAllString(line, -1)
		for _, match := range matches {
			// Only replace if it's not in the excluded list
			if !excludedDomains.MatchString(match) {
				lines[i] = strings.ReplaceAll(lines[i], match, "[DOMAIN]")
			}
		}
	}
	result = strings.Join(lines, "\n")

	// IP addresses (IPv4 and IPv6)
	ipv4Pattern := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	result = ipv4Pattern.ReplaceAllString(result, "[IP]")
	ipv6Pattern := regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)
	result = ipv6Pattern.ReplaceAllString(result, "[IP]")

	// API keys and tokens (common patterns)
	// GitHub tokens (ghp_, gho_, ghu_, ghs_, ghr_, github_pat_)
	apiKeyPattern := regexp.MustCompile(`\b(ghp_|gho_|ghu_|ghs_|ghr_|github_pat_)[a-zA-Z0-9_]{36,}\b`)
	result = apiKeyPattern.ReplaceAllString(result, "[API_KEY]")

	// Generic API key patterns (long alphanumeric strings that might be keys)
	// Match strings like "sk_live_...", "pk_test_...", etc.
	genericKeyPattern := regexp.MustCompile(`\b(sk_|pk_|api[_-]?key[=:]?\s*)[a-zA-Z0-9_\-]{20,}\b`)
	result = genericKeyPattern.ReplaceAllString(result, "[API_KEY]")

	// AWS access keys
	awsKeyPattern := regexp.MustCompile(`\b(AKIA[0-9A-Z]{16}|aws_access_key_id[=:]?\s*)[a-zA-Z0-9+/=]{20,}\b`)
	result = awsKeyPattern.ReplaceAllString(result, "[AWS_KEY]")

	// Replace other common identifying patterns in file paths
	// Paths with company-specific directories
	sensitivePathPattern := regexp.MustCompile(`(?i)(Old File:|New File:)\s*[^\s]*(internal|private|secret|confidential|proprietary)[^\s]*`)
	result = sensitivePathPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, " ", 2)
		if len(parts) == 2 {
			// Keep just the filename if possible, otherwise redact
			path := parts[1]
			if idx := strings.LastIndex(path, "/"); idx != -1 {
				filename := path[idx+1:]
				return parts[0] + " [REDACTED_DIR]/" + filename
			}
			return parts[0] + " [REDACTED_PATH]"
		}
		return match
	})

	return result
}

// replaceCompanyNameCaseInsensitive performs case-insensitive replacement of company name
// throughout the string. This ensures we catch all variations including as part of compound identifiers.
func replaceCompanyNameCaseInsensitive(text, companyName string) string {
	// Build a regex that matches the company name case-insensitively in any context
	// This catches all variations including as part of compound identifiers
	escapedName := regexp.QuoteMeta(companyName)

	// Match the company name case-insensitively, including when it's part of compound identifiers
	// This pattern will match:
	// - Standalone
	// - With prefix
	// - With suffix
	// - Both with prefix and suffix
	pattern := regexp.MustCompile(`(?i)` + escapedName)

	// Replace all occurrences
	result := pattern.ReplaceAllString(text, "[COMPANY]")

	return result
}

// SanitizePrompt removes identifying information and PII from prompt content before sending to LLM.
// This reuses the sanitizeDiff function but is intended for sanitizing complete prompts that may contain
// PR titles, descriptions, and other text in addition to diffs.
// It uses the company name configured in the Provider.
func (p *Provider) SanitizePrompt(prompt string) string {
	companyName := ""
	if p != nil {
		companyName = p.companyName
	}
	return sanitizeDiff(prompt, companyName)
}
