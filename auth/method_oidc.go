package auth

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"text/template"
	"time"

	cst "thy/constants"
	"thy/paths"

	"github.com/mitchellh/cli"
	"github.com/pkg/browser"
)

const (
	callbackTimeout = 5 * time.Minute
)

type authResponse struct {
	state    string
	authCode string
	err      error
}

type redirectRequest struct {
	Provider    string `json:"provider"`
	CallbackUrl string `json:"callback_url"`
}

type redirectResponse struct {
	RedirectUrl string `json:"redirect_url"`
}

func (a *authenticator) buildOIDCParams(at AuthType, provider string, callback string) (*requestBody, error) {
	if callback == "" {
		callback = cst.DefaultCallback
	}

	ui := cli.BasicUi{
		Writer:      os.Stdout,
		Reader:      os.Stdin,
		ErrorWriter: os.Stderr,
	}

	uri := paths.CreateURI("oidc/auth", nil)

	var redirectResp redirectResponse
	redirectReq := &redirectRequest{
		Provider:    provider,
		CallbackUrl: fmt.Sprintf("http://%s/callback", callback),
	}
	reqErr := a.requestClient.DoRequestOut(http.MethodPost, uri, redirectReq, &redirectResp)
	if reqErr != nil {
		return nil, errors.New(reqErr.Error())
	}
	callbackListener, err := net.Listen("tcp", callback)
	if err != nil {
		return nil, fmt.Errorf("unable to open callback listener: %v", err)
	}
	defer callbackListener.Close()

	authChannel := make(chan authResponse)
	http.HandleFunc("/callback", handleOidcAuth(at, authChannel))

	go func() {
		err := http.Serve(callbackListener, nil)
		if err != nil && err != http.ErrServerClosed {
			authChannel <- authResponse{err: err}
		}
	}()

	if err = browser.OpenURL(redirectResp.RedirectUrl); err != nil {
		ui.Info(fmt.Sprintf("Unable to open browser, complete login process here:\n %s", redirectResp.RedirectUrl))
	}

	data := &requestBody{
		GrantType: authTypeToGrantType[at],
		Provider:  provider,
	}

	select {
	case ar := <-authChannel:
		if ar.err != nil {
			return nil, ar.err
		}
		data.State = ar.state
		data.AuthorizationCode = ar.authCode
		ui.Info(fmt.Sprintf("Received response from %s provider, submitting authorization code to %s", at, cst.ProductName))

	case <-time.After(callbackTimeout):
		ui.Info(fmt.Sprintf("Timeout occurred waiting for callback from %s provider", at))
		return nil, errors.New("no callback occurred after redirect")
	}

	return data, nil
}

// handleOidcAuth handles OIDC and Thycotic One auths.
func handleOidcAuth(at AuthType, doneCh chan<- authResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- authResponse{err: fmt.Errorf("reading body: %w", err)}
		}

		code := req.URL.Query().Get("code")
		state := req.URL.Query().Get("state")

		if code == "" {
			doneCh <- authResponse{
				err: errors.New("missing values in callback, authorization code is empty"),
			}
			w.Write(b)
			return
		}
		if state == "" {
			doneCh <- authResponse{
				err: errors.New("missing values in callback, authorization state is empty"),
			}
			w.Write(b)
			return
		}

		tmpl, err := template.New("youDidIt").Parse(youDidIt)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- authResponse{err: fmt.Errorf("template parsing: %w", err)}
		}

		vars := map[string]interface{}{
			"providerName": string(at),
		}

		err = tmpl.Execute(w, vars)
		if err != nil {
			w.Write([]byte(err.Error()))
			doneCh <- authResponse{err: fmt.Errorf("template execution: %w", err)}
		}

		doneCh <- authResponse{
			err:      nil,
			authCode: code,
			state:    state,
		}
	}
}

const youDidIt = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <title>{{.providerName}} Sign In Complete</title>

  <style>
 /*
* Skeleton V2.0.4
* Copyright 2014, Dave Gamache
* www.getskeleton.com
* Free to use under the MIT license.
* http://www.opensource.org/licenses/mit-license.php
* 12/29/2014
*/

/* Grid
–––––––––––––––––––––––––––––––––––––––––––––––––– */
.container {
  position: relative;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 0 20px;
  box-sizing: border-box; }
.column,
.columns {
  width: 100%;
  float: left;
  box-sizing: border-box; }

/* For devices larger than 550px */
@media (min-width: 550px) {
  .container {
    width: 80%; }
  .column,
  .columns {
    margin-left: 4%; }
  .column:first-child,
  .columns:first-child {
    margin-left: 0; }

  .one.column,
  .one.columns                    { width: 4.66666666667%; }
  .two.columns                    { width: 13.3333333333%; }
  .three.columns                  { width: 22%;            }
  .four.columns                   { width: 30.6666666667%; }
  .five.columns                   { width: 39.3333333333%; }
  .six.columns                    { width: 48%;            }
  .seven.columns                  { width: 56.6666666667%; }
  .eight.columns                  { width: 65.3333333333%; }
  .nine.columns                   { width: 74.0%;          }
  .ten.columns                    { width: 82.6666666667%; }
  .eleven.columns                 { width: 91.3333333333%; }
  .twelve.columns                 { width: 100%; margin-left: 0; }

  .one-third.column               { width: 30.6666666667%; }
  .two-thirds.column              { width: 65.3333333333%; }

  .one-half.column                { width: 48%; }

}


/* Base Styles
–––––––––––––––––––––––––––––––––––––––––––––––––– */
/* NOTE
html is set to 62.5% so that all the REM measurements throughout Skeleton
are based on 10px sizing. So basically 1.5rem = 15px :) */
html {
  font-size: 62.5%; }
body {
  font-size: 1.5em;
  line-height: 1.6;
  font-weight: 400;
  font-family: "Raleway", "HelveticaNeue", "Helvetica Neue", Helvetica, Arial, sans-serif;
  color: #222; }


/* Typography
–––––––––––––––––––––––––––––––––––––––––––––––––– */
h1, h2, h3, h4, h5, h6 {
  margin-top: 0;
  margin-bottom: 2rem;
  font-weight: 300; }
h1 { font-size: 4.0rem; line-height: 1.2;  letter-spacing: -.1rem;}
h2 { font-size: 3.6rem; line-height: 1.25; letter-spacing: -.1rem; }
h3 { font-size: 3.0rem; line-height: 1.3;  letter-spacing: -.1rem; }
h4 { font-size: 2.4rem; line-height: 1.35; letter-spacing: -.08rem; }
h5 { font-size: 1.8rem; line-height: 1.5;  letter-spacing: -.05rem; }
h6 { font-size: 1.5rem; line-height: 1.6;  letter-spacing: 0; }

p {
  margin-top: 0; }


/* Links
–––––––––––––––––––––––––––––––––––––––––––––––––– */
a {
  color: #1EAEDB; }
a:hover {
  color: #0FA0CE; }




/* Code
–––––––––––––––––––––––––––––––––––––––––––––––––– */
code {
  padding: .2rem .5rem;
  margin: 0 .2rem;
  font-size: 90%;
  white-space: nowrap;
  background: #F1F1F1;
  border: 1px solid #E1E1E1;
  border-radius: 4px; }
pre > code {
  display: block;
  padding: 1rem 1.5rem;
  white-space: pre; }



/* Spacing
–––––––––––––––––––––––––––––––––––––––––––––––––– */
button,
.button {
  margin-bottom: 1rem; }
input,
textarea,
select,
fieldset {
  margin-bottom: 1.5rem; }
pre,
blockquote,
dl,
figure,
table,
p,
ul,
ol,
form {
  margin-bottom: 2.5rem; }


/* Utilities
–––––––––––––––––––––––––––––––––––––––––––––––––– */
.u-full-width {
  width: 100%;
  box-sizing: border-box; }
.u-max-full-width {
  max-width: 100%;
  box-sizing: border-box; }
.u-pull-right {
  float: right; }
.u-pull-left {
  float: left; }


/* Misc
–––––––––––––––––––––––––––––––––––––––––––––––––– */
hr {
  margin-top: 3rem;
  margin-bottom: 3.5rem;
  border-width: 0;
  border-top: 1px solid #E1E1E1; }


/* Clearing
–––––––––––––––––––––––––––––––––––––––––––––––––– */

/* Self Clearing Goodness */
.container:after,
.row:after,
.u-cf {
  content: "";
  display: table;
  clear: both; }


  </style>

</head>

<body>

  <div class="container">
    <div class="row">
      <div class="two-thirds column" style="margin-top: 25%">
        <h2>{{.providerName}} Provider Sign In Complete</h2>
        <h5>Return to the CLI to verify sign in to DevOps Secrets Vault finished.</h5>
      </div>
    </div>
  </div>

</body>

</html>
`
