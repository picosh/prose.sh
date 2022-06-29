package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	"github.com/picosh/cms"
	"github.com/picosh/cms/db/postgres"
	"github.com/picosh/prose.sh/internal"
	"github.com/picosh/proxy"
	"github.com/picosh/send"
	"github.com/picosh/send/scp"
)

type SSHServer struct{}

func (me *SSHServer) authHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func router(sh ssh.Handler, s ssh.Session) []wish.Middleware {
	cmd := s.Command()
	cfg := internal.NewConfigSite()
	mdw := []wish.Middleware{}

	if len(cmd) == 0 {
		mdw = append(mdw,
			bm.Middleware(cms.Middleware(&cfg.ConfigCms, cfg)),
			lm.Middleware(),
		)
	}

	if cmd[0] == "scp" {
		dbh := postgres.NewDB(&cfg.ConfigCms)
		handler := internal.NewDbHandler(dbh, cfg)
		defer dbh.Close()
		mdw = append(mdw, scp.Middleware(handler))
	}

	return mdw
}

func proxyMiddleware(server *ssh.Server) error {
	cfg := internal.NewConfigSite()
	dbh := postgres.NewDB(&cfg.ConfigCms)
	handler := internal.NewDbHandler(dbh, cfg)

	err := send.Middleware(handler)(server)
	if err != nil {
		return err
	}

	return proxy.WithProxy(router)(server)
}

func main() {
	cfg := internal.NewConfigSite()
	logger := cfg.Logger
	host := internal.GetEnv("PROSE_HOST", "0.0.0.0")
	port := internal.GetEnv("PROSE_SSH_PORT", "2222")

	sshServer := &SSHServer{}
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", host, port)),
		wish.WithHostKeyPath("ssh_data/term_info_ed25519"),
		wish.WithPublicKeyAuth(sshServer.authHandler),
		proxyMiddleware,
	)
	if err != nil {
		logger.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Infof("Starting SSH server on %s:%s", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	<-done
	logger.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}
}
