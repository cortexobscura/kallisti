package mumbled

import (
	"errors"

	"kallisti.chat/server/internal/homes"
)

// HomeServers: which other servers' accounts log in here (see package
// homes; admins only).
func (s *Server) HomeServers(actor uint32) (homes.Rules, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u := s.state.Users[actor]; u == nil || !s.isAdmin(u) {
		return homes.Rules{}, errors.New("only admins see this setting")
	}
	return s.homes, nil
}

// SetHomeServers checks, saves and returns new rules (admins only).
func (s *Server) SetHomeServers(actor uint32, r homes.Rules) (homes.Rules, error) {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
	s.mu.Lock()
	u := s.state.Users[actor]
	admin := u != nil && s.isAdmin(u)
	s.mu.Unlock()
	if !admin {
		return homes.Rules{}, errors.New("only admins change this setting")
	}
	r, err := homes.Normalize(r)
	if err != nil {
		return homes.Rules{}, err
	}
	if s.cfg.DB != nil {
		if err := s.cfg.DB.setSetting("home_servers", r); err != nil {
			return homes.Rules{}, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.homes = r
	return r, nil
}

// AllowsHome: an account from this home server (its host) may log in.
func (s *Server) AllowsHome(host string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.homes.Allows(host)
}
