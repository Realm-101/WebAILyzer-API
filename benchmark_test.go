package wappalyzer

import (
	"net/http"
	"testing"
)

func BenchmarkFingerprint(b *testing.B) {
	wappalyzer, err := New()
	if err != nil {
		b.Fatalf("Failed to create wappalyzer: %v", err)
	}

	headers := map[string][]string{
		"Server":         {"Apache/2.4.29"},
		"X-Powered-By":   {"PHP/7.4.3"},
		"Set-Cookie":     {"PHPSESSID=abc123; path=/"},
		"Content-Type":   {"text/html; charset=UTF-8"},
	}

	body := []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta name="generator" content="WordPress 5.8">
			<script src="/wp-includes/js/jquery/jquery.min.js?ver=3.6.0"></script>
		</head>
		<body>
			<div class="wp-content">
				<script>
					var wp_version = "5.8";
				</script>
			</div>
		</body>
		</html>
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wappalyzer.Fingerprint(headers, body)
	}
}

func BenchmarkFingerprintWithInfo(b *testing.B) {
	wappalyzer, err := New()
	if err != nil {
		b.Fatalf("Failed to create wappalyzer: %v", err)
	}

	headers := map[string][]string{
		"Server":       {"nginx/1.18.0"},
		"X-Powered-By": {"Express"},
	}

	body := []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<script src="https://cdn.jsdelivr.net/npm/react@17.0.2/umd/react.production.min.js"></script>
			<script src="https://unpkg.com/vue@3.2.31/dist/vue.global.js"></script>
		</head>
		<body>
			<div id="app"></div>
		</body>
		</html>
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wappalyzer.FingerprintWithInfo(headers, body)
	}
}

func BenchmarkPatternMatching(b *testing.B) {
	pattern, err := ParsePattern("Apache/([\\d.]+)\\;version:\\1")
	if err != nil {
		b.Fatalf("Failed to parse pattern: %v", err)
	}

	target := "Apache/2.4.29 (Ubuntu)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pattern.Evaluate(target)
	}
}

func BenchmarkHeaderNormalization(b *testing.B) {
	wappalyzer, err := New()
	if err != nil {
		b.Fatalf("Failed to create wappalyzer: %v", err)
	}

	headers := map[string][]string{
		"Server":           {"Apache/2.4.29 (Ubuntu)"},
		"X-Powered-By":     {"PHP/7.4.3", "Express"},
		"Content-Type":     {"text/html; charset=UTF-8"},
		"Cache-Control":    {"no-cache", "no-store", "must-revalidate"},
		"Set-Cookie":       {"session=abc123; path=/", "csrf=xyz789; secure"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wappalyzer.normalizeHeaders(headers)
	}
}

// Benchmark real-world scenario
func BenchmarkRealWorldScenario(b *testing.B) {
	wappalyzer, err := New()
	if err != nil {
		b.Fatalf("Failed to create wappalyzer: %v", err)
	}

	// Simulate a typical WordPress site response
	headers := map[string][]string{
		"Server":         {"nginx/1.18.0"},
		"X-Powered-By":   {"PHP/7.4.3"},
		"Content-Type":   {"text/html; charset=UTF-8"},
		"Set-Cookie":     {"wordpress_test_cookie=test; path=/"},
		"Link":           {"<https://example.com/wp-json/>; rel=\"https://api.w.org/\""},
	}

	body := []byte(`
		<!DOCTYPE html>
		<html lang="en-US">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<meta name="generator" content="WordPress 5.8.1">
			<title>Example Site</title>
			<link rel="stylesheet" href="/wp-content/themes/twentytwentyone/style.css?ver=1.4">
			<script src="/wp-includes/js/jquery/jquery.min.js?ver=3.6.0"></script>
			<script src="/wp-content/plugins/contact-form-7/includes/js/index.js?ver=5.4.2"></script>
		</head>
		<body class="home blog">
			<div id="page" class="site">
				<header id="masthead" class="site-header">
					<div class="site-branding">
						<h1 class="site-title">Example Site</h1>
					</div>
				</header>
				<main id="main" class="site-main">
					<article class="post">
						<div class="entry-content">
							<p>Welcome to WordPress!</p>
						</div>
					</article>
				</main>
				<footer id="colophon" class="site-footer">
					<div class="site-info">
						<span>Powered by WordPress</span>
					</div>
				</footer>
			</div>
			<script>
				var wp_version = "5.8.1";
				jQuery(document).ready(function($) {
					// WordPress specific JavaScript
				});
			</script>
		</body>
		</html>
	`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wappalyzer.Fingerprint(headers, body)
	}
}