package security

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// SecureFileMode defines the strongly typed wrapper we use for file permission constants.
type SecureFileMode fs.FileMode

const (
	// Recommended safe permission presets that follow a least-privilege strategy.
	SecureFileMode_ReadOnlyUser   SecureFileMode = 0o400
	SecureFileMode_ReadWriteUser  SecureFileMode = 0o600
	SecureFileMode_ReadOnlyAll    SecureFileMode = 0o444
	SecureFileMode_ReadWriteAll   SecureFileMode = 0o666 // generally discouraged
	SecureFileMode_ExecutableUser SecureFileMode = 0o700
	SecureFileMode_ExecutableAll  SecureFileMode = 0o755

	DefaultFileMode       = SecureFileMode_ReadWriteUser  // 0600
	DefaultDirMode        = SecureFileMode_ExecutableUser // 0700
	DefaultConfigMode     = SecureFileMode_ReadOnlyUser   // 0400
	DefaultLogMode        = SecureFileMode_ReadWriteUser  // 0600
	DefaultTempMode       = SecureFileMode_ReadWriteUser  // 0600
	DefaultExecutableMode = SecureFileMode_ExecutableUser // 0700
)

// SecureFileOptions controls how secure write helpers behave.
type SecureFileOptions struct {
	Mode      SecureFileMode
	CreateDir bool
}

// SecureWriteFile writes a file with strict path validation and hardened defaults.
func SecureWriteFile(filename string, data []byte, opts *SecureFileOptions) error {
	if opts == nil {
		opts = &SecureFileOptions{Mode: DefaultFileMode}
	}

	if err := ValidateSecurePath(filename, &SecurePathOptions{
		AllowAbsolute: true,
		AllowDotDot:   false,
		MaxDepth:      20,
	}); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	if opts.CreateDir {
		dir := filepath.Dir(filename)
		if dir != "." && dir != "" {
			// #nosec G301 -- directory is created with locked-down permissions
			if err := os.MkdirAll(dir, fs.FileMode(DefaultDirMode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	// #nosec G304 G306 -- path is already validated and mode is controlled by the caller
	return os.WriteFile(filename, data, fs.FileMode(opts.Mode))
}

// SecureCreateFile creates a truncated file while enforcing secure defaults.
func SecureCreateFile(filename string, mode SecureFileMode) (*os.File, error) {
	if err := ValidateSecurePath(filename, &SecurePathOptions{
		AllowAbsolute: true,
		AllowDotDot:   false,
		MaxDepth:      20,
	}); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	// #nosec G304 G306 -- path is already validated and explicit mode is supplied
	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(mode))
}

// SecureOpenFile opens a file after running the shared path validation helper.
func SecureOpenFile(filename string, flag int, mode SecureFileMode) (*os.File, error) {
	if err := ValidateSecurePath(filename, &SecurePathOptions{
		AllowAbsolute: true,
		AllowDotDot:   false,
		MaxDepth:      20,
	}); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	// #nosec G304 G306 -- path is already validated and explicit mode is supplied
	return os.OpenFile(filename, flag, fs.FileMode(mode))
}

// SecureMkdirAll mirrors os.MkdirAll but enforces path validation and safe defaults.
func SecureMkdirAll(path string, mode SecureFileMode) error {
	if err := ValidateSecurePath(path, &SecurePathOptions{
		AllowAbsolute: true,
		AllowDotDot:   false,
		MaxDepth:      20,
	}); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	// #nosec G301 -- caller supplies a locked-down mode
	return os.MkdirAll(path, fs.FileMode(mode))
}

// ValidateFilePermissions ensures the file is not more permissive than expected.
func ValidateFilePermissions(filename string, expectedMode SecureFileMode) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	actualMode := info.Mode().Perm()
	expected := fs.FileMode(expectedMode)

	if runtime.GOOS == "windows" {
		// Windows ignores POSIX permission bits; callers must rely on ACLs instead.
		return nil
	}

	if actualMode&^expected != 0 {
		return fmt.Errorf("file permissions are not secure: got %o, expected max %o", actualMode, expected)
	}

	return nil
}

// GetRecommendedMode provides a suggested mode for common file types.
func GetRecommendedMode(fileType string) SecureFileMode {
	switch fileType {
	case "config", "configuration":
		return DefaultConfigMode
	case "log", "logs":
		return DefaultLogMode
	case "temp", "temporary":
		return DefaultTempMode
	case "executable", "binary":
		return DefaultExecutableMode
	case "data", "json", "yaml", "xml":
		return DefaultFileMode
	default:
		return DefaultFileMode
	}
}

// IsSecurePermission returns true only if the mode excludes group/other access bits.
func IsSecurePermission(mode fs.FileMode) bool {
	perm := mode.Perm()

	if perm&0o070 != 0 {
		return false
	}

	if perm&0o007 != 0 {
		return false
	}

	return true
}
