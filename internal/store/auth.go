package store

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

func (s *Store) HasLock() (bool, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM vault WHERE id = 1`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) CreateLock(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err = s.db.Exec(`INSERT INTO vault(id, password_hash) VALUES(1, ?)`, string(hash)); err != nil {
		return err
	}
	s.unlockWith(password)
	if err := s.EnsureEncrypted(); err != nil {
		return err
	}
	return s.NormalizeFeedLocations()
}

func (s *Store) Unlock(password string) error {
	var hash string
	if err := s.db.QueryRow(`SELECT password_hash FROM vault WHERE id = 1`).Scan(&hash); err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return errors.New("bad vault password")
	}
	s.unlockWith(password)
	if err := s.EnsureEncrypted(); err != nil {
		return err
	}
	return s.NormalizeFeedLocations()
}

func (s *Store) unlockWith(password string) {
	sum := sha3.Sum256([]byte(password))
	s.key = sum[:]
	s.unlocked = true
}
