package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	packagegeneralinterfaces "github.com/golang-etl/package-general/src/interfaces"
	packagegeneralutils "github.com/golang-etl/package-general/src/utils"
	"github.com/golang-etl/package-http/src/consts"
	"github.com/golang-etl/package-http/src/interfaces"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

func GetValueFromRequest(c echo.Context, key string) (string, *interfaces.Response) {
	var value string

	if queryValue := c.QueryParam(key); queryValue != "" {
		value = queryValue

		return value, nil
	}

	if headerValue := c.Request().Header.Get(key); headerValue != "" {
		value = headerValue

		return value, nil
	}

	var bodyMap map[string]interface{}
	bodyBytes, err := io.ReadAll(c.Request().Body)

	if err != nil {
		return "", &interfaces.Response{
			StatusCode: 400,
			Headers:    consts.HeaderContentType.JSON,
			Body: interfaces.ResponseBodyError{
				Message:   "Error al obtener el valor JSON de la petición.",
				ErrorCode: "INVALID_REQUEST",
			},
		}
	}

	if len(bodyBytes) == 0 {
		return "", nil
	}

	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		return "", &interfaces.Response{
			StatusCode: 400,
			Headers:    consts.HeaderContentType.JSON,
			Body: interfaces.ResponseBodyError{
				Message:   "Error al obtener el valor JSON de la petición.",
				ErrorCode: "INVALID_REQUEST",
			},
		}
	}

	if bodyValue, ok := bodyMap[key]; ok {
		if bodyValueStr, ok := bodyValue.(string); ok {
			value = bodyValueStr
		}
	}

	return value, nil
}

func AdaptEchoResponse(c echo.Context, shared *packagegeneralinterfaces.Shared, res interfaces.Response) error {
	for k, v := range res.Headers {
		c.Response().Header().Set(k, v)
	}

	if shared != nil && shared.TraceToken != nil {
		c.Response().Header().Set("X-Trace-Token", *shared.TraceToken)

		jsonBytes, err := json.Marshal(res.Body)

		if err == nil {
			var bodyMap map[string]interface{}

			if err := json.Unmarshal(jsonBytes, &bodyMap); err == nil {
				bodyMap["x-trace-token"] = *shared.TraceToken
				res.Body = bodyMap
			}
		}
	}

	if res.IsFile {
		return c.Attachment(res.FilePath, res.FileName)
	}

	switch b := res.Body.(type) {
	case string:
		return c.String(res.StatusCode, b)
	case []byte:
		return c.Blob(res.StatusCode, res.Headers["Content-Type"], b)
	default:
		return c.JSON(res.StatusCode, res.Body)
	}
}

func ProxyRequest(c echo.Context, targetURL string, runtimeEnvironment string) error {
	targetURL = ReplacePathParams(targetURL, c)
	parsedURL, err := url.Parse(targetURL)

	if err != nil {
		return err
	}

	query := parsedURL.Query()

	for key, values := range c.QueryParams() {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest(c.Request().Method, parsedURL.String(), c.Request().Body)

	if err != nil {
		return err
	}

	for key, values := range c.Request().Header {
		req.Header[key] = values
	}

	for _, cookie := range c.Cookies() {
		req.AddCookie(cookie)
	}

	client := GetHttpClientByRuntimeEnvironment(runtimeEnvironment, parsedURL)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	c.Response().WriteHeader(resp.StatusCode)
	io.Copy(c.Response().Writer, resp.Body)

	return nil
}

func ReplacePathParams(urlTemplate string, c echo.Context) string {
	segments := strings.Split(urlTemplate, "/")

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			paramName := segment[1:]
			segments[i] = c.Param(paramName)
		}
	}

	return strings.Join(segments, "/")
}

func GetHttpClientByRuntimeEnvironment(runtimeEnvironment string, parsedURL *url.URL) *http.Client {
	var client *http.Client
	var err error

	if runtimeEnvironment == string(packagegeneralutils.RuntimeEnvironmentGCPCloudRun) ||
		runtimeEnvironment == string(packagegeneralutils.RuntimeEnvironmentGCPAppEngine) {
		ctx := context.Background()
		client, err = idtoken.NewClient(ctx, parsedURL.String())

		if err != nil {
			panic(fmt.Errorf("falló autenticación con ID Token: %v", err))
		}
	} else {
		client = &http.Client{}
	}

	return client
}

func ValidationErrorHandlerToUnprocessableEntityResponse(err error, obj interface{}, messages map[string]string) interfaces.Response {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		var errorsList []interfaces.ValidationError

		for _, e := range ve {
			fullNamespace := e.Namespace()
			partsNamespace := strings.Split(fullNamespace, ".")
			structName := partsNamespace[0]
			path := fullNamespace[len(structName)+1:]
			rule := e.Tag()
			message := messages[path+"."+rule]

			errorsList = append(errorsList, interfaces.ValidationError{
				Property: e.Field(),
				Path:     path,
				Rule:     e.Tag(),
				Message:  message,
			})
		}

		return interfaces.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body: map[string]interface{}{
				"errors": errorsList,
			},
		}
	}

	panic(err)
}
