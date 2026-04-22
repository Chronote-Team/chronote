package http

import "github.com/gin-gonic/gin"

type Server struct {
	address string
	router  *gin.Engine
}

func NewServer(address string, router *gin.Engine) *Server {
	return &Server{address: address, router: router}
}

func (s *Server) Run() error {
	return s.router.Run(s.address)
}
