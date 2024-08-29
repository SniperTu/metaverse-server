package conf

import (
	"encoding/json"
	"fmt"
	"net"
	gourl "net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/bluenviron/gortsplib/v4/pkg/base"
)

var rePathName = regexp.MustCompile(`^[0-9a-zA-Z_\-/\.~]+$`)

func isValidPathName(name string) error {
	if name == "" {
		return fmt.Errorf("cannot be empty")
	}

	if name[0] == '/' {
		return fmt.Errorf("can't begin with a slash")
	}

	if name[len(name)-1] == '/' {
		return fmt.Errorf("can't end with a slash")
	}

	if !rePathName.MatchString(name) {
		return fmt.Errorf("can contain only alphanumeric characters, underscore, dot, tilde, minus or slash")
	}

	return nil
}

func srtCheckPassphrase(passphrase string) error {
	switch {
	case len(passphrase) < 10 || len(passphrase) > 79:
		return fmt.Errorf("must be between 10 and 79 characters")

	default:
		return nil
	}
}

// FindPathConf returns the configuration corresponding to the given path name.
func FindPathConf(pathConfs map[string]*Path, name string) (string, *Path, []string, error) {
	err := isValidPathName(name)
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid path name: %w (%s)", err, name)
	}

	// normal path
	if pathConf, ok := pathConfs[name]; ok {
		return name, pathConf, nil, nil
	}

	// regular expression-based path
	for pathConfName, pathConf := range pathConfs {
		if pathConf.Regexp != nil && pathConfName != "all" && pathConfName != "all_others" {
			m := pathConf.Regexp.FindStringSubmatch(name)
			if m != nil {
				return pathConfName, pathConf, m, nil
			}
		}
	}

	// all_others
	for pathConfName, pathConf := range pathConfs {
		if pathConfName == "all" || pathConfName == "all_others" {
			m := pathConf.Regexp.FindStringSubmatch(name)
			if m != nil {
				return pathConfName, pathConf, m, nil
			}
		}
	}

	return "", nil, nil, fmt.Errorf("path '%s' is not configured", name)
}

// Path is a path configuration.
// WARNING: Avoid using slices directly due to https://github.com/golang/go/issues/21092
type Path struct {
	Regexp *regexp.Regexp `json:"-"`    // filled by Check()
	Name   string         `json:"name"` // filled by Check()

	// General
	Source                     string         `json:"source"`
	SourceFingerprint          string         `json:"sourceFingerprint"`
	SourceOnDemand             bool           `json:"sourceOnDemand"`
	SourceOnDemandStartTimeout StringDuration `json:"sourceOnDemandStartTimeout"`
	SourceOnDemandCloseAfter   StringDuration `json:"sourceOnDemandCloseAfter"`
	MaxReaders                 int            `json:"maxReaders"`
	SRTReadPassphrase          string         `json:"srtReadPassphrase"`
	Fallback                   string         `json:"fallback"`

	// Record
	Record                bool           `json:"record"`
	Playback              *bool          `json:"playback,omitempty"` // deprecated
	RecordPath            string         `json:"recordPath"`
	RecordFormat          RecordFormat   `json:"recordFormat"`
	RecordPartDuration    StringDuration `json:"recordPartDuration"`
	RecordSegmentDuration StringDuration `json:"recordSegmentDuration"`
	RecordDeleteAfter     StringDuration `json:"recordDeleteAfter"`

	// Authentication (deprecated)
	PublishUser *Credential `json:"publishUser,omitempty"` // deprecated
	PublishPass *Credential `json:"publishPass,omitempty"` // deprecated
	PublishIPs  *IPNetworks `json:"publishIPs,omitempty"`  // deprecated
	ReadUser    *Credential `json:"readUser,omitempty"`    // deprecated
	ReadPass    *Credential `json:"readPass,omitempty"`    // deprecated
	ReadIPs     *IPNetworks `json:"readIPs,omitempty"`     // deprecated

	// Publisher source
	OverridePublisher        bool   `json:"overridePublisher"`
	DisablePublisherOverride *bool  `json:"disablePublisherOverride,omitempty"` // deprecated
	SRTPublishPassphrase     string `json:"srtPublishPassphrase"`

	// RTSP source
	RTSPTransport       RTSPTransport  `json:"rtspTransport"`
	RTSPAnyPort         bool           `json:"rtspAnyPort"`
	SourceProtocol      *RTSPTransport `json:"sourceProtocol,omitempty"`      // deprecated
	SourceAnyPortEnable *bool          `json:"sourceAnyPortEnable,omitempty"` // deprecated
	RTSPRangeType       RTSPRangeType  `json:"rtspRangeType"`
	RTSPRangeStart      string         `json:"rtspRangeStart"`

	// Redirect source
	SourceRedirect string `json:"sourceRedirect"`

	// Raspberry Pi Camera source
	RPICameraCamID             int       `json:"rpiCameraCamID"`
	RPICameraWidth             int       `json:"rpiCameraWidth"`
	RPICameraHeight            int       `json:"rpiCameraHeight"`
	RPICameraHFlip             bool      `json:"rpiCameraHFlip"`
	RPICameraVFlip             bool      `json:"rpiCameraVFlip"`
	RPICameraBrightness        float64   `json:"rpiCameraBrightness"`
	RPICameraContrast          float64   `json:"rpiCameraContrast"`
	RPICameraSaturation        float64   `json:"rpiCameraSaturation"`
	RPICameraSharpness         float64   `json:"rpiCameraSharpness"`
	RPICameraExposure          string    `json:"rpiCameraExposure"`
	RPICameraAWB               string    `json:"rpiCameraAWB"`
	RPICameraAWBGains          []float64 `json:"rpiCameraAWBGains"`
	RPICameraDenoise           string    `json:"rpiCameraDenoise"`
	RPICameraShutter           int       `json:"rpiCameraShutter"`
	RPICameraMetering          string    `json:"rpiCameraMetering"`
	RPICameraGain              float64   `json:"rpiCameraGain"`
	RPICameraEV                float64   `json:"rpiCameraEV"`
	RPICameraROI               string    `json:"rpiCameraROI"`
	RPICameraHDR               bool      `json:"rpiCameraHDR"`
	RPICameraTuningFile        string    `json:"rpiCameraTuningFile"`
	RPICameraMode              string    `json:"rpiCameraMode"`
	RPICameraFPS               float64   `json:"rpiCameraFPS"`
	RPICameraIDRPeriod         int       `json:"rpiCameraIDRPeriod"`
	RPICameraBitrate           int       `json:"rpiCameraBitrate"`
	RPICameraProfile           string    `json:"rpiCameraProfile"`
	RPICameraLevel             string    `json:"rpiCameraLevel"`
	RPICameraAfMode            string    `json:"rpiCameraAfMode"`
	RPICameraAfRange           string    `json:"rpiCameraAfRange"`
	RPICameraAfSpeed           string    `json:"rpiCameraAfSpeed"`
	RPICameraLensPosition      float64   `json:"rpiCameraLensPosition"`
	RPICameraAfWindow          string    `json:"rpiCameraAfWindow"`
	RPICameraTextOverlayEnable bool      `json:"rpiCameraTextOverlayEnable"`
	RPICameraTextOverlay       string    `json:"rpiCameraTextOverlay"`

	// Hooks
	RunOnInit                  string         `json:"runOnInit"`
	RunOnInitRestart           bool           `json:"runOnInitRestart"`
	RunOnDemand                string         `json:"runOnDemand"`
	RunOnDemandRestart         bool           `json:"runOnDemandRestart"`
	RunOnDemandStartTimeout    StringDuration `json:"runOnDemandStartTimeout"`
	RunOnDemandCloseAfter      StringDuration `json:"runOnDemandCloseAfter"`
	RunOnUnDemand              string         `json:"runOnUnDemand"`
	RunOnReady                 string         `json:"runOnReady"`
	RunOnReadyRestart          bool           `json:"runOnReadyRestart"`
	RunOnNotReady              string         `json:"runOnNotReady"`
	RunOnRead                  string         `json:"runOnRead"`
	RunOnReadRestart           bool           `json:"runOnReadRestart"`
	RunOnUnread                string         `json:"runOnUnread"`
	RunOnRecordSegmentCreate   string         `json:"runOnRecordSegmentCreate"`
	RunOnRecordSegmentComplete string         `json:"runOnRecordSegmentComplete"`
}

func (pconf *Path) setDefaults() {
	// General
	pconf.Source = "publisher"
	pconf.SourceOnDemandStartTimeout = 10 * StringDuration(time.Second)
	pconf.SourceOnDemandCloseAfter = 10 * StringDuration(time.Second)

	// Record
	pconf.RecordPath = "./recordings/%path/%Y-%m-%d_%H-%M-%S-%f"
	pconf.RecordFormat = RecordFormatFMP4
	pconf.RecordPartDuration = StringDuration(1 * time.Second)
	pconf.RecordSegmentDuration = 3600 * StringDuration(time.Second)
	pconf.RecordDeleteAfter = 24 * 3600 * StringDuration(time.Second)

	// Publisher source
	pconf.OverridePublisher = true

	// Raspberry Pi Camera source
	pconf.RPICameraWidth = 1920
	pconf.RPICameraHeight = 1080
	pconf.RPICameraContrast = 1
	pconf.RPICameraSaturation = 1
	pconf.RPICameraSharpness = 1
	pconf.RPICameraExposure = "normal"
	pconf.RPICameraAWB = "auto"
	pconf.RPICameraAWBGains = []float64{0, 0}
	pconf.RPICameraDenoise = "off"
	pconf.RPICameraMetering = "centre"
	pconf.RPICameraFPS = 30
	pconf.RPICameraIDRPeriod = 60
	pconf.RPICameraBitrate = 1000000
	pconf.RPICameraProfile = "main"
	pconf.RPICameraLevel = "4.1"
	pconf.RPICameraAfMode = "continuous"
	pconf.RPICameraAfRange = "normal"
	pconf.RPICameraAfSpeed = "normal"
	pconf.RPICameraTextOverlay = "%Y-%m-%d %H:%M:%S - MediaMTX"

	// Hooks
	pconf.RunOnDemandStartTimeout = 10 * StringDuration(time.Second)
	pconf.RunOnDemandCloseAfter = 10 * StringDuration(time.Second)
}

func newPath(defaults *Path, partial *OptionalPath) *Path {
	pconf := &Path{}
	copyStructFields(pconf, defaults)
	copyStructFields(pconf, partial.Values)
	return pconf
}

// Clone clones the configuration.
func (pconf Path) Clone() *Path {
	enc, err := json.Marshal(pconf)
	if err != nil {
		panic(err)
	}

	var dest Path
	err = json.Unmarshal(enc, &dest)
	if err != nil {
		panic(err)
	}

	dest.Regexp = pconf.Regexp

	return &dest
}

func (pconf *Path) validate(
	conf *Conf,
	name string,
	deprecatedCredentialsMode bool,
) error {
	pconf.Name = name

	switch {
	case name == "all_others", name == "all":
		pconf.Regexp = regexp.MustCompile("^.*$")

	case name == "" || name[0] != '~': // normal path
		err := isValidPathName(name)
		if err != nil {
			return fmt.Errorf("invalid path name '%s': %w", name, err)
		}

	default: // regular expression-based path
		regexp, err := regexp.Compile(name[1:])
		if err != nil {
			return fmt.Errorf("invalid regular expression: %s", name[1:])
		}
		pconf.Regexp = regexp
	}

	// General

	if pconf.Source != "publisher" && pconf.Source != "redirect" &&
		pconf.Regexp != nil && !pconf.SourceOnDemand {
		return fmt.Errorf("a path with a regular expression (or path 'all') and a static source" +
			" must have 'sourceOnDemand' set to true")
	}
	switch {
	case pconf.Source == "publisher":

	case strings.HasPrefix(pconf.Source, "rtsp://") ||
		strings.HasPrefix(pconf.Source, "rtsps://"):
		_, err := base.ParseURL(pconf.Source)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}

	case strings.HasPrefix(pconf.Source, "rtmp://") ||
		strings.HasPrefix(pconf.Source, "rtmps://"):
		u, err := gourl.Parse(pconf.Source)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}

		if u.User != nil {
			pass, _ := u.User.Password()
			user := u.User.Username()
			if user != "" && pass == "" ||
				user == "" && pass != "" {
				return fmt.Errorf("username and password must be both provided")
			}
		}

	case strings.HasPrefix(pconf.Source, "http://") ||
		strings.HasPrefix(pconf.Source, "https://"):
		u, err := gourl.Parse(pconf.Source)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}

		if u.User != nil {
			pass, _ := u.User.Password()
			user := u.User.Username()
			if user != "" && pass == "" ||
				user == "" && pass != "" {
				return fmt.Errorf("username and password must be both provided")
			}
		}

	case strings.HasPrefix(pconf.Source, "udp://"):
		_, _, err := net.SplitHostPort(pconf.Source[len("udp://"):])
		if err != nil {
			return fmt.Errorf("'%s' is not a valid UDP URL", pconf.Source)
		}

	case strings.HasPrefix(pconf.Source, "srt://"):

		_, err := gourl.Parse(pconf.Source)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}

	case strings.HasPrefix(pconf.Source, "whep://") ||
		strings.HasPrefix(pconf.Source, "wheps://"):
		_, err := gourl.Parse(pconf.Source)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid URL", pconf.Source)
		}

	case pconf.Source == "redirect":

	case pconf.Source == "rpiCamera":

	default:
		return fmt.Errorf("invalid source: '%s'", pconf.Source)
	}
	if pconf.SourceOnDemand {
		if pconf.Source == "publisher" {
			return fmt.Errorf("'sourceOnDemand' is useless when source is 'publisher'")
		}
	}
	if pconf.SRTReadPassphrase != "" {
		err := srtCheckPassphrase(pconf.SRTReadPassphrase)
		if err != nil {
			return fmt.Errorf("invalid 'readRTPassphrase': %w", err)
		}
	}
	if pconf.Fallback != "" {
		if strings.HasPrefix(pconf.Fallback, "/") {
			err := isValidPathName(pconf.Fallback[1:])
			if err != nil {
				return fmt.Errorf("'%s': %w", pconf.Fallback, err)
			}
		} else {
			_, err := base.ParseURL(pconf.Fallback)
			if err != nil {
				return fmt.Errorf("'%s' is not a valid RTSP URL", pconf.Fallback)
			}
		}
	}

	// Record

	if conf.Playback {
		if !strings.Contains(pconf.RecordPath, "%Y") ||
			!strings.Contains(pconf.RecordPath, "%m") ||
			!strings.Contains(pconf.RecordPath, "%d") ||
			!strings.Contains(pconf.RecordPath, "%H") ||
			!strings.Contains(pconf.RecordPath, "%M") ||
			!strings.Contains(pconf.RecordPath, "%S") ||
			!strings.Contains(pconf.RecordPath, "%f") {
			return fmt.Errorf("record path '%s' is missing one of the mandatory elements"+
				" for the playback server to work: %%Y %%m %%d %%H %%M %%S %%f",
				pconf.RecordPath)
		}
	}

	// Authentication (deprecated)

	if deprecatedCredentialsMode {
		func() {
			var user Credential = "any"
			if pconf.PublishUser != nil && *pconf.PublishUser != "" {
				user = *pconf.PublishUser
			}

			var pass Credential
			if pconf.PublishPass != nil && *pconf.PublishPass != "" {
				pass = *pconf.PublishPass
			}

			ips := IPNetworks{mustParseCIDR("0.0.0.0/0")}
			if pconf.PublishIPs != nil && len(*pconf.PublishIPs) != 0 {
				ips = *pconf.PublishIPs
			}

			pathName := name
			if name == "all_others" || name == "all" {
				pathName = "~^.*$"
			}

			conf.AuthInternalUsers = append(conf.AuthInternalUsers, AuthInternalUser{
				User: user,
				Pass: pass,
				IPs:  ips,
				Permissions: []AuthInternalUserPermission{{
					Action: AuthActionPublish,
					Path:   pathName,
				}},
			})
		}()

		func() {
			var user Credential = "any"
			if pconf.ReadUser != nil && *pconf.ReadUser != "" {
				user = *pconf.ReadUser
			}

			var pass Credential
			if pconf.ReadPass != nil && *pconf.ReadPass != "" {
				pass = *pconf.ReadPass
			}

			ips := IPNetworks{mustParseCIDR("0.0.0.0/0")}
			if pconf.ReadIPs != nil && len(*pconf.ReadIPs) != 0 {
				ips = *pconf.ReadIPs
			}

			pathName := name
			if name == "all_others" || name == "all" {
				pathName = "~^.*$"
			}

			conf.AuthInternalUsers = append(conf.AuthInternalUsers, AuthInternalUser{
				User: user,
				Pass: pass,
				IPs:  ips,
				Permissions: []AuthInternalUserPermission{{
					Action: AuthActionRead,
					Path:   pathName,
				}},
			})
		}()
	}

	// Publisher source

	if pconf.DisablePublisherOverride != nil {
		pconf.OverridePublisher = !*pconf.DisablePublisherOverride
	}
	if pconf.SRTPublishPassphrase != "" {
		if pconf.Source != "publisher" {
			return fmt.Errorf("'srtPublishPassphase' can only be used when source is 'publisher'")
		}

		err := srtCheckPassphrase(pconf.SRTPublishPassphrase)
		if err != nil {
			return fmt.Errorf("invalid 'srtPublishPassphrase': %w", err)
		}
	}

	// RTSP source

	if pconf.SourceProtocol != nil {
		pconf.RTSPTransport = *pconf.SourceProtocol
	}
	if pconf.SourceAnyPortEnable != nil {
		pconf.RTSPAnyPort = *pconf.SourceAnyPortEnable
	}

	// Redirect source

	if pconf.Source == "redirect" {
		if pconf.SourceRedirect == "" {
			return fmt.Errorf("source redirect must be filled")
		}

		_, err := base.ParseURL(pconf.SourceRedirect)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid RTSP URL", pconf.SourceRedirect)
		}
	}

	// Raspberry Pi Camera source

	if pconf.Source == "rpiCamera" {
		for otherName, otherPath := range conf.Paths {
			if otherPath != pconf && otherPath != nil &&
				otherPath.Source == "rpiCamera" && otherPath.RPICameraCamID == pconf.RPICameraCamID {
				return fmt.Errorf("'rpiCamera' with same camera ID %d is used as source in two paths, '%s' and '%s'",
					pconf.RPICameraCamID, name, otherName)
			}
		}
	}
	switch pconf.RPICameraExposure {
	case "normal", "short", "long", "custom":
	default:
		return fmt.Errorf("invalid 'rpiCameraExposure' value")
	}
	switch pconf.RPICameraAWB {
	case "auto", "incandescent", "tungsten", "fluorescent", "indoor", "daylight", "cloudy", "custom":
	default:
		return fmt.Errorf("invalid 'rpiCameraAWB' value")
	}
	if len(pconf.RPICameraAWBGains) != 2 {
		return fmt.Errorf("invalid 'rpiCameraAWBGains' value")
	}
	switch pconf.RPICameraDenoise {
	case "off", "cdn_off", "cdn_fast", "cdn_hq":
	default:
		return fmt.Errorf("invalid 'rpiCameraDenoise' value")
	}
	switch pconf.RPICameraMetering {
	case "centre", "spot", "matrix", "custom":
	default:
		return fmt.Errorf("invalid 'rpiCameraMetering' value")
	}
	switch pconf.RPICameraAfMode {
	case "auto", "manual", "continuous":
	default:
		return fmt.Errorf("invalid 'rpiCameraAfMode' value")
	}
	switch pconf.RPICameraAfRange {
	case "normal", "macro", "full":
	default:
		return fmt.Errorf("invalid 'rpiCameraAfRange' value")
	}
	switch pconf.RPICameraAfSpeed {
	case "normal", "fast":
	default:
		return fmt.Errorf("invalid 'rpiCameraAfSpeed' value")
	}

	// Hooks

	if pconf.RunOnInit != "" && pconf.Regexp != nil {
		return fmt.Errorf("a path with a regular expression (or path 'all')" +
			" does not support option 'runOnInit'; use another path")
	}
	if (pconf.RunOnDemand != "" || pconf.RunOnUnDemand != "") && pconf.Source != "publisher" {
		return fmt.Errorf("'runOnDemand' and 'runOnUnDemand' can be used only when source is 'publisher'")
	}

	return nil
}

// Equal checks whether two Paths are equal.
func (pconf *Path) Equal(other *Path) bool {
	return reflect.DeepEqual(pconf, other)
}

// HasStaticSource checks whether the path has a static source.
func (pconf Path) HasStaticSource() bool {
	return strings.HasPrefix(pconf.Source, "rtsp://") ||
		strings.HasPrefix(pconf.Source, "rtsps://") ||
		strings.HasPrefix(pconf.Source, "rtmp://") ||
		strings.HasPrefix(pconf.Source, "rtmps://") ||
		strings.HasPrefix(pconf.Source, "http://") ||
		strings.HasPrefix(pconf.Source, "https://") ||
		strings.HasPrefix(pconf.Source, "udp://") ||
		strings.HasPrefix(pconf.Source, "srt://") ||
		strings.HasPrefix(pconf.Source, "whep://") ||
		strings.HasPrefix(pconf.Source, "wheps://") ||
		pconf.Source == "rpiCamera"
}

// HasOnDemandStaticSource checks whether the path has a on demand static source.
func (pconf Path) HasOnDemandStaticSource() bool {
	return pconf.HasStaticSource() && pconf.SourceOnDemand
}

// HasOnDemandPublisher checks whether the path has a on-demand publisher.
func (pconf Path) HasOnDemandPublisher() bool {
	return pconf.RunOnDemand != ""
}
