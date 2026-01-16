package validation

import (
	"strings"
	"testing"
)

func TestNewCustomValidator(t *testing.T) {
	cv := NewCustomValidator()
	if cv == nil {
		t.Fatal("NewCustomValidator returned nil")
	}
	if cv.validate == nil {
		t.Error("validate field should be initialized")
	}
}

func TestValidateStrongPassword(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		// Valid passwords (at least 3 character types)
		{"valid with upper lower number", "Password123", true},
		{"valid with upper lower special", "Password!@#", true},
		{"valid with lower number special", "password123!", true},
		{"valid with all types", "Password123!", true},
		{"valid minimum length", "Pass123!", true},

		// Invalid passwords
		{"too short", "Pass1!", false},
		{"only lowercase", "password", false},
		{"only uppercase", "PASSWORD", false},
		{"only numbers", "12345678", false},
		{"only two types lowercase number", "password1", false},
		{"empty password", "", false},
		{"too long", strings.Repeat("a", 129), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Password string `validate:"strong_password"`
			}

			err := cv.ValidateStruct(&TestStruct{Password: tt.password})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("password %q: got valid=%v, want valid=%v, err=%v",
					tt.password, isValid, tt.valid, err)
			}
		})
	}
}

func TestValidateSafeUsername(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name     string
		username string
		valid    bool
	}{
		// Valid usernames
		{"simple alphanumeric", "john123", true},
		{"with underscore", "john_doe", true},
		{"with hyphen", "john-doe", true},
		{"mixed case", "JohnDoe", true},
		{"minimum length", "abc", true},

		// Invalid usernames
		{"too short", "ab", false},
		{"too long", strings.Repeat("a", 51), false},
		{"starts with underscore", "_john", false},
		{"ends with underscore", "john_", false},
		{"starts with hyphen", "-john", false},
		{"ends with hyphen", "john-", false},
		{"contains space", "john doe", false},
		{"contains special char", "john@doe", false},
		{"forbidden admin", "admin", false},
		{"forbidden root", "root", false},
		{"forbidden system", "system", false},
		{"forbidden case insensitive", "ADMIN", false},
		{"empty username", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Username string `validate:"safe_username"`
			}

			err := cv.ValidateStruct(&TestStruct{Username: tt.username})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("username %q: got valid=%v, want valid=%v, err=%v",
					tt.username, isValid, tt.valid, err)
			}
		})
	}
}

func TestValidateSafeString(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Valid strings
		{"plain text", "Hello World", true},
		{"with numbers", "Test 123", true},
		{"with punctuation", "Hello, World!", true},
		{"empty string", "", true},

		// Invalid strings (XSS attempts)
		{"script tag", "<script>alert('xss')</script>", false},
		{"javascript protocol", "javascript:alert(1)", false},
		{"onclick handler", "<div onclick=alert(1)>", false},
		{"onload handler", "<img onload=alert(1)>", false},
		{"eval function", "eval('code')", false},
		{"iframe tag", "<iframe src='evil.com'>", false},
		{"vbscript protocol", "vbscript:msgbox(1)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Content string `validate:"safe_string"`
			}

			err := cv.ValidateStruct(&TestStruct{Content: tt.input})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("input %q: got valid=%v, want valid=%v",
					tt.input, isValid, tt.valid)
			}
		})
	}
}

func TestValidateURLSlug(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name  string
		slug  string
		valid bool
	}{
		// Valid slugs
		{"simple slug", "hello-world", true},
		{"with numbers", "post-123", true},
		{"only letters", "helloworld", true},
		{"only numbers", "12345", true},
		{"single char", "a", true},

		// Invalid slugs
		{"empty slug", "", false},
		{"too long", strings.Repeat("a", 101), false},
		{"uppercase letters", "Hello-World", false},
		{"starts with hyphen", "-hello", false},
		{"ends with hyphen", "hello-", false},
		{"consecutive hyphens", "hello--world", false},
		{"contains underscore", "hello_world", false},
		{"contains space", "hello world", false},
		{"contains special char", "hello@world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Slug string `validate:"url_slug"`
			}

			err := cv.ValidateStruct(&TestStruct{Slug: tt.slug})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("slug %q: got valid=%v, want valid=%v",
					tt.slug, isValid, tt.valid)
			}
		})
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name  string
		phone string
		valid bool
	}{
		// Valid phone numbers
		{"empty optional", "", true},
		{"chinese mobile", "13800138000", true},
		{"us format", "555-123-4567", true},
		{"international", "+86-13800138000", true},
		{"pure digits", "1234567890", true},

		// Invalid phone numbers
		{"too short", "12345", false},
		{"contains letters", "123abc4567", false},
		{"invalid format", "123.456.7890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Phone string `validate:"phone_number"`
			}

			err := cv.ValidateStruct(&TestStruct{Phone: tt.phone})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("phone %q: got valid=%v, want valid=%v",
					tt.phone, isValid, tt.valid)
			}
		})
	}
}

func TestValidateSafeHTML(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name  string
		html  string
		valid bool
	}{
		// Valid HTML
		{"plain text", "Hello World", true},
		{"safe tags", "<p>Hello</p><b>World</b>", true},
		{"with links", "<a href='url'>Link</a>", true},
		{"empty string", "", true},

		// Invalid HTML (dangerous tags)
		{"script tag", "<script>alert(1)</script>", false},
		{"iframe tag", "<iframe src='evil.com'></iframe>", false},
		{"form tag", "<form action='evil.com'>", false},
		{"input tag", "<input type='text'>", false},
		{"onclick event", "<div onclick='alert(1)'>", false},
		{"onload event", "<body onload='alert(1)'>", false},
		{"style tag", "<style>body{display:none}</style>", false},
		{"object tag", "<object data='evil.swf'>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				HTML string `validate:"safe_html"`
			}

			err := cv.ValidateStruct(&TestStruct{HTML: tt.html})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("html %q: got valid=%v, want valid=%v",
					tt.html, isValid, tt.valid)
			}
		})
	}
}

func TestValidateSafeFilename(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name     string
		filename string
		valid    bool
	}{
		// Valid filenames
		{"simple name", "document.txt", true},
		{"with numbers", "file123.pdf", true},
		{"with hyphen", "my-file.doc", true},
		{"with underscore", "my_file.doc", true},
		{"no extension", "readme", true},

		// Invalid filenames
		{"empty filename", "", false},
		{"too long", strings.Repeat("a", 256), false},
		{"contains slash", "path/file.txt", false},
		{"contains backslash", "path\\file.txt", false},
		{"contains colon", "file:name.txt", false},
		{"contains asterisk", "file*.txt", false},
		{"contains question", "file?.txt", false},
		{"contains quote", "file\".txt", false},
		{"contains less than", "file<.txt", false},
		{"contains greater than", "file>.txt", false},
		{"contains pipe", "file|.txt", false},
		{"dot only", ".", false},
		{"double dot", "..", false},
		{"reserved CON", "CON", false},
		{"reserved PRN", "PRN", false},
		{"reserved NUL", "NUL", false},
		{"starts with dot", ".hidden", false},
		{"ends with dot", "file.", false},
		{"starts with space", " file.txt", false},
		{"ends with space", "file.txt ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Filename string `validate:"safe_filename"`
			}

			err := cv.ValidateStruct(&TestStruct{Filename: tt.filename})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("filename %q: got valid=%v, want valid=%v",
					tt.filename, isValid, tt.valid)
			}
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	cv := NewCustomValidator()

	tests := []struct {
		name  string
		ip    string
		valid bool
	}{
		// Valid IPv4
		{"localhost", "127.0.0.1", true},
		{"private ip", "192.168.1.1", true},
		{"public ip", "8.8.8.8", true},
		{"zero ip", "0.0.0.0", true},
		{"max ip", "255.255.255.255", true},

		// Valid IPv6
		{"ipv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},

		// Invalid IPs
		{"empty", "", false},
		{"invalid format", "192.168.1", false},
		{"out of range", "256.1.1.1", false},
		{"leading zeros", "192.168.01.1", false},
		{"negative", "-1.0.0.0", false},
		{"letters", "abc.def.ghi.jkl", false},
		{"too many octets", "1.2.3.4.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				IP string `validate:"ip_address"`
			}

			err := cv.ValidateStruct(&TestStruct{IP: tt.ip})
			isValid := err == nil

			if isValid != tt.valid {
				t.Errorf("ip %q: got valid=%v, want valid=%v",
					tt.ip, isValid, tt.valid)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		// Valid emails
		{"simple email", "user@example.com", false},
		{"with subdomain", "user@mail.example.com", false},
		{"with plus", "user+tag@example.com", false},
		{"with dots", "first.last@example.com", false},

		// Invalid emails
		{"empty", "", true},
		{"too short", "a@b", true},
		{"too long", strings.Repeat("a", 250) + "@example.com", true},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"contains script", "user<script>@example.com", true},
		{"blocked domain tempmail", "user@tempmail.org", true},
		{"blocked domain mailinator", "user@mailinator.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			hasErr := err != nil

			if hasErr != tt.wantErr {
				t.Errorf("ValidateEmail(%q): got error=%v, want error=%v",
					tt.email, hasErr, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateStrongPassword(b *testing.B) {
	cv := NewCustomValidator()
	type TestStruct struct {
		Password string `validate:"strong_password"`
	}
	ts := &TestStruct{Password: "Password123!"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cv.ValidateStruct(ts)
	}
}

func BenchmarkValidateSafeUsername(b *testing.B) {
	cv := NewCustomValidator()
	type TestStruct struct {
		Username string `validate:"safe_username"`
	}
	ts := &TestStruct{Username: "john_doe123"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cv.ValidateStruct(ts)
	}
}

func BenchmarkValidateSafeString(b *testing.B) {
	cv := NewCustomValidator()
	type TestStruct struct {
		Content string `validate:"safe_string"`
	}
	ts := &TestStruct{Content: "This is a safe string with no XSS attempts."}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cv.ValidateStruct(ts)
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	email := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateEmail(email)
	}
}
