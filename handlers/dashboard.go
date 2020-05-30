package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/prometheus/common/log"
	"github.com/statping/statping/source"
	"github.com/statping/statping/types/errors"
	"github.com/statping/statping/types/users"
	"github.com/statping/statping/utils"
)

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	removeJwtToken(w)
	out := make(map[string]string)
	out["status"] = "success"
	returnJson(out, w, r)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	utils.LockLines.Lock()
	logs := make([]string, 0)
	length := len(utils.LastLines)
	// We need string log lines from end to start.
	for i := length - 1; i >= 0; i-- {
		logs = append(logs, utils.LastLines[i].FormatForHtml()+"\r\n")
	}
	utils.LockLines.Unlock()
	returnJson(logs, w, r)
}

type themeApi struct {
	Directory string `json:"directory,omitempty"`
	Base      string `json:"base"`
	Variables string `json:"variables"`
	Mobile    string `json:"mobile"`
}

func apiThemeViewHandler(w http.ResponseWriter, r *http.Request) {
	var base, variables, mobile, dir string
	assets := utils.Directory + "/assets"

	if _, err := os.Stat(assets); err == nil {
		dir = assets
	}

	if dir != "" {
		base, _ = utils.OpenFile(dir + "/scss/base.scss")
		variables, _ = utils.OpenFile(dir + "/scss/variables.scss")
		mobile, _ = utils.OpenFile(dir + "/scss/mobile.scss")
	} else {
		base, _ = source.TmplBox.String("scss/base.scss")
		variables, _ = source.TmplBox.String("scss/variables.scss")
		mobile, _ = source.TmplBox.String("scss/mobile.scss")
	}

	resp := &themeApi{
		Directory: dir,
		Base:      base,
		Variables: variables,
		Mobile:    mobile,
	}
	returnJson(resp, w, r)
}

func apiThemeSaveHandler(w http.ResponseWriter, r *http.Request) {
	var themes themeApi
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&themes)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := source.SaveAsset([]byte(themes.Base), "scss/base.scss"); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := source.SaveAsset([]byte(themes.Variables), "scss/variables.scss"); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := source.SaveAsset([]byte(themes.Mobile), "scss/mobile.scss"); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	if err := source.CompileSASS(source.DefaultScss...); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	resetRouter()
	sendJsonAction(themes, "saved", w, r)
}

func apiThemeCreateHandler(w http.ResponseWriter, r *http.Request) {
	dir := utils.Params.GetString("STATPING_DIR")
	if source.UsingAssets(dir) {
		err := errors.New("assets have already been created")
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}
	utils.Log.Infof("creating assets in folder: %s/%s", dir, "assets")
	if err := source.CreateAllAssets(dir); err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}
	if err := source.CompileSASS(source.DefaultScss...); err != nil {
		source.CopyToPublic(source.TmplBox, "css", "main.css")
		source.CopyToPublic(source.TmplBox, "css", "base.css")
		log.Errorln("Default 'base.css' was inserted because SASS did not work.")
	}
	resetRouter()
	sendJsonAction(dir+"/assets", "created", w, r)
}

func apiThemeRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if err := source.DeleteAllAssets(utils.Directory); err != nil {
		log.Errorln(fmt.Errorf("error deleting all assets %v", err))
	}
	sendJsonAction(utils.Directory+"/assets", "deleted", w, r)
}

func logsLineHandler(w http.ResponseWriter, r *http.Request) {
	if lastLine := utils.GetLastLine(); lastLine != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(lastLine.FormatForHtml()))
	}
}

type JwtClaim struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

func removeJwtToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieKey,
		Value:   "",
		Expires: time.Now(),
	})
}

func setJwtToken(user *users.User, w http.ResponseWriter) (JwtClaim, string) {
	expirationTime := time.Now().Add(72 * time.Hour)
	jwtClaim := JwtClaim{
		Username: user.Username,
		Admin:    user.Admin.Bool,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaim)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		log.Errorln("error setting token: ", err)
	}
	user.Token = tokenString
	// set cookies
	http.SetCookie(w, &http.Cookie{
		Name:    cookieKey,
		Value:   tokenString,
		Expires: expirationTime,
	})
	return jwtClaim, tokenString
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	form := parseForm(r)
	username := form.Get("username")
	password := form.Get("password")

	user, auth := users.AuthUser(username, password)
	if auth {
		log.Infoln(fmt.Sprintf("User %v logged in from IP %v", user.Username, r.RemoteAddr))
		claim, token := setJwtToken(user, w)
		resp := struct {
			Token   string `json:"token"`
			IsAdmin bool   `json:"admin"`
		}{
			token,
			claim.Admin,
		}
		returnJson(resp, w, r)
	} else {
		resp := struct {
			Error string `json:"error"`
		}{
			"incorrect authentication",
		}
		returnJson(resp, w, r)
	}
}
