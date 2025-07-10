package utils

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	packagegeneralinterfaces "github.com/golang-etl/package-general/src/interfaces"
	"github.com/golang-etl/package-http/src/consts"
	"github.com/golang-etl/package-http/src/interfaces"
	"github.com/labstack/echo/v4"
)

func InternalServerErrorResponse(c echo.Context, shared *packagegeneralinterfaces.Shared, moduleName string, includeMessage bool, includeStack bool) {
	err := recover()

	if err == nil {
		return
	}

	body := interfaces.ResponseBodyError{}
	body.ErrorCode = "INTERNAL_SERVER_ERROR"
	body.Message = "Tenemos problemas para responder a su solicitud."

	if includeMessage {
		errorMessage := ErrorToMessage(err)
		body.Message = errorMessage
	}

	if includeStack {
		beautyStack := ParseAndSortStackTrace(string(debug.Stack()), moduleName)
		body.Stack = &beautyStack
	}

	AdaptEchoResponse(c, shared, interfaces.Response{
		StatusCode: http.StatusInternalServerError,
		Headers:    consts.HeaderContentType.JSON,
		Body:       body,
	})

	panic(err)
}

func ErrorToMessage(err any) string {
	switch v := err.(type) {
	case error:
		return v.Error()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ParseAndSortStackTrace(stack, moduleName string) []interfaces.ResponseBodyErrorStackItem {
	lines := strings.Split(stack, "\n")
	var results []interfaces.ResponseBodyErrorStackItem

	reFunc := regexp.MustCompile(`^\s*(.+)\((.*)\)$`)
	reFile := regexp.MustCompile(`^\s*(\/.*):\d+`)

	for i := 0; i < len(lines)-1; i++ {
		funcLine := strings.TrimSpace(lines[i])
		fileLine := strings.TrimSpace(lines[i+1])

		if reFunc.MatchString(funcLine) && reFile.MatchString(fileLine) {
			item := interfaces.ResponseBodyErrorStackItem{
				FuncName: funcLine,
				File:     fileLine,
			}
			results = append(results, item)
			i++
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		inModuleI := strings.Contains(results[i].FuncName, moduleName)
		inModuleJ := strings.Contains(results[j].FuncName, moduleName)

		if inModuleI && !inModuleJ {
			return true
		}
		if !inModuleI && inModuleJ {
			return false
		}
		return results[i].File < results[j].File
	})

	return results
}
