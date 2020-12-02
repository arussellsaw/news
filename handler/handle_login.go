package handler

import (
	"fmt"
	"github.com/arussellsaw/news/dao"
	"github.com/arussellsaw/news/domain"
	"github.com/arussellsaw/news/idgen"
	"github.com/dgrijalva/jwt-go"
	"github.com/monzo/slog"
	"html/template"
	"net/http"
	"os"
)

type loginPage struct {
	base
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		loginRedirect(w, r)
		return
	}

	t := template.New("frame.html")
	t, err := t.ParseFiles("tmpl/frame.html", "tmpl/login.html")
	if err != nil {
		slog.Error(ctx, "Error parsing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	l := loginPage{
		base: base{
			ID:   "Login",
			User: domain.UserFromContext(ctx),
		},
	}
	err = t.Execute(w, &l)
	if err != nil {
		slog.Error(ctx, "Error executing template: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

func loginRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()
	username := r.Form.Get("username")
	pw := r.Form.Get("pw")
	signup := r.Form.Get("signup")

	if signup != "" {
		u, err := dao.GetUserByName(ctx, username)
		if err != nil {
			slog.Error(ctx, "Err: %s", err)
		}
		if u != nil {
			slog.Info(ctx, "User already exists: %s", username)
			goto login
		}
		u = domain.NewUser(ctx, username, pw)
		err = dao.SetUser(ctx, u)
		if err != nil {
			slog.Error(ctx, "Error creating user: %s", err)
			http.Error(w, "couldn't create user", 404)
			return
		}
		for _, src := range domain.GetSources() {
			src := src
			src.ID = idgen.New("src")
			src.OwnerID = u.ID
			err := dao.SetSource(ctx, &src)
			if err != nil {
				slog.Error(ctx, "Error creating user: %s", err)
				http.Error(w, "couldn't create user", 404)
				return
			}
		}
	}
login:
	u, err := dao.GetUserByName(ctx, username)
	if err != nil {
		slog.Error(ctx, "Error getting user: %s", err)
		http.Error(w, "couldn't find user", 404)
		return
	}

	if !u.ValidatePassword(pw) {
		http.Error(w, "couldn't find user", 404)
		return
	}

	sess, err := u.Session()
	if err != nil {
		slog.Error(ctx, "Error creating session: %s", err)
		http.Error(w, "error creating session", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "sess",
		Value: sess,
	})
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sessionCookie, err := r.Cookie("sess")
		if err != nil {
			slog.Info(ctx, "no cookie: %s", err)
			next.ServeHTTP(w, r)
			return
		}

		if sessionCookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(sessionCookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("TOKEN_SECRET")), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID, ok := claims["user"].(string)
			if !ok {
				slog.Info(ctx, "no user claim")
				next.ServeHTTP(w, r)
				return
			}
			u, err := dao.GetUser(ctx, userID)
			if err != nil {
				slog.Info(ctx, "no user: %s %s", userID, err)
				next.ServeHTTP(w, r)
				return
			}
			slog.Info(ctx, "User session: %s %s", u.ID, u.Name)
			ctx := domain.WithUser(ctx, u)
			r = r.WithContext(ctx)
		} else {
			slog.Error(ctx, "invalid token %s", token)
		}
		next.ServeHTTP(w, r)
	})
}
