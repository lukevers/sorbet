package main

import (
	"github.com/lukevers/mcgorcon"
	"time"
)

type Server struct {
	// Id is a uint64 that is the server's identification number.
	Id uint64

	// Host is a string that contains the link to the server.
	Host string

	// Port is an int which is the port that the query port is
	// set to.
	Port int

	// Password is a string which is the password to query the
	// minecraft server.
	Password string

	// CreatedAt is a timestamp of when the specific
	// user was created at.
	CreatedAt time.Time

	// UpdatedAt is a timestamp of when the specific
	// user was last updated at.
	UpdatedAt time.Time

	// Rcon is an unexported field that connects with a server.
	rcon mcgorcon.Client `sql:"-"`
}

// Initialize Rcon for an initalized server.
func (s *Server) initalizeRcon() {
	s.rcon, err = mcgorcon.Dial(s.Host, s.Port, s.Password)
	if err != nil {
		s.rcon = mcgorcon.Client{}
	}
}

func (s *Server) Cmd(command string) string {
	tmp := mcgorcon.Client{}
	if s.rcon != tmp {
		response, err := s.rcon.SendCommand(command)
		if err != nil {
			return ""
		}

		return response
	}

	return ""
}
