package consts

// http
const (
	UrlKey         = "url"
	QueryKey       = "query"
	VariablesKey   = "variables"
	ContentTypeKey = "contentType"
	HeadersKey     = "headers"
	CookiesKey     = "cookies"
	BodyKey        = "body"
	UsernameKey    = "username"
	PasswordKey    = "password"
	TokenKey       = "token"
	BasicAuthKey   = "basic-auth"
	BearerAuthKey  = "bearer-token"
	ApiTokenKey    = "apikey-auth"
	ApiAddressKey  = "API Address"
	RequestUrlKey  = "REQUEST_URL"

	BasicAuthPrefix  = "Basic "
	BearerAuthPrefix = "Bearer "

	BasicAuthPassword = "PASSWORD"
	BasicAuthUsername = "USERNAME"

	DefaultTimeout = 30

	RequestUrlMissing    = "no request url provided"
	TestConnectionFailed = "Test connection failed - incorrect connection credentials. Please validate your connection details and try again"
)

// openapi
const (
	TypeArray    = "array"
	TypeInteger  = "integer"
	TypeNumber   = "number"
	TypeBoolean  = "boolean"
	TypeBool     = "bool"
	TypeObject   = "object"
	TypeString   = "string"
	TypeJson     = "code:json"
	TypeDropdown = "dropdown"

	ParamPrefix            = "{"
	ParamSuffix            = "}"
	BodyParamDelimiter     = "__"
	ArrayDelimiter         = ","
	ParamPlaceholderPrefix = "Example: "
)
