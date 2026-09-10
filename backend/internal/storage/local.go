package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct {
	rootDir string
}

func NewLocalStore(rootDir string) (*LocalStore, error) {
	for _, sub := range []string{"avatars", "clubs", "diary", "social", "events", "site", "donations"} {
		if err := os.MkdirAll(filepath.Join(rootDir, sub), 0o755); err != nil {
			return nil, fmt.Errorf("create upload dir: %w", err)
		}
	}
	return &LocalStore{rootDir: rootDir}, nil
}

func (s *LocalStore) SaveSocialMedia(postID string, mediaID string, ext string, reader io.Reader) (string, error) {
	relativePath := filepath.ToSlash(filepath.Join("social", postID+"_"+mediaID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create social media file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write social media file: %w", err)
	}
	return relativePath, nil
}

func (s *LocalStore) SaveAvatar(userID string, ext string, reader io.Reader) (string, error) {
	if err := s.RemoveAvatar(userID); err != nil {
		return "", err
	}

	relativePath := filepath.ToSlash(filepath.Join("avatars", userID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create avatar file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write avatar file: %w", err)
	}

	return relativePath, nil
}

func (s *LocalStore) SaveClubLogo(clubID string, ext string, reader io.Reader) (string, error) {
	if err := s.RemoveClubLogo(clubID); err != nil {
		return "", err
	}

	relativePath := filepath.ToSlash(filepath.Join("clubs", clubID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create club logo file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write club logo file: %w", err)
	}

	return relativePath, nil
}

func (s *LocalStore) RemoveClubLogo(clubID string) error {
	clubDir := filepath.Join(s.rootDir, "clubs")
	entries, err := os.ReadDir(clubDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read clubs dir: %w", err)
	}

	prefix := clubID + "."
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), prefix) {
			if err := os.Remove(filepath.Join(clubDir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove club logo file: %w", err)
			}
		}
	}
	return nil
}

func (s *LocalStore) SaveDiaryPhoto(entryID string, ext string, reader io.Reader) (string, error) {
	if err := s.RemoveDiaryPhoto(entryID); err != nil {
		return "", err
	}

	relativePath := filepath.ToSlash(filepath.Join("diary", entryID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create diary photo file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write diary photo file: %w", err)
	}

	return relativePath, nil
}

func (s *LocalStore) RemoveDiaryPhoto(entryID string) error {
	dir := filepath.Join(s.rootDir, "diary")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read diary dir: %w", err)
	}

	prefix := entryID + "."
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), prefix) {
			if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove diary photo file: %w", err)
			}
		}
	}
	return nil
}

func (s *LocalStore) RemoveAvatar(userID string) error {
	avatarDir := filepath.Join(s.rootDir, "avatars")
	entries, err := os.ReadDir(avatarDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read avatar dir: %w", err)
	}

	prefix := userID + "."
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), prefix) {
			if err := os.Remove(filepath.Join(avatarDir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove avatar file: %w", err)
			}
		}
	}
	return nil
}

func (s *LocalStore) RemoveByRelativePath(relativePath string) error {
	if relativePath == "" {
		return nil
	}
	clean := filepath.Clean(relativePath)
	if strings.Contains(clean, "..") {
		return fmt.Errorf("invalid path")
	}
	fullPath := filepath.Join(s.rootDir, clean)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}

func (s *LocalStore) ensureParentDir(fullPath string) error {
	return os.MkdirAll(filepath.Dir(fullPath), 0o755)
}

func (s *LocalStore) SaveDonationReceipt(donationID, ext string, reader io.Reader) (string, error) {
	relativePath := filepath.ToSlash(filepath.Join("donations", donationID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)
	if err := s.ensureParentDir(fullPath); err != nil {
		return "", fmt.Errorf("create donation dir: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create donation receipt: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write donation receipt: %w", err)
	}
	return relativePath, nil
}

func (s *LocalStore) SaveSiteImage(name, ext string, reader io.Reader) (string, error) {
	relativePath := filepath.ToSlash(filepath.Join("site", name+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)
	if err := s.ensureParentDir(fullPath); err != nil {
		return "", fmt.Errorf("create site dir: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create site image: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write site image: %w", err)
	}
	return relativePath, nil
}

func (s *LocalStore) SaveEventImage(eventID, imageID, ext string, reader io.Reader) (string, error) {
	relativePath := filepath.ToSlash(filepath.Join("events", eventID+"_"+imageID+ext))
	fullPath := filepath.Join(s.rootDir, relativePath)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create event image: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("write event image: %w", err)
	}
	return relativePath, nil
}

func (s *LocalStore) RootDir() string {
	return s.rootDir
}
