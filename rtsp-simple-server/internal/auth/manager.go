// Package auth contains the authentication system.
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/bluenviron/gortsplib/v4/pkg/auth"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/headers"
	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// PauseAfterError is the pause to apply after an authentication failure.
	PauseAfterError = 2 * time.Second

	rtspAuthRealm    = "IPCAM"
	jwtRefreshPeriod = 60 * 60 * time.Second
)

// Protocol is a protocol.
type Protocol string

// protocols.
const (
	ProtocolRTSP   Protocol = "rtsp"
	ProtocolRTMP   Protocol = "rtmp"
	ProtocolHLS    Protocol = "hls"
	ProtocolWebRTC Protocol = "webrtc"
	ProtocolSRT    Protocol = "srt"
)

// Request is an authentication request.
type Request struct {
	User   string
	Pass   string
	IP     net.IP
	Action conf.AuthAction

	// only for ActionPublish, ActionRead, ActionPlayback
	Path        string
	Protocol    Protocol
	ID          *uuid.UUID
	Query       string
	RTSPRequest *base.Request
	RTSPNonce   string
}

// Error is a authentication error.
type Error struct {
	Message string
}

// Error implements the error interface.
func (e Error) Error() string {
	return "authentication failed: " + e.Message
}

func matchesPermission(perms []conf.AuthInternalUserPermission, req *Request) bool {
	for _, perm := range perms {
		if perm.Action == req.Action {
			if perm.Action == conf.AuthActionPublish ||
				perm.Action == conf.AuthActionRead ||
				perm.Action == conf.AuthActionPlayback {
				switch {
				case perm.Path == "":
					return true

				case strings.HasPrefix(perm.Path, "~"):
					regexp, err := regexp.Compile(perm.Path[1:])
					if err == nil && regexp.MatchString(req.Path) {
						return true
					}

				case perm.Path == req.Path:
					return true
				}
			} else {
				return true
			}
		}
	}

	return false
}

type customClaims struct {
	jwt.RegisteredClaims
	MediaMTXPermissions []conf.AuthInternalUserPermission `json:"mediamtx_permissions"`
}

// Manager is the authentication manager.
type Manager struct {
	Method          conf.AuthMethod
	InternalUsers   []conf.AuthInternalUser
	HTTPAddress     string
	HTTPExclude     []conf.AuthInternalUserPermission
	JWTJWKS         string
	ReadTimeout     time.Duration
	RTSPAuthMethods []auth.ValidateMethod

	mutex          sync.RWMutex
	jwtHTTPClient  *http.Client
	jwtLastRefresh time.Time
	jwtKeyFunc     keyfunc.Keyfunc
}

// ReloadInternalUsers reloads InternalUsers.
func (m *Manager) ReloadInternalUsers(u []conf.AuthInternalUser) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.InternalUsers = u
}

// Authenticate authenticates a request.
func (m *Manager) Authenticate(req *Request) error {
	err := m.authenticateInner(req)
	if err != nil {
		return Error{Message: err.Error()}
	}
	return nil
}

func (m *Manager) authenticateInner(req *Request) error {
	// if this is a RTSP request, fill username and password
	var rtspAuthHeader headers.Authorization

	if req.RTSPRequest != nil {
		err := rtspAuthHeader.Unmarshal(req.RTSPRequest.Header["Authorization"])
		if err == nil {
			if rtspAuthHeader.Method == headers.AuthMethodBasic {
				req.User = rtspAuthHeader.BasicUser
				req.Pass = rtspAuthHeader.BasicPass
			} else { // digest
				req.User = rtspAuthHeader.Username
			}
		}
	}

	switch m.Method {
	case conf.AuthMethodInternal:
		return m.authenticateInternal(req, &rtspAuthHeader)

	case conf.AuthMethodHTTP:
		return m.authenticateHTTP(req)

	default:
		return m.authenticateJWT(req)
	}
}

func (m *Manager) authenticateInternal(req *Request, rtspAuthHeader *headers.Authorization) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, u := range m.InternalUsers {
		if err := m.authenticateWithUser(req, rtspAuthHeader, &u); err == nil {
			return nil
		}
	}

	return fmt.Errorf("authentication failed")
}

func (m *Manager) authenticateWithUser(
	req *Request,
	rtspAuthHeader *headers.Authorization,
	u *conf.AuthInternalUser,
) error {
	if u.User != "any" && !u.User.Check(req.User) {
		return fmt.Errorf("wrong user")
	}

	if len(u.IPs) != 0 && !u.IPs.Contains(req.IP) {
		return fmt.Errorf("IP not allowed")
	}

	if !matchesPermission(u.Permissions, req) {
		return fmt.Errorf("user doesn't have permission to perform action")
	}

	if u.User != "any" {
		if req.RTSPRequest != nil && rtspAuthHeader.Method == headers.AuthMethodDigest {
			err := auth.Validate(
				req.RTSPRequest,
				string(u.User),
				string(u.Pass),
				m.RTSPAuthMethods,
				rtspAuthRealm,
				req.RTSPNonce)
			if err != nil {
				return err
			}
		} else if !u.Pass.Check(req.Pass) {
			return fmt.Errorf("invalid credentials")
		}
	}

	return nil
}

func (m *Manager) authenticateHTTP(req *Request) error {
	if matchesPermission(m.HTTPExclude, req) {
		return nil
	}

	enc, _ := json.Marshal(struct {
		IP       string     `json:"ip"`
		User     string     `json:"user"`
		Password string     `json:"password"`
		Action   string     `json:"action"`
		Path     string     `json:"path"`
		Protocol string     `json:"protocol"`
		ID       *uuid.UUID `json:"id"`
		Query    string     `json:"query"`
	}{
		IP:       req.IP.String(),
		User:     req.User,
		Password: req.Pass,
		Action:   string(req.Action),
		Path:     req.Path,
		Protocol: string(req.Protocol),
		ID:       req.ID,
		Query:    req.Query,
	})

	res, err := http.Post(m.HTTPAddress, "application/json", bytes.NewReader(enc))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		if resBody, err := io.ReadAll(res.Body); err == nil && len(resBody) != 0 {
			return fmt.Errorf("server replied with code %d: %s", res.StatusCode, string(resBody))
		}

		return fmt.Errorf("server replied with code %d", res.StatusCode)
	}

	return nil
}

func (m *Manager) authenticateJWT(req *Request) error {
	keyfunc, err := m.pullJWTJWKS()
	if err != nil {
		return err
	}

	v, err := url.ParseQuery(req.Query)
	if err != nil {
		return err
	}

	if len(v["jwt"]) != 1 {
		return fmt.Errorf("JWT not provided")
	}

	var cc customClaims
	_, err = jwt.ParseWithClaims(v["jwt"][0], &cc, keyfunc)
	if err != nil {
		return err
	}

	if !matchesPermission(cc.MediaMTXPermissions, req) {
		return fmt.Errorf("user doesn't have permission to perform action")
	}

	return nil
}

func (m *Manager) pullJWTJWKS() (jwt.Keyfunc, error) {
	now := time.Now()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if now.Sub(m.jwtLastRefresh) >= jwtRefreshPeriod {
		if m.jwtHTTPClient == nil {
			m.jwtHTTPClient = &http.Client{
				Timeout:   (m.ReadTimeout),
				Transport: &http.Transport{},
			}
		}

		res, err := m.jwtHTTPClient.Get(m.JWTJWKS)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var raw json.RawMessage
		err = json.NewDecoder(res.Body).Decode(&raw)
		if err != nil {
			return nil, err
		}

		tmp, err := keyfunc.NewJWKSetJSON(raw)
		if err != nil {
			return nil, err
		}

		m.jwtKeyFunc = tmp
		m.jwtLastRefresh = now
	}

	return m.jwtKeyFunc.Keyfunc, nil
}
