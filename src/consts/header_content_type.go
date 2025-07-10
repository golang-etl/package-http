package consts

const ContentTypeKey = "Content-Type"

var HeaderContentType = struct {
	JSON  map[string]string
	HTML  map[string]string
	EXCEL map[string]string
	PDF   map[string]string
}{
	JSON: map[string]string{
		ContentTypeKey: "application/json",
	},
	HTML: map[string]string{
		ContentTypeKey: "text/html",
	},
	EXCEL: map[string]string{
		ContentTypeKey: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	},
	PDF: map[string]string{
		ContentTypeKey: "application/pdf",
	},
}
