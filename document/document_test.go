package document

import (
	"bytes"
	"io"
	"lightsites/config"
	"lightsites/constants"
	"lightsites/helpers"

	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// TestProcessAttributes validates that the ProcessAttributes function
// correctly assigns attributes from an HTML document and throws an error
// if the mandatory title attribute is not set.
func TestProcessAttributes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	const fillerHTML = `<head></head><body></body>`

	tests := []struct {
		InputHTML        string
		InputDocument    Document
		ExpectedDocument Document
		ExpectedError    error
	}{
		{
			InputHTML: fmt.Sprintf(
				`%v<%v %v="%v"></%v>`,
				fillerHTML,
				constants.AttributeTag,
				constants.TitleAttribute,
				constants.TitleAttributeExample,
				constants.AttributeTag,
			),
			InputDocument: Document{
				Attributes: make(map[string]string),
			},
			ExpectedDocument: Document{
				Attributes: map[string]string{
					constants.TitleAttribute: constants.TitleAttributeExample,
				},
			},
			ExpectedError: nil,
		},
		{
			InputHTML: fillerHTML,
			InputDocument: Document{
				Attributes: make(map[string]string),
			},
			ExpectedDocument: Document{
				Attributes: make(map[string]string),
			},
			ExpectedError: fmt.Errorf(
				"document does not contain mandatory <%v %v=\"%v\"></%v> attribute",
				constants.AttributeTag,
				constants.TitleAttribute,
				constants.TitleAttributeExample,
				constants.AttributeTag,
			),
		},
		{
			InputHTML: fmt.Sprintf(
				`%v<%v %v="%v"></%v>`,
				fillerHTML,
				constants.AttributeTag,
				"testattr",
				"testval",
				constants.AttributeTag,
			),
			InputDocument: Document{
				Attributes: make(map[string]string),
			},
			ExpectedDocument: Document{
				Attributes: map[string]string{
					"testattr": "testval",
				},
			},
			ExpectedError: fmt.Errorf(
				"document does not contain mandatory <%v %v=\"%v\"></%v> attribute",
				constants.AttributeTag,
				constants.TitleAttribute,
				constants.TitleAttributeExample,
				constants.AttributeTag,
			),
		},
	}

	for _, test := range tests {
		doc, err := html.Parse(strings.NewReader(test.InputHTML))
		require.NoError(err)
		err = test.InputDocument.ProcessAttributes(doc)
		assert.Equal(test.ExpectedError, err)
		assert.Equal(test.ExpectedDocument, test.InputDocument)
	}
}

func TestProcessNode(t *testing.T) {
	assert := assert.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputNode     *html.Node
		ExpectOneAttr html.Attribute
	}{
		{
			"TestProcessNode happy style path",
			&defaultDocument,
			&html.Node{
				Type: html.ElementNode,
				Data: constants.ImgNode,
				Attr: []html.Attribute{
					{
						Key: constants.SrcAttribute,
						Val: "test.jpg",
					},
				},
			},
			html.Attribute{
				Key: constants.StyleAttribute,
				Val: defaultDocument.Config.Rules[constants.ImgNode][constants.StyleAttribute],
			},
		},
		{
			"TestProcessNode happy pre-existing style path",
			&defaultDocument,
			&html.Node{
				Type: html.ElementNode,
				Data: constants.ImgNode,
				Attr: []html.Attribute{
					{
						Key: constants.SrcAttribute,
						Val: "test.jpg",
					},
					{
						Key: constants.StyleAttribute,
						Val: "min-width: 50%;",
					},
				},
			},
			html.Attribute{
				Key: constants.StyleAttribute,
				Val: fmt.Sprintf("%v %v", "min-width: 50%;", defaultDocument.Config.Rules[constants.ImgNode][constants.StyleAttribute]),
			},
		},
		{
			"TestProcessNode happy table path (do not use this function for tables)",
			&defaultDocument,
			&html.Node{
				Type: html.ElementNode,
				Data: constants.TableNode,
				Attr: []html.Attribute{},
			},
			html.Attribute{
				Key: constants.ClassAttribute,
				Val: defaultDocument.Config.Rules[constants.TableNode][constants.ClassAttribute],
			},
		},
	}

	for _, test := range tests {
		test.InputDocument.ProcessNode(test.InputNode)
		testPass := false
		for _, a := range test.InputNode.Attr {
			if a.Key == test.ExpectOneAttr.Key && a.Val == test.ExpectOneAttr.Val {
				testPass = true
				break
			}
		}
		assert.True(testPass, test.TestName)
	}
}

func TestProcessTemplateNode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultConfig.Directories.Templates = "../tests/templates"
	defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
		ExpectError   bool
	}{
		{
			"ProcessTemplateNode happy path false heading",
			&defaultDocument,
			`<body><template file="alert.html" heading="false" alert-text="Heads up!"></template></body>`,
			`<html><head></head><body><div class="alert alert-primary">Heads up!</div></body></html>`,
			false,
		},
		{
			"ProcessTemplateNode happy path heading true",
			&defaultDocument,
			`<body><span>Hello</span><template file="alert.html" heading="true" alert-text="Heads up!"></template></body>`,
			`<html><head></head><body><div class="alert alert-primary">Heads up!</div><span>Hello</span></body></html>`,
			false,
		},
		{
			"ProcessTemplateNode happy path no heading",
			&defaultDocument,
			`<body><span>Hello</span><template file="alert.html" alert-text="Heads up!"></template></body>`,
			`<html><head></head><body><span>Hello</span><div class="alert alert-primary">Heads up!</div></body></html>`,
			false,
		},
		{
			"ProcessTemplateNode no file specified",
			&defaultDocument,
			`<body><span>Hello</span><template></template></body>`,
			`<html><head></head><body><span>Hello</span><template></template></body></html>`,
			true,
		},
		{
			"ProcessTemplateNode non-existent file specified",
			&defaultDocument,
			`<body><span>Hello</span><template file="does not exist"></template></body>`,
			`<html><head></head><body><span>Hello</span><template file="does not exist"></template></body></html>`,
			true,
		},
		{
			"ProcessTemplateNode bad html in file specified",
			&defaultDocument,
			`<body><span>Hello</span><template file="invalid.html"></template></body>`,
			`<html><head></head><body><span>Hello</span><template file="invalid.html"></template></body></html>`,
			true,
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)
		err = test.InputDocument.ProcessTemplateNode(
			helpers.GetNodeOfType(
				inputHTML[0],
				constants.TemplateNode,
			),
			inputHTML[0],
		)
		if !test.ExpectError {
			assert.NoError(err)
		} else {
			assert.Error(err)
		}

		var buf bytes.Buffer
		w := io.Writer(&buf)
		err = html.Render(w, inputHTML[0])
		assert.NoError(err, test.TestName)
		assert.Equal(test.OutputHTML, buf.String(), test.TestName)
	}
}

func TestProcessBodyNode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
	}{
		{
			"ProcessBodyNode scenario 1 - no children in body",
			&defaultDocument,
			`<body></body>`,
			fmt.Sprintf(
				`<html><head></head><body><div class="%v"><div class="%v"><div class="%v"></div></div></div></body></html>`,
				defaultDocument.Config.BodyConfig.ContainerClass,
				defaultDocument.Config.BodyConfig.RowClass,
				defaultDocument.Config.BodyConfig.ColClass,
			),
		},
		{
			"ProcessBodyNode scenario 1 - one child in body",
			&defaultDocument,
			`<body><p>Test</p></body>`,
			fmt.Sprintf(
				`<html><head></head><body><div class="%v"><div class="%v"><div class="%v"><p>Test</p></div></div></div></body></html>`,
				defaultDocument.Config.BodyConfig.ContainerClass,
				defaultDocument.Config.BodyConfig.RowClass,
				defaultDocument.Config.BodyConfig.ColClass,
			),
		},
		{
			"ProcessBodyNode scenario 3 - multiple children",
			&defaultDocument,
			`<body><p>Test!</p><p>Test 2!</p><p>Test 3!</p><p>Test 4!</p></body>`,
			fmt.Sprintf(
				`<html><head></head><body><div class="%v"><div class="%v"><div class="%v"><p>Test!</p><p>Test 2!</p><p>Test 3!</p><p>Test 4!</p></div></div></div></body></html>`,
				defaultDocument.Config.BodyConfig.ContainerClass,
				defaultDocument.Config.BodyConfig.RowClass,
				defaultDocument.Config.BodyConfig.ColClass,
			),
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)

		assert.NoError(test.InputDocument.ProcessBodyNode(helpers.GetNodeOfType(inputHTML[0], constants.BodyNode)))

		var buf bytes.Buffer
		w := io.Writer(&buf)
		err = html.Render(w, inputHTML[0])
		assert.NoError(err, test.TestName)
		assert.Equal(test.OutputHTML, buf.String(), test.TestName)
	}
}

func TestProcessHeadNode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	docScenario1 := Document{Config: &defaultConfig, Attributes: map[string]string{constants.TitleAttribute: "Test 1"}}
	docScenario2 := Document{Config: &defaultConfig, Attributes: map[string]string{constants.TitleAttribute: "Test 2"}}
	docScenario3 := Document{Config: &defaultConfig, Attributes: map[string]string{constants.TitleAttribute: "Test 3"}}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
	}{
		{
			"ProcessHeadNode scenario 1 - no children in body",
			&docScenario1,
			`<head></head>`,
			`<html><head><link href="/assets/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous"/><link href="/assets/custom.css" rel="stylesheet" crossorigin="anonymous"/><title>Test 1</title></head><body></body></html>`,
		},
		{
			"ProcessHeadNode scenario 2 - one child in body that gets removed since it's a title",
			&docScenario2,
			`<head><title>Test</title></head>`,
			`<html><head><link href="/assets/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous"/><link href="/assets/custom.css" rel="stylesheet" crossorigin="anonymous"/><title>Test 2</title></head><body></body></html>`,
		},
		{
			"ProcessHeadNode scenario 3 - multiple children, including extra unneeded title node",
			&docScenario3,
			`<head><title>Test</title><link href="/assets/test.css" rel="stylesheet"/></head>`,
			`<html><head><link href="/assets/test.css" rel="stylesheet"/><link href="/assets/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous"/><link href="/assets/custom.css" rel="stylesheet" crossorigin="anonymous"/><title>Test 3</title></head><body></body></html>`,
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)

		assert.NoError(test.InputDocument.ProcessHeadNode(helpers.GetNodeOfType(inputHTML[0], constants.HeadNode)))

		var buf bytes.Buffer
		w := io.Writer(&buf)
		err = html.Render(w, inputHTML[0])
		assert.NoError(err, test.TestName)
		assert.Equal(test.OutputHTML, buf.String(), test.TestName)
	}
}

func TestProcessDirectoryNode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultDocument := Document{Config: &defaultConfig}

	defaultDocument.DocumentDirectory = &[]string{
		"testDocument",
		".testDocument",
	}

	nilDocDirectory := config.GetDefaultConfig()
	nilDocument := Document{
		Config:            &nilDocDirectory,
		DocumentDirectory: nil,
	}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
		ExpectError   bool
	}{
		{
			"ProcessDirectoryNode - Happy path",
			&defaultDocument,
			"<directory></directory>",
			fmt.Sprintf(
				`<html><head></head><body><ul><li><a href="%v%v%v" rel="%v">%v</a></li></ul></body></html>`,
				defaultDocument.Config.Routing.RoutePrefix,
				(*defaultDocument.DocumentDirectory)[0],
				defaultConfig.Routing.UrlFileSuffix,
				constants.RelValue,
				(*defaultDocument.DocumentDirectory)[0],
			),
			false,
		},
		{
			"ProcessDirectoryNode - nil pointer for document.DocumentDirectory",
			&nilDocument,
			"<directory></directory>",
			``,
			true,
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)

		err = test.InputDocument.ProcessDirectoryNode(helpers.GetNodeOfType(inputHTML[0], constants.DirectoryNode))
		if test.ExpectError {
			assert.Error(err, test.TestName)
		} else {
			var buf bytes.Buffer
			w := io.Writer(&buf)
			err = html.Render(w, inputHTML[0])
			assert.NoError(err, test.TestName)
			assert.Equal(test.OutputHTML, buf.String(), test.TestName)
		}

	}
}

func TestProcessTableNode(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
		ExpectError   bool
	}{
		{
			"ProcessTableNode scenario 1 - table has no pre-defined classes",
			&defaultDocument,
			`<body><table><tr><th></th></tr><tr><td></td></tr></table></body>`,
			fmt.Sprintf(
				`<html><head></head><body><%v %v="%v"><table class="%v"><tbody><tr><th></th></tr><tr><td></td></tr></tbody></table></%v></body></html>`,
				constants.DivNode,
				constants.ClassAttribute,
				constants.DivTableResponsiveClass,
				constants.TableClasses,
				constants.DivNode,
			),
			false,
		},
		{
			"ProcessTableNode scenario 2 - table has pre-defined classes",
			&defaultDocument,
			`<body><table class="test-class"><tr><th></th></tr><tr><td></td></tr></table></body>`,
			fmt.Sprintf(
				`<html><head></head><body><%v %v="%v"><table class="test-class %v"><tbody><tr><th></th></tr><tr><td></td></tr></tbody></table></%v></body></html>`,
				constants.DivNode,
				constants.ClassAttribute,
				constants.DivTableResponsiveClass,
				constants.TableClasses,
				constants.DivNode,
			),
			false,
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)

		assert.NoError(test.InputDocument.ProcessTableNode(helpers.GetNodeOfType(inputHTML[0], constants.TableNode)))

		var buf bytes.Buffer
		w := io.Writer(&buf)
		err = html.Render(w, inputHTML[0])
		if test.ExpectError {
			assert.Error(err)
			continue
		}

		assert.NoError(err, test.TestName)
		assert.Equal(test.OutputHTML, buf.String(), test.TestName)
	}
}

// TestProcessTableNodeNoParent handles an edge case where a node doesn't
// have a parent. Generally the HTML Parse and ParseFragment functions will
// create a fully structured HTML document, even if a lone <table> node
// is provided.
func TestProcessTableNodeNoParent(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultDocument := Document{Config: &defaultConfig}

	inputHTMLStr := fmt.Sprintf(`<table class="%v"><tr><th></th></tr><tr><td></td></tr></table>`, defaultDocument.Config.Rules[constants.TableNode][constants.ClassAttribute])

	inputHTML, err := html.ParseFragment(strings.NewReader(inputHTMLStr), nil)

	require.NoError(err)
	require.Len(inputHTML, 1)

	tableNode := helpers.GetNodeOfType(inputHTML[0], constants.TableNode)
	tableNode.Parent.RemoveChild(tableNode)

	assert.Error(defaultDocument.ProcessTableNode(tableNode))
}

func TestProcessHTMLTree(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultConfig.Directories.Templates = "../tests/templates"
	defaultDocument := Document{
		Config:            &defaultConfig,
		Attributes:        make(map[string]string),
		DocumentDirectory: &[]string{"test1", "test2"},
	}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		OutputHTML    string
		ExpectError   bool
	}{
		{
			"ProcessHTMLTree scenario happy path",
			&defaultDocument,
			`<html><head></head><body><attributes title="Your Document Title"></attributes><template file="alert.html" heading="false" alert-text="Heads up!"></template><table><tr><th></th></tr><tr><td></td></tr></table><img src="test.jpg"/><directory></directory></body></html>`,
			`<html><head><link href="/assets/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous"/><link href="/assets/custom.css" rel="stylesheet" crossorigin="anonymous"/><title>Your Document Title</title></head><body><div class="container"><div class="row"><div class="col-lg-12"><div class="alert alert-primary">Heads up!</div><table><tbody><tr><th></th></tr><tr><td></td></tr></tbody></table><img src="test.jpg" style="max-width: 100%;"/><ul><li><a href="/content/test1.html" rel="noopener noreferrer">test1</a></li><li><a href="/content/test2.html" rel="noopener noreferrer">test2</a></li></ul></div></div></div></body></html>`,
			false,
		},
		{
			"ProcessHTMLTree scenario no attributes specified",
			&defaultDocument,
			`<html><head></head><body><attributes></attributes><template file="alert.html" heading="false" alert-text="Heads up!"></template><table><tr><th></th></tr><tr><td></td></tr></table><img src="test.jpg"/><directory></directory></body></html>`,
			``,
			true,
		},
		{
			"ProcessHTMLTree scenario invalid template specified",
			&defaultDocument,
			`<html><head></head><body><attributes title="Your Document Title"></attributes><template></template><table><tr><th></th></tr><tr><td></td></tr></table><img src="test.jpg"/><directory></directory></body></html>`,
			``,
			true,
		},
	}

	for _, test := range tests {
		inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		require.NoError(err)
		require.Len(inputHTML, 1)

		actualStr, err := test.InputDocument.ProcessHTMLTree(test.InputHTML)
		if test.ExpectError {
			assert.Error(err, test.TestName)
		} else {
			assert.NoError(err, test.TestName)
		}
		// log.Printf("%v\n\n%v", test.OutputHTML, actualStr)
		assert.Equal(test.OutputHTML, actualStr, test.TestName)
	}
}

func TestProcessNodesOfType(t *testing.T) {
	assert := assert.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultConfig.Directories.Templates = "../tests/templates"
	defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName      string
		InputDocument *Document
		InputHTML     string
		InputNodeType string
		OutputHTML    string
		ExpectError   bool
	}{
		{
			"ProcessNodesOfType happy path - template",
			&defaultDocument,
			`<body><template file="alert.html" heading="false" alert-text="Heads up!"></template><template file="alert.html" heading="false" alert-text="Heads up 2!"></template></body>`,
			constants.TemplateNode,
			`<html><head></head><body><div class="alert alert-primary">Heads up!</div><div class="alert alert-primary">Heads up 2!</div></body></html>`,
			false,
		},
		{
			"ProcessNodesOfType HTML template render failure",
			&defaultDocument,
			`<body><template></template><template file="alert.html" heading="false" alert-text="Heads up 2!"></template></body>`,
			constants.TemplateNode,
			``,
			true,
		},
	}

	for _, test := range tests {
		// inputHTML, err := html.ParseFragment(strings.NewReader(test.InputHTML), nil)
		// require.NoError(err)
		// require.Len(inputHTML, 1)
		actual, err := test.InputDocument.ProcessNodesOfType(
			test.InputHTML,
			constants.TemplateNode,
		)
		if !test.ExpectError {
			assert.NoError(err)
		} else {
			assert.Error(err)
		}
		assert.Equal(test.OutputHTML, actual)
	}
}

func TestParseDocument(t *testing.T) {
	assert := assert.New(t)

	defaultConfig := config.GetDefaultConfig()
	defaultConfig.Directories.Templates = "../tests/templates"
	defaultConfig.Directories.Documents = "../tests/test-docs"
	documentFiles, err := helpers.ReadDirectory(defaultConfig.Directories.Documents)
	if err != nil {
		fmt.Println(err)
	}
	documentDirectory := helpers.LoadFileDirectory(documentFiles)
	documents := []Document{}
	// defaultDocument := Document{Config: &defaultConfig}

	tests := []struct {
		TestName               string
		InputConf              *config.Config
		InputDocuments         *[]Document
		InputDocumentDirectory *[]string
		InputFileName          string
		OutputString           string
		ExpectError            bool
	}{
		{
			"ParseDocument happy path",
			&defaultConfig,
			&documents,
			&documentDirectory,
			"test1",
			// this next line is clunky because we remove the empty <title>
			// node from the tree, but useless whitespace remains
			`<!DOCTYPE html><html><head>` + "\n  " + `
  <meta name="GENERATOR" content="github.com/gomarkdown/markdown markdown processor for Go"/>
  <meta charset="utf-8"/>
<link href="/assets/bootstrap.min.css" rel="stylesheet" crossorigin="anonymous"/><link href="/assets/custom.css" rel="stylesheet" crossorigin="anonymous"/><title>Test Document</title></head>
<body><div class="container"><div class="row"><div class="col-lg-12">

<p></p>

<h1 id="test-document">Test Document</h1>

<p>This is a test document.</p>

<p></p><div class="alert alert-primary">Hi</div><p></p>

<p></p><ul><li><a href="/content/test2.html" rel="noopener noreferrer">test2</a></li><li><a href="/content/test1.html" rel="noopener noreferrer">test1</a></li></ul><p></p>



</div></div></div></body></html>`,
			false,
		},
		{
			"ParseDocument file does not exist",
			&defaultConfig,
			&documents,
			&documentDirectory,
			"does_not_exist.md",
			``,
			true,
		},
	}

	for _, test := range tests {
		actual, err := ParseDocument(
			test.InputConf,
			test.InputDocuments,
			test.InputDocumentDirectory,
			test.InputFileName,
		)
		if test.ExpectError {
			assert.Error(err, test.TestName)
		} else {
			assert.NoError(err, test.TestName)
		}

		assert.Equal(test.OutputString, actual, test.TestName)
	}
}
