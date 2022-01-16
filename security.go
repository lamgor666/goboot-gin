package goboot

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
	GCorsSettings "github.com/lamgor666/goboot-common/CorsSettings"
	GJwtSettings "github.com/lamgor666/goboot-common/JwtSettings"
	"github.com/lamgor666/goboot-common/enum/JwtVerifyErrno"
	"github.com/lamgor666/goboot-common/util/castx"
	"github.com/lamgor666/goboot-common/util/fsx"
	"io/ioutil"
	"math"
	"os"
	"time"
)

var corsSettings *GCorsSettings.Settings
var jwtPublicKeyPemFile string
var jwtPrivateKeyPemFile string
var jwtSettings map[string]*GJwtSettings.Settings

func CorsSettings(settings ...interface{}) *GCorsSettings.Settings {
	if len(settings) > 0 {
		if settings[0] == nil {
			return nil
		}

		var _settings *GCorsSettings.Settings

		if st, ok := settings[0].(*GCorsSettings.Settings); ok {
			_settings = st
		} else if map1, ok := settings[0].(map[string]interface{}); ok && len(map1) > 0 {
			_settings = GCorsSettings.New(map1)
		}

		if _settings != nil {
			corsSettings = _settings
		}

		return nil
	}

	return corsSettings
}

func JwtPublicKeyPemFile(fpath ...string) string {
	if len(fpath) > 0 {
		if fpath[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(fpath[0])

		if stat, err := os.Stat(s1); err == nil && !stat.IsDir() {
			jwtPublicKeyPemFile = s1
		}

		return ""
	}

	return jwtPublicKeyPemFile
}

func JwtPrivateKeyPemFile(fpath ...string) string {
	if len(fpath) > 0 {
		if fpath[0] == "" {
			return ""
		}

		s1 := fsx.GetRealpath(fpath[0])

		if stat, err := os.Stat(s1); err == nil && !stat.IsDir() {
			jwtPrivateKeyPemFile = s1
		}

		return ""
	}

	return jwtPrivateKeyPemFile
}

func JwtSettings(key string, settings ...interface{}) *GJwtSettings.Settings {
	if len(settings) > 0 {
		if settings[0] == nil {
			return nil
		}

		var _settings *GJwtSettings.Settings

		if st, ok := settings[0].(*GJwtSettings.Settings); ok {
			_settings = st
		} else if map1, ok := settings[0].(map[string]interface{}); ok && len(map1) > 0 {
			map1["publicKeyPemFile"] = jwtPublicKeyPemFile
			map1["privateKeyPemFile"] = jwtPrivateKeyPemFile
			_settings = GJwtSettings.New(map1)
		}

		if _settings == nil {
			return nil
		}

		if len(jwtSettings) < 1 {
			jwtSettings = map[string]*GJwtSettings.Settings{key: _settings}
		} else {
			jwtSettings[key] = _settings
		}

		return nil
	}

	if len(jwtSettings) < 1 {
		return nil
	}

	return jwtSettings[key]
}

func ParseJwt(token string, pubpem ...string) (*jwt.Token, error) {
	var fpath string

	if len(pubpem) > 0 && pubpem[0] != "" {
		fpath = pubpem[0]
	}

	keyBytes := loadKeyPem("pub", fpath)

	if len(keyBytes) < 1 {
		return nil, errors.New("fail to load public key from pem file")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)

	if err != nil {
		return nil, err
	}

	return jwt.Parse(token, func(tk *jwt.Token) (interface{}, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tk.Header["alg"])
		}

		return publicKey, nil
	})
}

// @param *jwt.Token|string arg0
func VerifyJwt(arg0 interface{}, settings *GJwtSettings.Settings) int {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJwt(s1, settings.PublicKeyPemFile())
		token = tk
	}

	if token == nil || !token.Valid {
		return JwtVerifyErrno.Invalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return JwtVerifyErrno.Invalid
	}

	iss := settings.Issuer()

	if iss != "" && castx.ToString(claims["iss"]) != iss {
		return JwtVerifyErrno.Invalid
	}

	exp := castx.ToInt64(claims["exp"])

	if exp > 0 && time.Now().Unix() > exp {
		return JwtVerifyErrno.Expired
	}

	return 0
}

// @param *JwtSettings|string arg0
func BuildJwt(arg0 interface{}, isRefreshToken bool, claims ...map[string]interface{}) (token string, err error) {
	var settings *GJwtSettings.Settings

	if s1, ok := arg0.(*GJwtSettings.Settings); ok && s1 != nil {
		settings = s1
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		settings = JwtSettings(s1)
	}

	if settings == nil {
		err = errors.New("in goboot.BuildJsonWebToken function, *JwtSettings is nil")
		return
	}

	keyBytes := loadKeyPem("pri", settings.PrivateKeyPemFile())

	if len(keyBytes) < 1 {
		err = errors.New("in goboot.BuildJsonWebToken function, fail to load private key from pem file")
		return
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(keyBytes)

	if err != nil {
		return
	}

	var exp int64

	if isRefreshToken {
		exp = time.Now().Add(settings.RefreshTokenTtl()).Unix()
	} else {
		exp = time.Now().Add(settings.Ttl()).Unix()
	}

	mapClaims := jwt.MapClaims{
		"iss": settings.Issuer(),
		"exp": exp,
	}

	if len(claims) > 0 && len(claims[0]) > 0 {
		for claimName, claimValue := range claims[0] {
			mapClaims[claimName] = claimValue
		}
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, mapClaims).SignedString(privateKey)
	return
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaim(arg0 interface{}, name string, defaultValue ...interface{}) string {
	var dv string

	if len(defaultValue) > 0 {
		if s1, err := castx.ToStringE(defaultValue[0]); err == nil {
			dv = s1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if s1 := castx.ToString(claims[name]); s1 != "" {
		return s1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimBool(arg0 interface{}, name string, defaultValue ...interface{}) bool {
	var dv bool

	if len(defaultValue) > 0 {
		if b1, err := castx.ToBoolE(defaultValue[0]); err == nil {
			dv = b1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if b1, err := castx.ToBoolE(claims[name]); err == nil {
		return b1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimInt(arg0 interface{}, name string, defaultValue ...interface{}) int {
	dv := math.MinInt32

	if len(defaultValue) > 0 {
		if n1, err := castx.ToIntE(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToIntE(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimInt64(arg0 interface{}, name string, defaultValue ...interface{}) int64 {
	dv := int64(math.MinInt64)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToInt64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToInt64E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimFloat32(arg0 interface{}, name string, defaultValue ...interface{}) float32 {
	dv := float32(math.SmallestNonzeroFloat32)

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat32E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToFloat32E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimFloat64(arg0 interface{}, name string, defaultValue ...interface{}) float64 {
	dv := math.SmallestNonzeroFloat64

	if len(defaultValue) > 0 {
		if n1, err := castx.ToFloat64E(defaultValue[0]); err == nil {
			dv = n1
		}
	}

	token := getTokenInternal(arg0)

	if token == nil {
		return dv
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return dv
	}

	if n1, err := castx.ToFloat64E(claims[name]); err == nil {
		return n1
	}

	return dv
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimStringSlice(arg0 interface{}, name string) []string {
	token := getTokenInternal(arg0)

	if token == nil {
		return make([]string, 0)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return make([]string, 0)
	}

	return castx.ToStringSlice(claims[name])
}

// @param *jwt.Token|*gin.Context|string arg0
func JwtClaimIntSlice(arg0 interface{}, name string) []int {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*gin.Context); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJwt(s1)
		token = tk
	}

	if token == nil {
		return make([]int, 0)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return make([]int, 0)
	}

	return castx.ToIntSlice(claims[name])
}

func loadKeyPem(typ string, arg1 interface{}) []byte {
	var fpath string

	if s1, ok := arg1.(string); ok && s1 != "" {
		fpath = s1
	} else if s1, ok := arg1.(*GJwtSettings.Settings); ok && s1 != nil {
		switch typ {
		case "pub":
			fpath = s1.PublicKeyPemFile()
		case "pri":
			fpath = s1.PrivateKeyPemFile()
		}
	}

	if fpath == "" {
		switch typ {
		case "pub":
			fpath = jwtPublicKeyPemFile
		case "pri":
			fpath = jwtPrivateKeyPemFile
		}
	}

	if fpath == "" {
		return make([]byte, 0)
	}

	buf, err := ioutil.ReadFile(fpath)

	if err != nil {
		return make([]byte, 0)
	}

	return buf
}

func getTokenInternal(arg0 interface{}) *jwt.Token {
	var token *jwt.Token

	if tk, ok := arg0.(*jwt.Token); ok {
		token = tk
	} else if ctx, ok := arg0.(*gin.Context); ok {
		token = GetJwt(ctx)
	} else if s1, ok := arg0.(string); ok && s1 != "" {
		tk, _ := ParseJwt(s1)
		token = tk
	}

	return token
}
