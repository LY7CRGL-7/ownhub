package server

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	v1 "ownhub/api/ownhub/v1"
	"ownhub/internal/biz"
	"ownhub/internal/conf"
	"ownhub/internal/service"
)

func NewHTTPServer(c *conf.Server, shortSvc *service.ShortService, toolSvc *service.ToolService, portalSvc *service.PortalService, app *conf.App, logger log.Logger) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(recovery.Recovery()),
		khttp.Logger(logger),
	}
	if c != nil && c.Http != nil {
		opts = append(opts, khttp.Network(c.Http.Network))
		opts = append(opts, khttp.Address(c.Http.Addr))
		if c.Http.Timeout != nil {
			opts = append(opts, khttp.Timeout(c.Http.Timeout.AsDuration()))
		}
	}
	srv := khttp.NewServer(opts...)

	// 注册标准服务
	v1.RegisterShortServiceHTTPServer(srv, shortSvc)
	v1.RegisterToolServiceHTTPServer(srv, toolSvc)
	registerPortalRoutes(srv, portalSvc, app)

	// 重定向路由
	srv.HandleFunc("/s/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if code == "" {
			code = strings.TrimPrefix(r.URL.Path, "/s/")
		}

		reply, err := shortSvc.Redirect(r.Context(), &v1.RedirectReq{Code: code})
		if err != nil {
			http.Error(w, "Short link not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, reply.Url, http.StatusFound)
	})

	// 根路由 - 服务前端页面
	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// 读取 web 目录下的 index.html
		content, err := os.ReadFile("D:/golang/code/ownhub/web/index.html")
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<h1>OwnHub</h1><p>API works! But index.html not found.</p>`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	return srv
}

func registerPortalRoutes(srv *khttp.Server, portalSvc *service.PortalService, app *conf.App) {
	githubSecret := ""
	uploadDir := "./storage/uploads"
	if app != nil {
		githubSecret = app.GithubWebhookSecret
		if strings.TrimSpace(app.UploadDir) != "" {
			uploadDir = app.UploadDir
		}
	}

	srv.HandlePrefix("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	srv.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Username    string `json:"username"`
			Password    string `json:"password"`
			DisplayName string `json:"display_name"`
		}
		if err := service.DecodeJSON(r, &req); err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		user, err := portalSvc.Register(r.Context(), req.Username, req.Password, req.DisplayName)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusCreated, user)
	})

	srv.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := service.DecodeJSON(r, &req); err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		token, user, err := portalSvc.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusOK, map[string]interface{}{"token": token, "user": user})
	})

	srv.HandleFunc("/api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		user, err := currentUser(portalSvc, r)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusOK, user)
	})

	srv.HandleFunc("/api/v1/toolbox/ip", func(w http.ResponseWriter, r *http.Request) {
		info, err := portalSvc.LookupIP(r.Context(), r.URL.Query().Get("ip"))
		if err != nil {
			service.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusOK, info)
	})

	srv.HandleFunc("/api/v1/uploads/image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		user, err := currentUser(portalSvc, r)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		defer file.Close()
		asset, err := portalSvc.SaveUploadedFile(r.Context(), user.ID, file, header)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusCreated, asset)
	})

	srv.HandleFunc("/api/v1/notes", func(w http.ResponseWriter, r *http.Request) {
		user, err := currentUser(portalSvc, r)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		switch r.Method {
		case http.MethodPost:
			var req struct {
				Kind    string `json:"kind"`
				Title   string `json:"title"`
				Content string `json:"content"`
				URL     string `json:"url"`
				Tags    string `json:"tags"`
			}
			if err := service.DecodeJSON(r, &req); err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			note := &biz.NoteItem{UserID: user.ID, Kind: req.Kind, Title: req.Title, Content: req.Content, URL: req.URL, Tags: req.Tags}
			if err := portalSvc.CreateNote(r.Context(), note); err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			service.WriteJSON(w, http.StatusCreated, note)
		case http.MethodGet:
			items, err := portalSvc.ListNotes(r.Context(), user.ID, r.URL.Query().Get("kind"))
			if err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			service.WriteJSON(w, http.StatusOK, items)
		default:
			http.NotFound(w, r)
		}
	})

	srv.HandleFunc("/api/v1/reminders", func(w http.ResponseWriter, r *http.Request) {
		user, err := currentUser(portalSvc, r)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		switch r.Method {
		case http.MethodPost:
			var req struct {
				Title   string `json:"title"`
				Content string `json:"content"`
				DueAt   string `json:"due_at"`
			}
			if err := service.DecodeJSON(r, &req); err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			dueAt, err := service.ParseTime(req.DueAt)
			if err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			item := &biz.Reminder{UserID: user.ID, Title: req.Title, Content: req.Content, DueAt: dueAt}
			if err := portalSvc.CreateReminder(r.Context(), item); err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			service.WriteJSON(w, http.StatusCreated, item)
		case http.MethodGet:
			items, err := portalSvc.ListReminders(r.Context(), user.ID, r.URL.Query().Get("status"))
			if err != nil {
				service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			service.WriteJSON(w, http.StatusOK, items)
		default:
			http.NotFound(w, r)
		}
	})

	srv.HandleFunc("/api/v1/reminders/done", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		user, err := currentUser(portalSvc, r)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid reminder id"})
			return
		}
		if err := portalSvc.MarkReminderDone(r.Context(), user.ID, id); err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "status": "done"})
	})

	srv.HandleFunc("/api/v1/webhooks/github", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			service.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		event, err := portalSvc.SaveGitHubWebhook(
			r.Context(),
			githubSecret,
			r.Header.Get("X-Hub-Signature-256"),
			r.Header.Get("X-GitHub-Event"),
			r.Header.Get("X-GitHub-Delivery"),
			payload,
		)
		if err != nil {
			service.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		service.WriteJSON(w, http.StatusAccepted, event)
	})

	_ = os.MkdirAll(filepath.Clean(uploadDir), 0o755)
}

func currentUser(portalSvc *service.PortalService, r *http.Request) (*biz.User, error) {
	authHeader := r.Header.Get("Authorization")
	if strings.TrimSpace(authHeader) == "" {
		return nil, errors.New("missing authorization header")
	}
	return portalSvc.CurrentUser(authHeader)
}
