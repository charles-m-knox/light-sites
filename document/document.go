package document

import (
	"lightsites/config"
	"lightsites/constants"
	"lightsites/helpers"

	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"golang.org/x/net/html"
)

type Document struct {
	FileName          string
	FileContents      string
	DocumentName      string
	RenderedContents  string
	Title             string
	TitleURL          string
	ID                string
	DateCreated       time.Time
	DateModified      time.Time
	Attributes        map[string]string
	DocumentDirectory *[]string
	Config            *config.Config
}

// ProcessAttributes parses an input HTML node recursively for something
// like an `<attributes title="Document Title"></attributes>` tag. It will
// also assign the attribute values to the document.Attributes map.
//
// The value of htmlNode should be the output of:
//
// doc, err := html.Parse(strings.NewReader(htmlstr))
//
// where htmlstr is the full contents of an HTML document as a string.
func (document *Document) ProcessAttributes(htmlNode *html.Node) error {
	// take note of whether or not this htmlNode contains the title attributes
	// tags - if not, throw an error
	containsTitleAttribute := false

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case constants.AttributeTag:
				for _, attribute := range n.Attr {
					document.Attributes[attribute.Key] = attribute.Val
					if attribute.Key == constants.TitleAttribute && attribute.Val != "" {
						containsTitleAttribute = true
					}
				}
				n.Parent.RemoveChild(n)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(htmlNode)

	// error out if the mandatory tag/attributes are not defined
	if !containsTitleAttribute {
		return fmt.Errorf(
			"document does not contain mandatory <%v %v=\"%v\"></%v> attribute",
			constants.AttributeTag,
			constants.TitleAttribute,
			constants.TitleAttributeExample,
			constants.AttributeTag,
		)
	}

	return nil
}

// ProcessBodyNode sets up the <body> tag with a bootstrap-compatible grid,
// which is important because a Markdown->HTML document just contains
// <p> and <h1> tags (for example), but not anything layout-related
func (document *Document) ProcessBodyNode(n *html.Node) error {
	// create the nodes that make up the grid layout
	containerNode := &html.Node{
		Type: html.ElementNode,
		Data: constants.DivNode,
		Attr: []html.Attribute{{Key: constants.ClassAttribute, Val: document.Config.BodyConfig.ContainerClass}},
	}
	rowNode := &html.Node{
		Type: html.ElementNode,
		Data: constants.DivNode,
		Attr: []html.Attribute{{Key: constants.ClassAttribute, Val: document.Config.BodyConfig.RowClass}},
	}
	colNode := &html.Node{
		Type: html.ElementNode,
		Data: constants.DivNode,
		Attr: []html.Attribute{{Key: constants.ClassAttribute, Val: document.Config.BodyConfig.ColClass}},
	}

	rowNode.AppendChild(colNode)
	containerNode.AppendChild(rowNode)

	// there are 3 possible logic paths (scenarios):
	// 1. <body></body> - has no children
	// 2. <body><p>abc</p></body> - has one child
	// 3. <body><p>abc</p><p>def</p></body> - has >1 children
	//
	// In all three scenarios, the <body> element must contain the container
	// node as its first and only child.
	//
	// In scenario 1, this is simply a matter of appending the container node.
	// In scenario 2 and 3, both need to check if the first child pointer is nil.
	// In scenario 3, we also need to check if the first child has a sibling
	if n.FirstChild != nil {
		// scenario 3 - multiple children
		if n.FirstChild.NextSibling != nil {
			// takes all of the children of the <body> node and moves them into
			// the <div class="col-x"> node - these nodes are the <p>/<h1>/etc tags
			for bodyChild := n.FirstChild; bodyChild != nil; bodyChild = n.FirstChild.NextSibling {
				bodyChild.Parent.RemoveChild(bodyChild)
				colNode.AppendChild(bodyChild)
			}
		}
		// scenario 2 - one child
		// also covers the final child for scenario 3
		n.InsertBefore(containerNode, n.FirstChild)

		// now put the final child just after the first child node, since the above loop skipped it
		bodyLastChild := n.LastChild
		n.RemoveChild(bodyLastChild)
		if colNode.FirstChild != nil && colNode.FirstChild.NextSibling != nil {
			colNode.InsertBefore(bodyLastChild, colNode.FirstChild.NextSibling)
		} else {
			colNode.AppendChild(bodyLastChild)
		}
		return nil
	} else {
		// scenario 1 - easy
		n.AppendChild(containerNode)
	}
	return nil
}

// ProcessHeadNode sets up the <head> tag with bootstrap CSS and custom CSS
func (document *Document) ProcessHeadNode(n *html.Node) error {
	// in order for the bootstrap/custom CSS imports to work, the <head> tag
	// has to have each one as a child
	for _, cssImport := range document.Config.CSSImports {
		n.AppendChild(&html.Node{
			Type: html.ElementNode,
			Data: constants.LinkNode,
			Attr: []html.Attribute{
				{Key: constants.HrefAttribute, Val: fmt.Sprintf("%v%v", document.Config.Routing.AssetsPrefix, cssImport)},
				{Key: constants.RelAttribute, Val: constants.StylesheetVal},
				{Key: constants.CrossOriginAttribute, Val: constants.AnonymousVal},
			},
		})
	}
	return nil
}

// ProcessTableNode sets up the <table> tag to be responsive as well as some
// other common CSS. It gets wrapped in a <div> tag that uses the
// table-responsive bootstrap class (by default)
func (document *Document) ProcessTableNode(n *html.Node) error {
	if n.Parent == nil {
		return fmt.Errorf("cannot find parent for table node")
	}

	newNode := &html.Node{
		Type: html.ElementNode,
		Data: constants.DivNode,
		Attr: []html.Attribute{
			{
				Key: constants.ClassAttribute,
				Val: constants.DivTableResponsiveClass,
			},
		},
	}

	// get the parent node, add a <div> with table-responsive,
	// then move this node to that new node
	n.Parent.InsertBefore(newNode, n)
	n.Parent.RemoveChild(n)
	newNode.AppendChild(n)

	// apply table class rules from config
	document.ProcessNode(n)

	return nil
}

// ProcessDirectoryNode creates an HTML listing of all available documents
func (document *Document) ProcessDirectoryNode(n *html.Node) error {
	if document.DocumentDirectory == nil {
		return fmt.Errorf("document directory not initialized")
	}
	listNode := &html.Node{Type: html.ElementNode, Data: "ul"}
	for _, doc := range *document.DocumentDirectory {
		// don't show hidden files
		if strings.Index(doc, ".") == 0 {
			continue
		}
		newListItem := &html.Node{Type: html.ElementNode, Data: "li"}
		newListItemLink := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: constants.HrefAttribute, Val: fmt.Sprintf("%v%v%v", constants.ContentPrefixURL, doc, document.Config.Routing.UrlFileSuffix)},
				{Key: constants.RelAttribute, Val: constants.RelValue},
			},
		}
		newListItemLinkContent := &html.Node{
			Type: html.TextNode,
			Data: doc,
		}
		newListItemLink.AppendChild(newListItemLinkContent)
		newListItem.AppendChild(newListItemLink)
		listNode.AppendChild(newListItem)
	}
	n.Parent.InsertBefore(listNode, n)
	n.Parent.RemoveChild(n)

	return nil
}

// ProcessNode applies rules to nodes in a generalized manner, according
// to the configured rules
func (document *Document) ProcessNode(n *html.Node) {
	for key := range document.Config.Rules[n.Data] {
		newAttrVal := fmt.Sprintf("%v", document.Config.Rules[n.Data][key])

		for i, a := range n.Attr {
			if a.Key == key {
				n.Attr[i].Val = fmt.Sprintf("%v %v", a.Val, newAttrVal)
				return
			}
		}

		n.Attr = append(n.Attr, html.Attribute{
			Key: key,
			Val: newAttrVal,
		})
	}
}

// ProcessTemplateNode applies a template to a <template> node
func (document *Document) ProcessTemplateNode(n *html.Node, parentDoc *html.Node) error {
	templateAttributes := make(map[string]string)

	for _, attr := range n.Attr {
		templateAttributes[attr.Key] = attr.Val
	}

	_, ok := templateAttributes[constants.TemplateFileKey]
	if !ok {
		return fmt.Errorf("must specify template HTML attribute %v, none was specified", constants.TemplateFileKey)
	}

	content, err := ioutil.ReadFile(fmt.Sprintf("%v/%v", document.Config.Directories.Templates, templateAttributes[constants.TemplateFileKey]))
	if err != nil {
		return fmt.Errorf("failed to read template file %v: %v", templateAttributes[constants.TemplateFileKey], err.Error())
	}

	contentStr := string(content)
	// replace the values in the template with their rendered value
	for templateKey, templateVal := range templateAttributes {
		contentStr = strings.ReplaceAll(contentStr, fmt.Sprintf("{{%v}}", templateKey), templateVal)
	}

	// render the template contentStr as HTML
	templateHTMLNodes, err := html.Parse(strings.NewReader(contentStr))
	if err != nil {
		return fmt.Errorf("failed to parse template html: %v", err.Error())
	}

	// the HTML parse function returns something like
	// <html><head></head><body><rendered template></body></html>
	//
	// So, get the <body> node, so that we can extract only its child,
	// which is the rendered template
	renderedTemplateBodyNode := helpers.GetNodeOfType(templateHTMLNodes, constants.BodyNode)
	if renderedTemplateBodyNode == nil || renderedTemplateBodyNode.FirstChild == nil {
		return fmt.Errorf("unable to identify rendered template body or parent node for template %v", templateAttributes[constants.TemplateFileKey])
	}
	renderedTemplateNode := renderedTemplateBodyNode.FirstChild
	renderedTemplateNode.Parent.RemoveChild(renderedTemplateNode)

	// check if the "heading" variable is set to true
	_, ok = templateAttributes[constants.TemplateHeadingKey]
	if !ok {
		// The heading variable hasn't been set, so it's ok to render
		// the template here.
		// The FirstChild of the renderedTemplateBodyNode is the template
		// itself, so we insert it accordingly
		n.Parent.InsertBefore(renderedTemplateNode, n)
		n.Parent.RemoveChild(n)

		return nil
	}

	// if the template is a heading, it needs to be rendered first
	isHeading, err := strconv.ParseBool(templateAttributes[constants.TemplateHeadingKey])
	if err != nil || !isHeading {
		n.Parent.InsertBefore(renderedTemplateNode, n)
		n.Parent.RemoveChild(n)
		return nil
	}

	// the final logic path is where the heading value is set to true,
	// so push this template to the top of the page
	bodyNode := helpers.GetNodeOfType(parentDoc, constants.BodyNode)
	bodyNode.InsertBefore(renderedTemplateNode, bodyNode.FirstChild)
	n.Parent.RemoveChild(n)

	return nil
}

// ProcessNodesOfType is the decision tree for special-case HTML elements,
// such as <table>, <directory>, or <template>. These have special rules
// that require the entire HTML document to be passed in as a string,
// rendered as HTML nodes, and then each node of `nodeType` handled specially,
// before re-rendering as HTML.
// It has to be done this way, because the go-to method of recursively
// traversing the node tree from the topmost HTML node is incompatible with
// the special cases covered below - each of them manipulates one or more parent
// nodes, causing the traversal to exit early, leaving remaining nodes
// unprocessed.
func (document *Document) ProcessNodesOfType(htmlstr string, nodeType string) (output string, err error) {
	doc, err := html.Parse(strings.NewReader(htmlstr))
	if err != nil {
		return output, fmt.Errorf("failed to parse html: %v", err.Error())
	}

	for n := helpers.GetNodeOfType(doc, nodeType); n != nil; n = helpers.GetNodeOfType(doc, nodeType) {
		switch nodeType {
		case constants.TemplateNode:
			err = document.ProcessTemplateNode(n, doc)
			if err != nil {
				return output, fmt.Errorf("failed to process %v node: %v", constants.TemplateNode, err.Error())
			}
		case constants.DirectoryNode:
			err = document.ProcessDirectoryNode(n)
			if err != nil {
				return output, fmt.Errorf("failed to process %v node: %v", constants.DirectoryNode, err.Error())
			}
		case constants.TableNode:
			err = document.ProcessTableNode(n)
			if err != nil {
				return output, fmt.Errorf("failed to process %v node: %v", constants.TableNode, err.Error())
			}
		}
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)
	err = html.Render(w, doc)
	if err != nil {
		return output, fmt.Errorf("failed to render html: %v", err.Error())
	}
	return buf.String(), nil
}

// ProcessHTMLTree applies special handling of various nodes contained in an
// HTML document, which is contained within the `htmlstr` argument. The output
// is the rendered document to serve to users.
func (document *Document) ProcessHTMLTree(htmlstr string) (output string, err error) {
	doc, err := html.Parse(strings.NewReader(htmlstr))
	if err != nil {
		return output, fmt.Errorf("failed to parse html: %v", err.Error())
	}

	// extract document attributes before anything else
	err = document.ProcessAttributes(doc)
	if err != nil {
		return output, fmt.Errorf("failed to process HTML tree: %v", err.Error())
	}

	// render to HTML doc to a string
	wipHTML, err := helpers.RenderNode(doc)
	if err != nil {
		return output, fmt.Errorf("failed to render node for processing: %v", err.Error())
	}

	// handle special-case nodes
	templatedHTML, err := document.ProcessNodesOfType(wipHTML, constants.TemplateNode)
	if err != nil {
		return output, fmt.Errorf("failed to process %v nodes: %v", constants.TemplateNode, err.Error())
	}

	directoryProcessedHTML, err := document.ProcessNodesOfType(templatedHTML, constants.DirectoryNode)
	if err != nil {
		return output, fmt.Errorf("failed to process %v nodes: %v", constants.DirectoryNode, err.Error())
	}

	tableProcessedHTML, err := document.ProcessNodesOfType(directoryProcessedHTML, constants.DirectoryNode)
	if err != nil {
		return output, fmt.Errorf("failed to process %v nodes: %v", constants.TableNode, err.Error())
	}

	// render as HTML for further processing
	doc, err = html.Parse(strings.NewReader(tableProcessedHTML))
	if err != nil {
		return output, fmt.Errorf("failed to parse html: %v", err.Error())
	}

	// general manipulation
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// log.Printf("handling %v node", n.Data)
			switch n.Data {
			case constants.BodyNode:
				err := document.ProcessBodyNode(n)
				if err != nil {
					log.Printf("failed to process %v node: %v", constants.BodyNode, err.Error())
				}
			case constants.HeadNode:
				err := document.ProcessHeadNode(n)
				if err != nil {
					log.Printf("failed to process %v node: %v", constants.HeadNode, err.Error())
				}
			default:
				// only the table node likely has rules by default, and to avoid
				// double-processing table nodes, explicitly ignore them here
				if n.Data != constants.TableNode {
					document.ProcessNode(n)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// render the doc
	if doc == nil {
		return "", fmt.Errorf("doc is nil")
	}

	var buf bytes.Buffer
	w := io.Writer(&buf)
	err = html.Render(w, doc)
	if err != nil {
		return output, fmt.Errorf("failed to render html: %v", err.Error())
	}
	return buf.String(), nil
}

func GetMarkdownExtensionsConfig() parser.Extensions {
	return parser.NoIntraEmphasis | parser.Tables | parser.FencedCode | parser.Autolink | parser.Strikethrough | parser.SpaceHeadings | parser.Footnotes | parser.HeadingIDs | parser.Titleblock | parser.AutoHeadingIDs | parser.BackslashLineBreak | parser.DefinitionLists | parser.MathJax | parser.OrderedListStart | parser.SuperSubscript | parser.Footnotes | parser.HeadingIDs
}

func GetMarkdownHTMLFlags() mdhtml.Flags {
	return mdhtml.CommonFlags | mdhtml.CompletePage | mdhtml.NoopenerLinks | mdhtml.NoreferrerLinks | mdhtml.HrefTargetBlank | mdhtml.FootnoteReturnLinks | mdhtml.Smartypants | mdhtml.SmartypantsFractions | mdhtml.SmartypantsDashes | mdhtml.SmartypantsLatexDashes /* | mdhtml.TOC */
}

func ParseDocument(conf *config.Config, documents *[]Document, documentDirectory *[]string, fileName string) (finalMarkdown string, err error) {
	newDoc := Document{
		FileName: fileName,
		// FileContents: finalMarkdown,
		// DocumentName: documentName,
		ID:                fileName,
		Attributes:        make(map[string]string),
		DocumentDirectory: documentDirectory,
		Config:            conf,
	}
	*documents = append(*documents, newDoc)
	newDocIndex := len(*documents) - 1

	(*documents)[newDocIndex].FileContents = finalMarkdown

	// read the file
	content, err := ioutil.ReadFile(fmt.Sprintf("%v/%v%v", conf.Directories.Documents, fileName, constants.MarkdownFileSuffix))
	// content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read file %v: %v", fileName, err.Error())
	}
	// trim leading whitespace from the file
	content = []byte(strings.TrimLeft(string(content), "\n"))
	(*documents)[newDocIndex].DocumentName = strings.TrimRight(fileName, ".md")

	// configure the markdown parser and renderer
	MDParser := parser.NewWithExtensions(GetMarkdownExtensionsConfig())

	opts := mdhtml.RendererOptions{
		HeadingIDPrefix: "",
		Flags:           GetMarkdownHTMLFlags(),
	}
	MDRenderer := mdhtml.NewRenderer(opts)

	renderedMarkdown := markdown.ToHTML(content, MDParser, MDRenderer)
	finalMarkdown, err = newDoc.ProcessHTMLTree(string(renderedMarkdown))
	if err != nil {
		return "", fmt.Errorf("failed to process HTML tree for file %v: %v", fileName, err.Error())
	}

	(*documents)[newDocIndex].FileContents = finalMarkdown

	return finalMarkdown, nil
}
