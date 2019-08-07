package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	slackauth "github.com/phoenixcoder/slack-golang-sdk/auth"
	"github.com/phoenixcoder/slack-golang-sdk/slashcmd"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	timeout                 = 2
	forbiddenErrRespMsg     = "uh uh uh...you didn't say the magic word."
	internalErrRespMsg      = "Sorry...we uh...messed up."
	funcNotFoundErrMsgFmt   = "We're embarrassed for you, but we don't know a '%s'. Try these instead:\n'%s'"
	logMsg                  = "%s, %s"
	routerRegistryUrlEnvVar = "ROUTER_REGISTRY_URL_ENV_VAR"
	requestUrlRootEnvVar    = "REQUEST_URL_ROOT"
	cmdRequestUrlRoot       = "requestUrl"
	cmdFunctionsReg         = "functions"
	contentTypeHeader       = "content-type"
)

var (
	// TODO Convert this registry to a generic Data Access Object (DAO)
	//      that can pull from any data source.
	registry       commandRegistry
	registryUrl    = os.Getenv(routerRegistryUrlEnvVar)
	requestUrlRoot = os.Getenv(requestUrlRootEnvVar)
)

type httpClientInterface interface {
	Post(url string, contentType string, body io.Reader) (*http.Response, error)
}

func setInternalErrCode(resp *events.APIGatewayProxyResponse, reason string) {
	setErredStatusCode(resp, internalErrRespMsg, reason, http.StatusInternalServerError)
}

func setForbiddenErrCode(resp *events.APIGatewayProxyResponse) {
	setErredStatusCode(resp, forbiddenErrRespMsg, "You're just not allowed.", http.StatusForbidden)
}

func setErredStatusCode(resp *events.APIGatewayProxyResponse, msg string, reason string, statusCode int) {
	resp.StatusCode = statusCode
	resp.Body = msg + " (" + http.StatusText(resp.StatusCode) + ")"
	log.Printf(logMsg, resp.Body, reason)
}

// Handles things...duh
// 1. Authenticate the request.
// 2. Route the request to the appropriate function.
//    * Extract the function name from the request.
//    * Check whether the function name exists.
//    * Retrieves a function URL endpoint to send a request to.
//    * Create/send request to endpoint with arguments from this request.
// 3. Send immediate status OK reponse to caller unless authN failed.
// 4. Create/send a request to response_url once response returns from endpoint.
func routerHandler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	authNOk, err := slackauth.AuthenticateLambdaReq(&request)
	resp := &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}

	if err != nil {
		// TODO All errors with a response returning immediately must be marked as 200 with an error message.
		setInternalErrCode(resp, err.Error())
		return resp, nil
	}

	if !authNOk {
		setForbiddenErrCode(resp)
		return resp, nil
	}

	slashCmdInfo, err := slashcmd.Parse(request.Body)
	if err != nil {
		setInternalErrCode(resp, err.Error())
		return resp, nil
	}

	// TODO Check for help keyword.
	funcName := slashCmdInfo.Arguments[0]
	pReqUrlRoot, err := url.Parse(requestUrlRoot)
	if err != nil {
		return nil, err
	}

	// TODO Retrieve function record when help keyword is accessed.
	reqUrl := pReqUrlRoot
	reqUrl.Path = path.Join(pReqUrlRoot.Path, funcName)
	return routeRequest(reqUrl.String(), request.Body, &http.Client{Timeout: time.Second * timeout})
}

func routeRequest(requestUrl string, body string, client httpClientInterface) (*events.APIGatewayProxyResponse, error) {
	log.Printf("Route Request Url: %s\n", requestUrl)
	log.Printf("Route Request Body: %s\n", body)
	resp := &events.APIGatewayProxyResponse{
		Headers:    make(map[string]string),
		StatusCode: http.StatusOK,
		Body:       http.StatusText(http.StatusOK),
	}

	routedResp, err := client.Post(requestUrl, "", strings.NewReader(body))
	if err != nil {
		setInternalErrCode(resp, err.Error())
		return resp, err
	}

	resp.StatusCode = routedResp.StatusCode
	routedRespBody, err := ioutil.ReadAll(routedResp.Body)
	if err != nil {
		setInternalErrCode(resp, err.Error())
		return resp, err
	}
	resp.Body = string(routedRespBody)
	resp.Headers[contentTypeHeader] = routedResp.Header.Get(contentTypeHeader)
	return resp, err
}

func main() {
	// TODO Perform smart loading of contents for local testing. Flags > Local Variables > Environment Variable > Configuration File search.
	registry, err := NewCommandRegistryFromUrl(registryUrl)
	if err != nil {
		log.Fatalf("Failed Command Registry Loading: %+v\n", err)
	} else {
		log.Printf("Command Registry Loaded: %+v\n", registry)
		lambda.Start(routerHandler)
	}
}
