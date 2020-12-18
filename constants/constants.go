package constants

const (
	ConfigFile = "config.yml"

	// AllowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"
	AssetsPrefixURL       = "/assets/" // don't forget the trailing slash
	AttributeTag          = "attributes"
	ContentPrefixURL      = "/content/" // don't forget the trailing slash
	DistDirectory         = RootDataDirectory + "/content"
	AssetsDirectory       = RootDataDirectory + "/assets"
	TemplatesDirectory    = RootDataDirectory + "/templates"
	RootDataDirectory     = "./src"
	TemplateFileKey       = "file"
	TemplateHeadingKey    = "heading"
	TitleAttribute        = "title"
	TitleAttributeExample = "Your Document Title"

	URLFileSuffix      = ".html"
	MarkdownFileSuffix = ".md"

	// special HTML nodes
	BodyNode      = "body"
	DirectoryNode = "directory"
	HeadNode      = "head"
	ImgNode       = "img"
	TableNode     = "table"
	TemplateNode  = "template"
	TitleNode     = "title"
	LinkNode      = "link"
	DivNode       = "div"

	// commonly used HTML attributes
	StyleAttribute       = "style"
	ClassAttribute       = "class"
	SrcAttribute         = "src"
	HrefAttribute        = "href"
	RelAttribute         = "rel"
	RelValue             = "noopener noreferrer"
	StylesheetVal        = "stylesheet"
	CrossOriginAttribute = "crossorigin"
	AnonymousVal         = "anonymous"

	// https://getbootstrap.com/docs/4.0/content/tables/
	// want classes: "table table-bordered table-striped table-hover table-sm"
	// eventually, want to wrap in a <div class="table-responsive"></div> element
	// https://godoc.org/golang.org/x/net/html#ex-Parse
	TableClasses            = "table table-bordered table-striped table-hover table-sm"
	DivTableResponsiveClass = "table-responsive"
	ImgStyles               = "max-width: 100%;"
	ContainerClass          = "container"
	RowClass                = "row"
	ColClass                = "col-lg-12"
)
